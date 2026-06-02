package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// UpdateEnvAccentConfigRequest is the PUT-style replacement body for the
// org's env_accent_config. Sending an empty values map clears the mapping
// (everything renders neutral). The handler validates the color allowlist
// before persisting.
type UpdateEnvAccentConfigRequest struct {
	LabelKey string                        `json:"label_key"`
	Values   map[string]app.EnvAccentColor `json:"values"`
}

// @ID						UpdateEnvAccentConfig
// @Summary				Update env accent config
// @Description			Replaces the env accent config used to paint install indicators in the dashboard.
// @Param					req	body	UpdateEnvAccentConfigRequest	true	"Input"
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Org
// @Router					/v1/orgs/current/env-accent-config [PUT]
func (s *service) UpdateEnvAccentConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req UpdateEnvAccentConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}

	cfg := app.EnvAccentConfig{
		LabelKey: req.LabelKey,
		Values:   req.Values,
	}
	if err := cfg.Validate(); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         err,
			Description: err.Error(),
		})
		return
	}

	updated, err := s.updateEnvAccentConfig(ctx, org.ID, cfg)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

func (s *service) updateEnvAccentConfig(ctx context.Context, orgID string, cfg app.EnvAccentConfig) (*app.Org, error) {
	res := s.db.WithContext(ctx).
		Model(&app.Org{}).
		Where(&app.Org{ID: orgID}).
		Update("env_accent_config", cfg)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update env accent config: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("org not found %w", gorm.ErrRecordNotFound)
	}

	var org app.Org
	if err := s.db.WithContext(ctx).Where(&app.Org{ID: orgID}).First(&org).Error; err != nil {
		return nil, fmt.Errorf("unable to reload org: %w", err)
	}
	return &org, nil
}
