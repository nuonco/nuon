package require

import (
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type Router struct {
	router gin.IRouter
	api    permissions.API
}

func Routes(router gin.IRouter, api permissions.API) Router {
	return Router{router: router, api: api}
}

func (r Router) Group(path string, handlers ...gin.HandlerFunc) Router {
	return Routes(r.router.Group(path, handlers...), r.api)
}

func (r Router) For(rule Rule, additional ...Rule) *gin.RouterGroup {
	route := operationGroup(r.router, r.api, rule.handler(r.api))
	for _, next := range additional {
		route.Use(next.handler(r.api))
	}
	return route
}

func (r Router) GET(path string, rule Rule, handlers ...gin.HandlerFunc) {
	r.For(rule).GET(path, handlers...)
}

func (r Router) POST(path string, rule Rule, handlers ...gin.HandlerFunc) {
	r.For(rule).POST(path, handlers...)
}

func (r Router) PUT(path string, rule Rule, handlers ...gin.HandlerFunc) {
	r.For(rule).PUT(path, handlers...)
}

func (r Router) PATCH(path string, rule Rule, handlers ...gin.HandlerFunc) {
	r.For(rule).PATCH(path, handlers...)
}

func (r Router) DELETE(path string, rule Rule, handlers ...gin.HandlerFunc) {
	r.For(rule).DELETE(path, handlers...)
}
