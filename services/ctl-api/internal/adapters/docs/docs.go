package docs

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/services/config"
	"github.com/powertoolsdev/mono/services/ctl-api/admin"
	"github.com/powertoolsdev/mono/services/ctl-api/docs"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"

	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

type Docs struct {
	cfg *internal.Config
}

func (r *Docs) RegisterRoutes(g *gin.Engine) error {
	docs.SwaggerInfo.Schemes = []string{"https"}
	docs.SwaggerInfo.Version = r.cfg.Version

	switch r.cfg.Env {
	case config.Development:
		docs.SwaggerInfo.Host = "localhost:8081"
		docs.SwaggerInfo.Schemes = []string{"http"}
	case config.Production:
		docs.SwaggerInfo.Host = "ctl.prod.nuon.co"
	case config.Stage:
		docs.SwaggerInfo.Host = "ctl.stage.nuon.co"
	}

	g.GET("/oapi/v3", r.getOAPI3publicSpec)
	g.GET("/oapi/v2", r.getOAPI2PublicSpec)
	g.GET("/docs/*any", swagger.WrapHandler(
		swaggerfiles.Handler,
	))

	return nil
}

func (r *Docs) RegisterInternalRoutes(g *gin.Engine) error {
	switch r.cfg.Env {
	case "development":
		admin.SwaggerInfoadmin.Host = "localhost:8082"
	case "prod":
		admin.SwaggerInfoadmin.Host = "ctl.nuon.us-west-2.prod.nuon.cloud"
	case "stage":
		admin.SwaggerInfoadmin.Host = "ctl.nuon.us-west-2.stage.nuon.cloud"
	}

	admin.SwaggerInfoadmin.Version = r.cfg.Version
	admin.SwaggerInfoadmin.Title = "Nuon Admin API"
	admin.SwaggerInfoadmin.Schemes = []string{"http"}

	g.GET("/oapi/v3", r.getOAPI3AdminSpec)
	g.GET("/oapi/v2", r.getOAPI2AdminSpec)
	g.GET("/docs/*any", swagger.WrapHandler(
		swaggerfiles.Handler,
		swagger.InstanceName("admin"),
	))

	return nil
}

func New(cfg *internal.Config) *Docs {
	return &Docs{
		cfg: cfg,
	}
}
