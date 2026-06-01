package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type CreateAppBranchRequest struct {
	Name string `json:"name" validate:"required,min=1"`
}

func (c *CreateAppBranchRequest) Validate(v *validator.Validate) error {
	return v.Struct(c)
}

// @ID						CreateAppBranch
// @Description.markdown	create_app_branch.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppBranchRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppBranch
// @Router					/v1/apps/{app_id}/branches [post]
func (s *service) CreateAppBranch(ctx *gin.Context) {
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

	var req CreateAppBranchRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Create app branch (VCS config is set via AppBranchConfig, not at branch creation time)
	branch, err := s.helpers.CreateAppBranch(ctx, appID, req.Name)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app branch: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, branch)
}
