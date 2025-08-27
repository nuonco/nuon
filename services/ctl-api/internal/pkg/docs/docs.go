package docs

import (
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/services/config"
	"github.com/powertoolsdev/mono/services/ctl-api/docs/admin"
	"github.com/powertoolsdev/mono/services/ctl-api/docs/public"
	"github.com/powertoolsdev/mono/services/ctl-api/docs/runner"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"

	swagger "github.com/nuonco/gin-swagger"
	swaggerfiles "github.com/swaggo/files"
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
	default:
		u, err := url.Parse(r.cfg.PublicAPIURL)
		if err != nil {
			return errors.Wrap(err, "unable to parse public api url")
		}
		public.SwaggerInfo.Host = u.Host
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
	case config.Development:
		admin.SwaggerInfoadmin.Host = "localhost:8082"
	default:
		u, err := url.Parse(r.cfg.AdminAPIURL)
		if err != nil {
			return errors.Wrap(err, "unable to parse admin api url")
		}
		admin.SwaggerInfoadmin.Host = u.Host

		if u.Path != "" {
			admin.SwaggerInfoadmin.BasePath = u.Path
		}
	}

	admin.SwaggerInfoadmin.Version = r.cfg.Version
	admin.SwaggerInfoadmin.Title = "Nuon Admin API"

	g.GET("/oapi/v3", r.getOAPI3AdminSpec)
	g.GET("/oapi/v2", r.getOAPI2AdminSpec)
	g.GET("/docs/*any", swagger.WrapHandler(
		swaggerfiles.Handler,
		swagger.InstanceName("admin"),
		swagger.PersistAuthorization(true),
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
	default:
		u, err := url.Parse(r.cfg.RunnerAPIURL)
		if err != nil {
			return errors.Wrap(err, "unable to parse runner api url")
		}
		admin.SwaggerInfoadmin.Host = u.Host
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
