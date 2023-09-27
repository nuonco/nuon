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
	{"POST", "/v1/orgs"}:                          {},
	{"GET", "/v1/orgs"}:                           {},
	{"POST", "/v1/general/metrics"}:               {},
	{"GET", "/v1/general/current-user"}:           {},
	{"GET", "/v1/sandboxes"}:                      {},
	{"GET", "/v1/sandboxes/:sandbox_id"}:          {},
	{"GET", "/v1/sandboxes/:sandbox_id/releases"}: {},
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
		method := ctx.Request.Method
		// full path will return the _matched_ path, such as `/v1/sandboxes/:id`
		path := ctx.FullPath()

		key := [2]string{
			method,
			path,
		}
		_, found := globalEndpointList[key]
		if found {
			m.l.Debug("marking request as global", zap.String("endpoint", fmt.Sprintf("%s:%s", method, path)))
		}

		ctx.Set(isGlobalKey, found)
	}
}

func New(l *zap.Logger) *middleware {
	return &middleware{
		l: l,
	}
}
