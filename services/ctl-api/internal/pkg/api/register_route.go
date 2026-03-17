package api

import "github.com/gin-gonic/gin"

type RouteRegister struct {
	EndpointAudit *EndpointAudit
}

type basePathRouter interface {
	BasePath() string
}

func fullPath(router gin.IRouter, relativePath string) string {
	if r, ok := router.(basePathRouter); ok {
		basePath := r.BasePath()
		if basePath == "/" {
			return relativePath
		}
		return basePath + relativePath
	}
	return relativePath
}

func (r *RouteRegister) POST(router gin.IRouter, relativePath string, handler gin.HandlerFunc, context APIContextType, deprecated bool) {
	router.POST(relativePath, handler)
	if deprecated {
		r.EndpointAudit.Add("POST", string(context), fullPath(router, relativePath))
	}
}

func (r *RouteRegister) GET(router gin.IRouter, relativePath string, handler gin.HandlerFunc, context APIContextType, deprecated bool) {
	router.GET(relativePath, handler)
	if deprecated {
		r.EndpointAudit.Add("GET", string(context), fullPath(router, relativePath))
	}
}

func (r *RouteRegister) PUT(router gin.IRouter, relativePath string, handler gin.HandlerFunc, context APIContextType, deprecated bool) {
	router.PUT(relativePath, handler)
	if deprecated {
		r.EndpointAudit.Add("PUT", string(context), fullPath(router, relativePath))
	}
}

func (r *RouteRegister) DELETE(router gin.IRouter, relativePath string, handler gin.HandlerFunc, context APIContextType, deprecated bool) {
	router.DELETE(relativePath, handler)
	if deprecated {
		r.EndpointAudit.Add("DELETE", string(context), fullPath(router, relativePath))
	}
}

func (r *RouteRegister) PATCH(router gin.IRouter, relativePath string, handler gin.HandlerFunc, context APIContextType, deprecated bool) {
	router.PATCH(relativePath, handler)
	if deprecated {
		r.EndpointAudit.Add("PATCH", string(context), fullPath(router, relativePath))
	}
}
