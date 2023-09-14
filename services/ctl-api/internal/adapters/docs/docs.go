package docs

import (
	"github.com/gin-gonic/gin"
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
	case "developnent":
		docs.SwaggerInfo.Host = "localhost:8081"
	case "prod":
		docs.SwaggerInfo.Host = "ctl.prod.nuon.co"
	case "stage":
		docs.SwaggerInfo.Host = "ctl.stage.nuon.co"
	}

	g.GET("/docs/*any", swagger.WrapHandler(swaggerfiles.Handler))
	return nil
}

func (r *Route) RegisterInternalRoutes(g *gin.Engine) error {
	return nil
}

func New(cfg *internal.Config) *Route {
	return &Route{
		cfg: cfg,
	}
}
