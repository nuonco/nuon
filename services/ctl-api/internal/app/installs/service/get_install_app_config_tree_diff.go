package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// InstallAppConfigTreeDiffResponse is the hierarchical config diff for one
// install. There is no single old config id: stack, sandbox, and each
// component are compared to their own applied app config.
type InstallAppConfigTreeDiffResponse struct {
	ConfigID string           `json:"config_id"`
	Diff     *diff.Diff       `json:"diff"`
	Summary  diff.DiffSummary `json:"summary"`
	Changed  string           `json:"changed"`
}

// @ID						GetInstallAppConfigTreeDiff
// @Summary				diff an app config against an install's applied entities
// @Description			Compares a new app config to the install. Stack, runner, sandbox, and each component use that entity's applied app config. An empty applied config compares that entity to nothing.
// @Param					install_id	path	string	true	"install ID"
// @Param					config_id	path	string	true	"new app config ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	service.InstallAppConfigTreeDiffResponse
// @Router					/v1/installs/{install_id}/app-configs/{config_id}/diff [GET]
func (s *service) GetInstallAppConfigTreeDiff(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	configID := ctx.Param("config_id")

	var install app.Install
	if err := s.db.WithContext(ctx.Request.Context()).
		Where(app.Install{ID: installID, OrgID: org.ID}).
		First(&install).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	var cfg app.AppConfig
	if err := s.db.WithContext(ctx.Request.Context()).
		Where(app.AppConfig{ID: configID, AppID: install.AppID, OrgID: org.ID}).
		First(&cfg).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get app config: %w", err))
		return
	}

	tree, err := s.helpers.CompositeAppConfigTree(ctx.Request.Context(), &install, configID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if tree == nil {
		tree = diff.NewDiff(diff.WithKey("app_config"))
	}

	ctx.JSON(http.StatusOK, InstallAppConfigTreeDiffResponse{
		ConfigID: configID,
		Diff:     tree,
		Summary:  tree.Summary(),
		Changed:  tree.FormatChanged(""),
	})
}
