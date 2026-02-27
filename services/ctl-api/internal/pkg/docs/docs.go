package docs

import (
	"fmt"
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

// getOverrideSwaggerHTML returns optimized Swagger UI HTML with custom title
func getOverrideSwaggerHTML(title string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>%s</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.31.2/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.31.2/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.31.2/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = () => {
            window.ui = SwaggerUIBundle({
                url: '/oapi/v2',
                dom_id: '#swagger-ui',
                deepLinking: true,
                displayRequestDuration: true,
                docExpansion: 'none',
                defaultModelsExpandDepth: -1,
                defaultModelExpandDepth: 0,
                filter: true,
                syntaxHighlight: false,
                tryItOutEnabled: true,
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
</html>`, title)
}

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

	// Create handler that serves override HTML for index.html, passes through for other assets
	swaggerAssetsHandler := swagger.WrapHandler(swaggerfiles.Handler)
	overrideSwaggerHTML := getOverrideSwaggerHTML("Nuon API Documentation")

	customDocsHandler := func(c *gin.Context) {
		// Intercept index.html requests and serve override version
		if c.Request.URL.Path == "/docs/" || c.Request.URL.Path == "/docs/index.html" {
			c.Header("Content-Type", "text/html")
			c.String(200, overrideSwaggerHTML)
			return
		}
		// Pass through to original handler for CSS, JS, and other assets
		swaggerAssetsHandler(c)
	}

	g.GET("/docs/*any", customDocsHandler)

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

	// Handler for admin docs
	adminAssetsHandler := swagger.WrapHandler(swaggerfiles.Handler, swagger.InstanceName("admin"))
	overrideAdminHTML := getOverrideSwaggerHTML("Nuon Admin API Documentation")

	customAdminDocsHandler := func(c *gin.Context) {
		if c.Request.URL.Path == "/docs/" || c.Request.URL.Path == "/docs/index.html" {
			c.Header("Content-Type", "text/html")
			c.String(200, overrideAdminHTML)
			return
		}
		adminAssetsHandler(c)
	}

	g.GET("/docs/*any", customAdminDocsHandler)

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

	// Handler for runner docs
	runnerAssetsHandler := swagger.WrapHandler(swaggerfiles.Handler, swagger.InstanceName("runner"))
	overrideRunnerHTML := getOverrideSwaggerHTML("Nuon Runner API Documentation")

	customRunnerDocsHandler := func(c *gin.Context) {
		if c.Request.URL.Path == "/docs/" || c.Request.URL.Path == "/docs/index.html" {
			c.Header("Content-Type", "text/html")
			c.String(200, overrideRunnerHTML)
			return
		}
		runnerAssetsHandler(c)
	}

	g.GET("/docs/*any", customRunnerDocsHandler)

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
