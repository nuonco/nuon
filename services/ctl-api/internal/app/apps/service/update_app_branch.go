package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type UpdateAppBranchRequest struct {
	Name string `json:"name" validate:"required,min=1"`
}

func (c *UpdateAppBranchRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return err
	}
	return nil
}

// @ID						UpdateAppBranch
// @Summary				update app branch metadata
// @Description			Updates app branch metadata (name only). To update configuration, create a new AppBranchConfig via POST /branches/:id/configs
// @Tags					apps
// @Accept					json
// @Param					req				body	UpdateAppBranchRequest	true	"Input"
// @Param					app_id			path	string					true	"app ID"
// @Param					app_branch_id	path	string					true	"app branch ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.AppBranch
// @Router					/v1/apps/{app_id}/branches/{app_branch_id} [patch]
func (s *service) UpdateAppBranch(ctx *gin.Context) {
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

	var req UpdateAppBranchRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Verify branch exists and belongs to this org/app
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

	// Check for name uniqueness within the app
	var existingBranch app.AppBranch
	res = s.db.WithContext(ctx).
		Where(app.AppBranch{
			AppID: appID,
			Name:  req.Name,
		}).
		Where("id != ?", appBranchID).
		First(&existingBranch)
	if res.Error == nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("branch name already exists: %s", req.Name),
			Description: "a branch with this name already exists for this app",
		})
		return
	}

	// Update only the name field
	branch.Name = req.Name
	res = s.db.WithContext(ctx).Save(&branch)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to update app branch: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, branch)
}
