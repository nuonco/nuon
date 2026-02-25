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
	e.GET("/health", h.Livez)
	e.GET("/livez", h.Livez)
	e.GET("/readyz", h.Readyz)
	e.GET("/version", h.Version)
	return nil
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
