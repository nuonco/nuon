package public

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	isPublicKey string = "is_public"
)

var publicEndpointList map[[2]string]struct{} = map[[2]string]struct{}{
	{"GET", "/livez"}:         {},
	{"GET", "/version"}:       {},
	{"GET", "/readyz"}:        {},
	{"OPTIONS", "*"}:          {},
	{"GET", "/docs/*any"}:     {},
	{"GET", "/oapi3/swagger"}: {},

	// cli / ui methods
	{"GET", "/v1/general/cli-config"}:       {},
	{"POST", "/v1/vcs/connection-callback"}: {},

	// installer methods
	{"GET", "/v1/installer/:installer_slug/render"}:              {},
	{"GET", "/v1/installer/:installer_slug/install/:install_id"}: {},
	{"POST", "/v1/installer/:installer_slug/installs"}:           {},
}

func IsPublic(ctx *gin.Context) bool {
	isPublic, exists := ctx.Get(isPublicKey)
	if !exists {
		return false
	}

	return isPublic.(bool)
}

type middleware struct {
	l *zap.Logger
}

func (m middleware) Name() string {
	return "public"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		// full path will return the _matched_ path, such as `/v1/sandboxes/:id`
		path := ctx.FullPath()
		key := [2]string{
			method,
			path,
		}
		_, found := publicEndpointList[key]
		if found {
			m.l.Debug("marking request as public", zap.String("endpoint", fmt.Sprintf("%s:%s", method, path)))
			ctx.Set(isPublicKey, true)
			return
		}

		wildcardKey := [2]string{
			method,
			"*",
		}
		_, found = publicEndpointList[wildcardKey]
		if found {
			m.l.Debug("marking request as public due to wildcard", zap.String("endpoint", fmt.Sprintf("%s:%s", method, path)))
			ctx.Set(isPublicKey, true)
			return
		}

		ctx.Set(isPublicKey, false)
	}
}

func New(l *zap.Logger) *middleware {
	return &middleware{
		l: l,
	}
}
