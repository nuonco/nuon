package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Service) GetVersionHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"version":                 s.cfg.Version,
		"git_ref":                 s.cfg.GitRef,
		"recommended_cli_version": s.cfg.RecommendedCLIVersion,
	})
}
