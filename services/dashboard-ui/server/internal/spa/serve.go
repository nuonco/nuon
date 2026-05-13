package spa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	authmw "github.com/nuonco/nuon/services/dashboard-ui/server/internal/middlewares/auth"
)

type clientConfig struct {
	APIUrl                string `json:"apiUrl"`
	TemporalUIUrl         string `json:"temporalUiUrl,omitempty"`
	AuthServiceUrl        string `json:"authServiceUrl,omitempty"`
	AppUrl                string `json:"appUrl"`
	GithubAppName         string `json:"githubAppName"`
	PylonAppID            string `json:"pylonAppId,omitempty"`
	DatadogEnv            string `json:"datadogEnv,omitempty"`
	DatadogAPIKey         string `json:"datadogApiKey,omitempty"`
	DatadogApplicationKey string `json:"datadogApplicationKey,omitempty"`
	DatadogTraceDebug     bool   `json:"datadogTraceDebug,omitempty"`
	DatadogAPIUrl         string `json:"datadogApiUrl,omitempty"`
	Version               string `json:"version,omitempty"`
	GitRef                string `json:"gitRef,omitempty"`
	IsBYOC                bool   `json:"isByoc"`
	SFTrialEndpoint       string `json:"sfTrialEndpoint,omitempty"`
	OnboardingV2          bool   `json:"onboardingV2,omitempty"`
	AdminDashboardUrl     string `json:"adminDashboardUrl,omitempty"`
}

func buildClientConfig(cfg *internal.Config) clientConfig {
	return clientConfig{
		APIUrl:                cfg.APIUrl,
		TemporalUIUrl:         cfg.TemporalUIUrl,
		AuthServiceUrl:        cfg.AuthServiceUrl,
		AppUrl:                cfg.AppUrl,
		GithubAppName:         cfg.GithubAppName,
		PylonAppID:            cfg.PylonAppID,
		DatadogEnv:            cfg.DatadogEnv,
		DatadogAPIKey:         cfg.DatadogAPIKey,
		DatadogApplicationKey: cfg.DatadogApplicationKey,
		DatadogTraceDebug:     cfg.DatadogTraceDebug,
		DatadogAPIUrl:         cfg.DatadogAPIUrl,
		Version:               cfg.Version,
		GitRef:                cfg.GitRef,
		IsBYOC:                cfg.IsBYOC,
		SFTrialEndpoint:       cfg.SFTrialEndpoint,
		OnboardingV2:          cfg.OnboardingV2,
		AdminDashboardUrl:     cfg.AdminDashboardUrl,
	}
}

// Handler serves SPA static assets and the index.html fallback.
type Handler struct {
	cfg       *internal.Config
	l         *zap.Logger
	indexHTML []byte
}

func NewHandler(cfg *internal.Config, l *zap.Logger) *Handler {
	return &Handler{cfg: cfg, l: l}
}

func (h *Handler) publicFS() fs.FS {
	publicDir := h.cfg.PublicDir
	if publicDir == "" {
		publicDir = "./public"
	}
	f := os.DirFS(publicDir)
	if _, err := fs.Stat(f, "."); err != nil {
		return nil
	}
	return f
}

// RegisterRoutes registers the SPA catch-all routes on the Gin engine.
// This MUST be called after all API routes are registered so that API routes
// take precedence.
func (h *Handler) RegisterRoutes(e *gin.Engine) error {
	distDir := h.cfg.DistDir
	if distDir == "" {
		distDir = "./dist"
	}

	distFS := os.DirFS(distDir)

	hasDistDir := true
	if _, err := fs.Stat(distFS, "."); err != nil {
		h.l.Warn("dist directory missing — static file serving from dist disabled",
			zap.String("dist_dir", distDir), zap.Error(err))
		hasDistDir = false
	}

	hasSPAFallback := false
	if hasDistDir {
		if _, err := fs.Stat(distFS, "index.html"); err != nil {
			h.l.Warn("no index.html in dist — SPA fallback disabled", zap.String("dist_dir", distDir))
		} else {
			hasSPAFallback = true
			raw, err := fs.ReadFile(distFS, "index.html")
			if err != nil {
				h.l.Error("failed to read index.html at startup", zap.Error(err))
				hasSPAFallback = false
			} else {
				cc := buildClientConfig(h.cfg)
				b, _ := json.Marshal(cc)
				h.l.Info("injecting client config into index.html", zap.String("apiUrl", cc.APIUrl), zap.String("appUrl", cc.AppUrl))
				script := fmt.Sprintf(`<script id="nuon-config">window.__NUON_CONFIG__=%s;</script>`, b)
				h.indexHTML = bytes.Replace(raw, []byte("</head>"), []byte(script+"</head>"), 1)
			}
		}
	}

	var distFileServer http.Handler
	if hasDistDir {
		distFileServer = http.FileServer(http.FS(distFS))
		e.GET("/assets/*filepath", func(c *gin.Context) {
			fp := c.Param("filepath")
			ct := ""
			if strings.HasSuffix(fp, ".js") {
				ct = "application/javascript"
			} else if strings.HasSuffix(fp, ".css") {
				ct = "text/css"
			}
			w := c.Writer
			if ct != "" {
				w = &mimeOverrideWriter{ResponseWriter: w, contentType: ct}
			}
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			distFileServer.ServeHTTP(w, c.Request)
		})
	}

	publicFS := h.publicFS()
	authHandler := authmw.New(h.cfg, h.l).Handler()

	e.NoRoute(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Status(http.StatusNotFound)
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		filePath := strings.TrimPrefix(c.Request.URL.Path, "/")

		if publicFS != nil {
			if _, err := fs.Stat(publicFS, filePath); err == nil {
				http.FileServer(http.FS(publicFS)).ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		if hasDistDir && filePath != "" && filePath != "index.html" {
			if _, err := fs.Stat(distFS, filePath); err == nil {
				distFileServer.ServeHTTP(c.Writer, c.Request)
				return
			}
		}

		authHandler(c)
		if c.IsAborted() {
			return
		}

		if !hasSPAFallback {
			c.Status(http.StatusNotFound)
			return
		}

		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Data(http.StatusOK, "text/html; charset=utf-8", h.indexHTML)
	})

	return nil
}

// mimeOverrideWriter forces a Content-Type on the response, preventing
// http.FileServer from sniffing and setting an incorrect MIME type.
type mimeOverrideWriter struct {
	gin.ResponseWriter
	contentType string
	wroteHeader bool
}

func (w *mimeOverrideWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.Header().Set("Content-Type", w.contentType)
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *mimeOverrideWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
