package docs

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/admin"
	"github.com/powertoolsdev/mono/services/ctl-api/docs"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"

	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

type Route struct {
	cfg *internal.Config
}

func (r *Route) RegisterRoutes(g *gin.Engine) error {
	switch r.cfg.Env {
	case "development":
		docs.SwaggerInfo.Host = "http://localhost:8081"
	case "prod":
		docs.SwaggerInfo.Host = "https://ctl.prod.nuon.co"
	case "stage":
		docs.SwaggerInfo.Host = "https://ctl.stage.nuon.co"
	}

	g.GET("/docs/*any", swagger.WrapHandler(
		swaggerfiles.Handler,
	))
	return nil
}

func (r *Route) RegisterInternalRoutes(g *gin.Engine) error {
	switch r.cfg.Env {
	case "development":
		admin.SwaggerInfoadmin.Host = "localhost:8082"
	case "prod":
		admin.SwaggerInfoadmin.Host = "http://ctl.nuon.us-west-2.prod.nuon.cloud"
	case "stage":
		admin.SwaggerInfoadmin.Host = "http://ctl.nuon.us-west-2.stage.nuon.cloud"
	}

	docs.SwaggerInfo.Title = "Nuon Admin API"
	g.GET("/docs/*any", swagger.WrapHandler(
		swaggerfiles.Handler,
		swagger.InstanceName("admin"),
	))
	return nil
}

func New(cfg *internal.Config) *Route {
	return &Route{
		cfg: cfg,
	}
}
