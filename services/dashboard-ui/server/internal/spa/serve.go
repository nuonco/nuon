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
	if h.cfg.DashboardDev {
		h.l.Info("dashboard dev mode: SPA requests will be proxied to Vite dev server")
		return h.registerDevProxy(e)
	}

	return h.registerStatic(e)
}

// registerStatic serves the SPA from the dist directory on disk.
// Static files (favicons, robots.txt, etc.) are served from the public directory.
func (h *Handler) registerStatic(e *gin.Engine) error {
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
		}
	}

	var distFileServer http.Handler
	if hasDistDir {
		distFileServer = http.FileServer(http.FS(distFS))
		e.GET("/assets/*filepath", func(c *gin.Context) {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
			distFileServer.ServeHTTP(c.Writer, c.Request)
		})
	}

	publicFS := h.publicFS()

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

		if !hasSPAFallback {
			c.Status(http.StatusNotFound)
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

// registerDevProxy proxies non-API requests to the dev server.
// Static files from the public directory are served directly without proxying.
func (h *Handler) registerDevProxy(e *gin.Engine) error {
	publicFS := h.publicFS()

	e.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		if publicFS != nil {
			filePath := strings.TrimPrefix(c.Request.URL.Path, "/")
			if _, err := fs.Stat(publicFS, filePath); err == nil {
				http.FileServer(http.FS(publicFS)).ServeHTTP(c.Writer, c.Request)
				return
			}
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
