package public

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

var publicEndpointList map[[2]string]struct{} = map[[2]string]struct{}{
	{"GET", "/livez"}:                    {},
	{"GET", "/version"}:                  {},
	{"GET", "/readyz"}:                   {},
	{"OPTIONS", "*"}:                     {},
	{"GET", "/docs/*any"}:                {},
	{"GET", "/oapi/v2"}:                  {},
	{"GET", "/oapi/v3"}:                  {},
	{"GET", "/v1/general/config-schema"}: {},

	{"*", "/httpbin/*any"}: {},

	// cli / ui methods
	{"GET", "/v1/general/cli-config"}:                              {},
	{"GET", "/v1/general/cloud-platform/:cloud_platform/regions"}:  {},
	{"POST", "/v1/vcs/connection-callback"}:                        {},
	{"POST", "/v1/installs/:install_id/phone-home/:phone_home_id"}: {},
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
			cctx.SetPublicContext(ctx, true)
			return
		}

		wildcardKey := [2]string{
			method,
			"*",
		}
		_, found = publicEndpointList[wildcardKey]
		if found {
			m.l.Debug("marking request as public due to wildcard", zap.String("endpoint", fmt.Sprintf("%s:%s", method, path)))
			cctx.SetPublicContext(ctx, true)
			return
		}

		wildcardKey = [2]string{
			"*",
			path,
		}
		_, found = publicEndpointList[wildcardKey]
		if found {
			m.l.Debug("marking request as public due to wildcard", zap.String("endpoint", fmt.Sprintf("%s:%s", method, path)))
			cctx.SetPublicContext(ctx, true)
			return
		}

		cctx.SetPublicContext(ctx, false)
	}
}

func New(l *zap.Logger) *middleware {
	return &middleware{
		l: l,
	}
}
