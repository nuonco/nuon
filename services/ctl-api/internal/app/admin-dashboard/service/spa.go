package service

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
)

type adminClientConfig struct {
	AppURL string `json:"appUrl"`
}

func (s *service) registerSPARoutes(api *gin.Engine) {
	s.registerStaticSPA(api)
}

func (s *service) registerStaticSPA(e *gin.Engine) {
	distDir := s.cfg.AdminDashboardDistDir
	if distDir == "" {
		distDir = "./internal/app/admin-dashboard/dist"
	}

	distFS := os.DirFS(distDir)

	hasDistDir := true
	if _, err := fs.Stat(distFS, "."); err != nil {
		s.l.Warn("admin dashboard dist directory missing — static file serving disabled",
			zap.String("dist_dir", distDir), zap.Error(err))
		hasDistDir = false
	}

	var indexHTML []byte
	hasSPAFallback := false

	if hasDistDir {
		if _, err := fs.Stat(distFS, "index.html"); err != nil {
			s.l.Warn("no index.html in admin dashboard dist — SPA fallback disabled", zap.String("dist_dir", distDir))
		} else {
			raw, err := fs.ReadFile(distFS, "index.html")
			if err != nil {
				s.l.Error("failed to read admin dashboard index.html", zap.Error(err))
			} else {
				hasSPAFallback = true
				cc := adminClientConfig{
					AppURL: s.cfg.PublicAPIURL,
				}
				b, _ := json.Marshal(cc)
				script := fmt.Sprintf(`<script id="admin-config">window.__ADMIN_CONFIG__=%s;</script>`, b)
				indexHTML = bytes.Replace(raw, []byte("</head>"), []byte(script+"</head>"), 1)
			}
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
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})
}
