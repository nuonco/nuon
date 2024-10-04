package docs

import (
	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/pkg/services/config"
	"github.com/powertoolsdev/mono/services/ctl-api/docs/admin"
	"github.com/powertoolsdev/mono/services/ctl-api/docs/public"
	"github.com/powertoolsdev/mono/services/ctl-api/docs/runner"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"

	swaggerfiles "github.com/swaggo/files"
	swagger "github.com/swaggo/gin-swagger"
)

type Docs struct {
	cfg *internal.Config
}

var _ api.Service = (*Docs)(nil)

func (r *Docs) RegisterPublicRoutes(g *gin.Engine) error {
	public.SwaggerInfo.Schemes = []string{"https"}
	public.SwaggerInfo.Version = r.cfg.Version

	switch r.cfg.Env {
	case config.Development:
		public.SwaggerInfo.Host = "localhost:8081"
		public.SwaggerInfo.Schemes = []string{"http"}
	case config.Production:
		public.SwaggerInfo.Host = "api.nuon.co"
	case config.Stage:
		public.SwaggerInfo.Host = "api.stage.nuon.co"
	}

	g.GET("/oapi/v3", r.getOAPI3publicSpec)
	g.GET("/oapi/v2", r.getOAPI2PublicSpec)
	g.GET("/docs/*any", swagger.WrapHandler(
		swaggerfiles.Handler,
		swagger.PersistAuthorization(true),
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

func (r *Docs) RegisterRunnerRoutes(g *gin.Engine) error {
	runner.SwaggerInforunner.Title = "Nuon Runner API"
	runner.SwaggerInforunner.Description = "Runner API"
	runner.SwaggerInforunner.Schemes = []string{"https"}
	runner.SwaggerInforunner.Version = r.cfg.Version

	switch r.cfg.Env {
	case config.Development:
		runner.SwaggerInforunner.Host = "localhost:8083"
		runner.SwaggerInforunner.Schemes = []string{"http"}
	case config.Production:
		runner.SwaggerInforunner.Host = "runner.nuon.co"
	case config.Stage:
		runner.SwaggerInforunner.Host = "runner.stage.nuon.co"
	}

	g.GET("/oapi/v3", r.getOAPI3RunnerSpec)
	g.GET("/oapi/v2", r.getOAPI2RunnerSpec)
	g.GET("/docs/*any", swagger.WrapHandler(
		swaggerfiles.Handler,
		swagger.PersistAuthorization(true),
		swagger.InstanceName("runner"),
	))

	return nil
}

func New(cfg *internal.Config) *Docs {
	return &Docs{
		cfg: cfg,
	}
}
