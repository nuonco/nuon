package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetLivezHandler reports whether the process is alive and the HTTP
// server can answer. It must remain dependency-free; failing this
// triggers a pod restart.
func (s *Service) GetLivezHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, map[string]any{
		"status": "ok",
		// only for backward compatibility, until dashboard migrates to /readyz
		"degraded": []string{},
	})
}
