package docs

import (
	"github.com/gin-gonic/gin"
	docs "github.com/powertoolsdev/mono/services/ctl-api/docs"
	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

type Route struct{}

func (r *Route) RegisterRoutes(g *gin.Engine) error {
	return nil
}

func (r *Route) RegisterInternalRoutes(g *gin.Engine) error {
	docs.SwaggerInfo.BasePath = "/api/v1"
	g.GET("/docs/*any", swagger.WrapHandler(swaggerfiles.Handler))
	return nil
}

func New() *Route {
	return &Route{}
}
