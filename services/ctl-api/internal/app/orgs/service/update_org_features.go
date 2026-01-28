package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type UpdateOrgFeaturesRequest struct {
	Features map[string]bool `json:"features" validate:"required"`
}

// @ID						UpdateOrgFeatures
// @Summary				update org features (requires user-managed-features flag)
// @Description.markdown	update_org_features.md
// @Tags					orgs
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Param					req	body	UpdateOrgFeaturesRequest	true	"Input"
// @Produce				json
// @Success				200	{object}	app.Org
// @Failure				403	{object}	stderr.ErrResponse
// @Router					/v1/orgs/current/features  [PATCH]
func (s *service) UpdateOrgFeatures(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Check if user-managed-features flag is enabled for this org
	if !org.Features[string(app.OrgFeatureUserManagedFeatures)] {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("user-managed-features flag is not enabled for this organization"),
			Description: "Your organization does not have permission to manage feature flags. Please contact support to enable this capability.",
		})
		return
	}

	var req UpdateOrgFeaturesRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	// Filter out any attempts to modify the user-managed-features flag itself
	// This flag can only be toggled by admins
	if _, exists := req.Features[string(app.OrgFeatureUserManagedFeatures)]; exists {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("the %s flag cannot be modified through this endpoint", app.OrgFeatureUserManagedFeatures),
			Description: fmt.Sprintf("The '%s' feature flag can only be modified by administrators. Please contact support if you need this flag enabled.", app.OrgFeatureUserManagedFeatures),
		})
		return
	}

	// Validate that all requested features are user-manageable
	manageableFeatures := app.GetUserManageableFeatures()
	manageableMap := make(map[string]bool)
	for _, feature := range manageableFeatures {
		manageableMap[string(feature)] = true
	}

	for featureName := range req.Features {
		if !manageableMap[featureName] {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("feature %s is not user-manageable", featureName),
				Description: fmt.Sprintf("The feature '%s' cannot be modified by users. This is an administrator-only feature flag.", featureName),
			})
			return
		}
	}

	// Update the features
	if err := s.features.Enable(ctx, org.ID, req.Features); err != nil {
		ctx.Error(errors.Wrap(err, "unable to update org features"))
		return
	}

	// Return updated org
	updatedOrg, err := s.getOrg(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to retrieve updated org: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, updatedOrg)
}
