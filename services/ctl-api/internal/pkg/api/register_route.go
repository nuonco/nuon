package api

import "github.com/gin-gonic/gin"

//

type RouteRegister struct {
	EndpointAudit *EndpointAudit
}

func (r *RouteRegister) POST(api *gin.Engine, relativePath string, handler gin.HandlerFunc, context APIContextType, deprecated bool) {
	api.POST(relativePath, handler)
	if deprecated {
		r.EndpointAudit.Add("POST", string(context), relativePath)
	}
}

func (r *RouteRegister) GET(api *gin.Engine, relativePath string, handler gin.HandlerFunc, context APIContextType, deprecated bool) {
	api.GET(relativePath, handler)
	if deprecated {
		r.EndpointAudit.Add("GET", string(context), relativePath)
	}
}

func (r *RouteRegister) PUT(api *gin.Engine, relativePath string, handler gin.HandlerFunc, context APIContextType, deprecated bool) {
	api.PUT(relativePath, handler)
	if deprecated {
		r.EndpointAudit.Add("PUT", string(context), relativePath)
	}
}

func (r *RouteRegister) DELETE(api *gin.Engine, relativePath string, handler gin.HandlerFunc, context APIContextType, deprecated bool) {
	api.DELETE(relativePath, handler)
	if deprecated {
		r.EndpointAudit.Add("DELETE", string(context), relativePath)
	}
}

func (r *RouteRegister) PATCH(api *gin.Engine, relativePath string, handler gin.HandlerFunc, context APIContextType, deprecated bool) {
	api.PATCH(relativePath, handler)
	if deprecated {
		r.EndpointAudit.Add("PATCH", string(context), relativePath)
	}
}
