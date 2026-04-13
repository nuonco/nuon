package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CLIConfig struct {
	DashboardURL    string `json:"dashboard_url"`
	NuonAuthEnabled bool   `json:"nuon_auth_enabled"`
	RootDomain      string `json:"root_domain"`
}

// @ID						GetCLIConfig
// @Summary				Get config for cli
// @Description.markdown	get_cli_config.md
// @Tags					general
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	CLIConfig
// @Router					/v1/general/cli-config [GET]
func (s *service) GetCLIConfig(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, &CLIConfig{
		DashboardURL:    s.cfg.AppURL,
		NuonAuthEnabled: s.cfg.NuonAuthProviderType != "",
		RootDomain:      s.cfg.RootDomain,
	})
}
