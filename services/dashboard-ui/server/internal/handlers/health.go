package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

type HealthHandler struct {
	cfg *internal.Config
}

func NewHealthHandler(cfg *internal.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

func (h *HealthHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/", h.Index)
	e.GET("/health", h.Livez)
	e.GET("/livez", h.Livez)
	e.GET("/readyz", h.Readyz)
	e.GET("/version", h.Version)
	return nil
}

func (h *HealthHandler) Index(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Nuon Dashboard</title>
  <link rel="icon" href="/favicon.svg" type="image/svg+xml">
  <link rel="stylesheet" href="/styles.css">
</head>
<body>
  <h1>Nuon Dashboard</h1>
  <p>Service is running.</p>
  <script src="/app.js"></script>
</body>
</html>`))
}

func (h *HealthHandler) Livez(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *HealthHandler) Readyz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *HealthHandler) Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": h.cfg.Version,
		"git_ref": h.cfg.GitRef,
	})
}
