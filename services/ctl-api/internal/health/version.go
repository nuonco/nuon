package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Service) GetVersionHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"version":                          s.cfg.Version,
		"git_ref":                          s.cfg.GitRef,
		"server_side_sync_min_cli_version": s.cfg.ServerSideSyncMinCLIVersion,
	})
}
