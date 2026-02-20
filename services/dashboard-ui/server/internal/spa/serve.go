package spa

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

// Handler serves SPA static assets and the index.html fallback.
type Handler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewHandler(cfg *internal.Config, l *zap.Logger) *Handler {
	return &Handler{cfg: cfg, l: l}
}

// RegisterRoutes registers the SPA catch-all routes on the Gin engine.
// This MUST be called after all API routes are registered so that API routes
// take precedence.
func (h *Handler) RegisterRoutes(e *gin.Engine) error {
	if h.cfg.DashboardDev {
		h.l.Info("dashboard dev mode: SPA requests will be proxied to Vite dev server")
		return h.registerDevProxy(e)
	}

	return h.registerStatic(e)
}

// registerStatic serves the SPA from the dist directory on disk.
// In production, the Dockerfile copies the Vite build output to a known path.
// The config's DistDir field controls where to find it (default: "./dist").
func (h *Handler) registerStatic(e *gin.Engine) error {
	distDir := h.cfg.DistDir
	if distDir == "" {
		distDir = "./dist"
	}

	distFS := os.DirFS(distDir)

	// Verify dist directory exists and contains index.html.
	if _, err := fs.Stat(distFS, "index.html"); err != nil {
		h.l.Warn("dist directory missing or no index.html — SPA serving disabled",
			zap.String("dist_dir", distDir), zap.Error(err))
		return nil
	}

	fileServer := http.FileServer(http.FS(distFS))

	// Serve /assets/* with aggressive caching — Vite produces
	// content-hashed filenames so these are immutable.
	e.GET("/assets/*filepath", func(c *gin.Context) {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// SPA fallback: any unmatched GET request serves index.html with
	// no-cache so the browser always fetches the latest version which
	// references the current hashed asset bundles.
	e.NoRoute(func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Status(http.StatusNotFound)
			return
		}
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Content-Type", "text/html; charset=utf-8")

		f, err := distFS.Open("index.html")
		if err != nil {
			h.l.Error("failed to open index.html", zap.Error(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		defer f.Close()

		c.Status(http.StatusOK)
		io.Copy(c.Writer, f)
	})

	return nil
}

// registerDevProxy proxies non-API requests to the Vite dev server for HMR.
func (h *Handler) registerDevProxy(e *gin.Engine) error {
	e.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		proxy := &http.Transport{}
		target := "http://localhost:5173" + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			target += "?" + c.Request.URL.RawQuery
		}

		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, target, c.Request.Body)
		if err != nil {
			c.Status(http.StatusBadGateway)
			return
		}
		req.Header = c.Request.Header

		resp, err := proxy.RoundTrip(req)
		if err != nil {
			h.l.Warn("vite dev server proxy error", zap.Error(err))
			c.Status(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for k, vs := range resp.Header {
			for _, v := range vs {
				c.Writer.Header().Add(k, v)
			}
		}
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	})

	return nil
}
