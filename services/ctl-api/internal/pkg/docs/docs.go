package docs

import (
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/services/config"
	"github.com/nuonco/nuon/services/ctl-api/docs/admin"
	"github.com/nuonco/nuon/services/ctl-api/docs/public"
	"github.com/nuonco/nuon/services/ctl-api/docs/runner"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"

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

	// Redoc - faster alternative to Swagger UI for large specs
	g.GET("/redoc", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, `<!DOCTYPE html>
<html>
<head>
    <title>Nuon API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <redoc spec-url='/oapi/v2' hide-download-button></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`)
	})

	// Custom fast Swagger UI with all performance optimizations
	g.GET("/swagger-fast", func(c *gin.Context) {
		c.Header("Content-Type", "text/html")
		c.String(200, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Nuon API - Fast Swagger UI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = () => {
            window.ui = SwaggerUIBundle({
                url: '/oapi/v2',
                dom_id: '#swagger-ui',
                deepLinking: false,
                displayRequestDuration: true,
                docExpansion: 'none',
                defaultModelsExpandDepth: -1,
                defaultModelExpandDepth: 0,
                filter: true,
                syntaxHighlight: false,
                tryItOutEnabled: false,
                persistAuthorization: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout"
            });
        };
    </script>
</body>
</html>`)
	})

	// Keep original Swagger UI for compatibility
	g.GET("/docs/*any", swagger.WrapHandler(
		swaggerfiles.Handler,
		swagger.PersistAuthorization(true),
		swagger.DocExpansion("none"),
		swagger.DeepLinking(false),
		swagger.DefaultModelsExpandDepth(-1),
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
		swagger.DocExpansion("none"),
		swagger.DeepLinking(false),
		swagger.DefaultModelsExpandDepth(-1),
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
		swagger.DocExpansion("none"),
		swagger.DeepLinking(false),
		swagger.DefaultModelsExpandDepth(-1),
	))

	return nil
}

func (s *Docs) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *Docs) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func New(cfg *internal.Config) *Docs {
	return &Docs{
		cfg: cfg,
	}
}
