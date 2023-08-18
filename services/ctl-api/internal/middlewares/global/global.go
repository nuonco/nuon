package global

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	isGlobalKey string = "is_global"
)

var globalEndpointList map[[2]string]struct{} = map[[2]string]struct{}{
	{"POST", "/v1/orgs"}:                {},
	{"GET", "/v1/orgs"}:                 {},
	{"POST", "/v1/general/metrics"}:     {},
	{"GET", "/v1/general/current-user"}: {},
}

func IsGlobal(ctx *gin.Context) bool {
	isGlobal, exists := ctx.Get(isGlobalKey)
	if !exists {
		return false
	}

	return isGlobal.(bool)
}

type middleware struct {
	l *zap.Logger
}

func (m middleware) Name() string {
	return "global"
}

func (m middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		m.l.Info("public middleware")
		method := ctx.Request.Method
		path := ctx.Request.URL.Path

		m.l.Info("request", zap.String("path", path), zap.String("method", method))
		key := [2]string{
			method,
			path,
		}
		_, found := globalEndpointList[key]
		if found {
			m.l.Info("marking request as public", zap.String("endpoint", fmt.Sprintf("%s:%s", method, path)))
		}

		ctx.Set(isGlobalKey, found)
	}
}

func New(l *zap.Logger) *middleware {
	return &middleware{
		l: l,
	}
}
