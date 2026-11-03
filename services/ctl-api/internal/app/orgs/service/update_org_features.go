package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateOrgFeaturesRequest struct {
	Features map[string]bool `json:"features" validate:"required"`
}

func (r *UpdateOrgFeaturesRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
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
func (s *service) UpdateOrgFeatures(ctx *gin.Context) (require.Response, error) {
	access, err := require.MutationAccessFromContext(ctx)
	if err != nil {
		return require.Response{}, err
	}
	tx := access.DB()
	orgID := access.Target().OrgID
	var org app.Org
	if err := tx.Where(app.Org{ID: orgID}).Take(&org).Error; err != nil {
		return require.Response{}, err
	}

	// Check if user-managed-features flag is enabled for this org
	if !org.Features[string(app.OrgFeatureUserManagedFeatures)] {
		return require.Response{}, stderr.ErrUser{
			Err:         fmt.Errorf("user-managed-features flag is not enabled for this organization"),
			Description: "Your organization does not have permission to manage feature flags. Please contact support to enable this capability.",
		}
	}

	var req UpdateOrgFeaturesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.Response{}, stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		}
	}
	if err := req.Validate(s.v); err != nil {
		return require.Response{}, fmt.Errorf("invalid request: %w", err)
	}

	// Filter out any attempts to modify the user-managed-features flag itself
	// This flag can only be toggled by admins
	if _, exists := req.Features[string(app.OrgFeatureUserManagedFeatures)]; exists {
		return require.Response{}, stderr.ErrUser{
			Err:         fmt.Errorf("the %s flag cannot be modified through this endpoint", app.OrgFeatureUserManagedFeatures),
			Description: fmt.Sprintf("The '%s' feature flag can only be modified by administrators. Please contact support if you need this flag enabled.", app.OrgFeatureUserManagedFeatures),
		}
	}

	forced := app.ForcedFeatures()
	for featureName := range req.Features {
		if forced[featureName] {
			return require.Response{}, stderr.ErrUser{
				Err:         fmt.Errorf("feature %s is forced on for all organizations", featureName),
				Description: fmt.Sprintf("The '%s' feature flag is enabled for all organizations and cannot be changed.", featureName),
			}
		}
	}

	// Validate that all requested features are user-manageable
	manageableFeatures := app.GetUserManageableFeatures()
	manageableMap := make(map[string]bool)
	for _, feature := range manageableFeatures {
		manageableMap[string(feature)] = true
	}

	for featureName := range req.Features {
		if !manageableMap[featureName] {
			return require.Response{}, stderr.ErrUser{
				Err:         fmt.Errorf("feature %s is not user-manageable", featureName),
				Description: fmt.Sprintf("The feature '%s' cannot be modified by users. This is an administrator-only feature flag.", featureName),
			}
		}
	}

	// Update the features
	err = features.New(features.Params{DB: tx, Cfg: s.cfg, V: s.v}).Enable(tx.Statement.Context, orgID, req.Features)
	if err != nil {
		return require.Response{}, errors.Wrap(err, "unable to update org features")
	}

	// Return updated org
	updatedOrg, err := getOrg(tx, orgID)
	if err != nil {
		return require.Response{}, fmt.Errorf("unable to retrieve updated org: %w", err)
	}

	return require.Response{Status: http.StatusOK, Body: updatedOrg}, nil
}
