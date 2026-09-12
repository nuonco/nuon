package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MaxRunnerVersion struct {
	Version string `json:"version"`
}

// @ID						GetMaxRunnerVersion
// @Summary				get max runner version
// @Description.markdown	get_max_runner_version.md
// @Tags					general/runner
// @Accept					json
// @Produce				json
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	MaxRunnerVersion
// @Router					/v1/general/max-runner-version [get]
func (s *service) GetMaxRunnerVersion(ctx *gin.Context) {
	version := s.cfg.MaxRunnerVersion
	if version == "" {
		version = s.cfg.Version
	}

	ctx.JSON(http.StatusOK, MaxRunnerVersion{Version: version})
}
