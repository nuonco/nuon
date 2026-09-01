package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type UpdateAppBranchConfigRequest struct {
	// IgnoreChangesRegex marks a run not-attempted when every changed file path in
	// it matches this RE2 pattern. Send an empty string to clear it.
	IgnoreChangesRegex *string `json:"ignore_changes_regex,omitempty"`

	// SendStatusesOnIgnore posts a successful commit status for ignored runs.
	SendStatusesOnIgnore *bool `json:"send_statuses_on_ignore,omitempty"`
}

func (c *UpdateAppBranchConfigRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return err
	}
	if c.IgnoreChangesRegex == nil && c.SendStatusesOnIgnore == nil {
		return fmt.Errorf("at least one field must be provided")
	}
	if c.IgnoreChangesRegex != nil {
		if err := helpers.ValidateIgnoreChangesRegex(*c.IgnoreChangesRegex); err != nil {
			return err
		}
	}
	return nil
}

// @ID						UpdateAppBranchConfig
// @Summary				update app branch config settings
// @Description			Updates mutable settings on an existing app branch config (e.g. webhook trigger behavior).
// @Tags					apps
// @Accept					json
// @Param					req				body	UpdateAppBranchConfigRequest	true	"Input"
// @Param					app_id			path	string							true	"app ID"
// @Param					app_branch_id	path	string							true	"app branch ID"
// @Param					config_id		path	string							true	"app branch config ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.AppBranchConfig
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/configs/{config_id} [patch]
func (s *service) UpdateAppBranchConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	enabled, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check features: %w", err))
		return
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppBranches))
		return
	}

	appID := ctx.Param("app_id")
	appBranchID := ctx.Param("app_branch_id")
	configID := ctx.Param("config_id")

	var req UpdateAppBranchConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{
			OrgID: org.ID,
			AppID: appID,
		}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	var config app.AppBranchConfig
	res = s.db.WithContext(ctx).
		Where(app.AppBranchConfig{
			OrgID:       org.ID,
			AppBranchID: appBranchID,
		}).
		First(&config, "id = ?", configID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch config: %w", res.Error))
		return
	}

	if req.IgnoreChangesRegex != nil {
		config.IgnoreChangesRegex = *req.IgnoreChangesRegex
	}
	if req.SendStatusesOnIgnore != nil {
		config.SendStatusesOnIgnore = *req.SendStatusesOnIgnore
	}

	res = s.db.WithContext(ctx).Save(&config)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to update app branch config: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, config)
}
