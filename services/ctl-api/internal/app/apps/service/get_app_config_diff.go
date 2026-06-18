package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type AppConfigDiffResponse struct {
	ConfigID    string           `json:"config_id"`
	OldConfigID string           `json:"old_config_id,omitempty"`
	Diff        *diff.Diff       `json:"diff"`
	Summary     diff.DiffSummary `json:"summary"`
	Changed     string           `json:"changed"`
}

// @ID						GetAppConfigDiff
// @Summary				diff two app configs
// @Description			Compares a new app config against an old one and returns a hierarchical diff.
// @Param					app_id		path	string	true	"app ID"
// @Param					config_id	path	string	true	"new config ID"
// @Param					old_config_id	query	string	false	"previous config ID to compare against"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	AppConfigDiffResponse
// @Router					/v1/apps/{app_id}/configs/{config_id}/diff [get]
func (s *service) GetAppConfigDiff(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	appID := ctx.Param("app_id")
	configID := ctx.Param("config_id")
	oldConfigID := ctx.Query("old_config_id")

	// Load the new config's intermediate representation
	newCfg, err := s.loadIntermediateConfig(ctx, org.ID, appID, configID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to load new config: %w", err))
		return
	}

	// Load the old config (nil if not specified — diff will show everything as added)
	var oldCfg *config.AppConfig
	if oldConfigID != "" {
		oldCfg, err = s.loadIntermediateConfig(ctx, org.ID, appID, oldConfigID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to load old config: %w", err))
			return
		}
	}

	// Compute diff
	d := newCfg.Diff(oldCfg)
	summary := d.Summary()

	ctx.JSON(http.StatusOK, AppConfigDiffResponse{
		ConfigID:    configID,
		OldConfigID: oldConfigID,
		Diff:        d,
		Summary:     summary,
		Changed:     d.FormatChanged(""),
	})
}

// loadIntermediateConfig fetches an app config from the DB and deserializes
// its intermediate config blob into the config.AppConfig struct.
func (s *service) loadIntermediateConfig(ctx *gin.Context, orgID, appID, configID string) (*config.AppConfig, error) {
	var appCfg app.AppConfig
	res := s.db.WithContext(ctx).
		Where("id = ? AND org_id = ? AND app_id = ?", configID, orgID, appID).
		First(&appCfg)
	if res.Error != nil {
		return nil, fmt.Errorf("config not found: %w", res.Error)
	}

	if appCfg.IntermediateConfig == nil {
		return nil, fmt.Errorf("config %s has no intermediate config", configID)
	}

	intermediateJSON, err := appCfg.IntermediateConfig.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load intermediate config: %w", err)
	}

	var cfg config.AppConfig
	if err := json.Unmarshal([]byte(intermediateJSON), &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse intermediate config: %w", err)
	}

	return &cfg, nil
}
