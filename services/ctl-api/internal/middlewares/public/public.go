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
	{"POST", "/v1/orgs"}: {},
	{"GET", "/v1/orgs"}:  {},
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
		m.l.Info("public middleware")
		method := ctx.Request.Method
		path := ctx.Request.URL.RawPath
		key := [2]string{
			method,
			path,
		}
		_, found := publicEndpointList[key]
		if found {
			m.l.Info("marking request as public", zap.String("endpoint", fmt.Sprintf("%s:%s", method, path)))
		}

		ctx.Set(isPublicKey, found)
	}
}

func New(l *zap.Logger) *middleware {
	return &middleware{
		l: l,
	}
}
