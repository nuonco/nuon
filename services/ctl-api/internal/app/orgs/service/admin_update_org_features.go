package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminUpdateOrgFeaturesRequest struct {
	Features map[string]bool `json:"features" validate:"required"`
}

// @ID AdminUpdateOrgFeatures
// @Summary get available org features
// @Description.markdown admin_update_org_features.md
// @Param			org_id	path	string				true	"org ID"
// @Tags			orgs/admin
// @Security AdminEmail
// @Accept			json
// @Param			req	body AdminUpdateOrgFeaturesRequest	true	"Input"
// @Produce		json
// @Success		200				{object}	app.Org
// @Router			/v1/orgs/{org_id}/admin-features  [PATCH]
func (s *service) AdminUpdateOrgFeatures(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminUpdateOrgFeaturesRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable update org: %w", err))
		return
	}

	err = s.validateOrgFeatures(ctx, req.Features)
	if err != nil {
		ctx.Error(fmt.Errorf("unable update org: %w", err))
		return
	}

	err = s.updateOrgFeatures(ctx, org, req.Features)
	if err != nil {
		ctx.Error(fmt.Errorf("unable update org: %w", err))
		return
	}

	org, err = s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable update org: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, org)
}

func (s *service) validateOrgFeatures(ctx context.Context, features map[string]bool) error {
	orgFeatures := make(map[string]bool)
	if _, ok := features["all"]; ok {
		return nil
	}

	for _, value := range app.GetFeatures() {
		orgFeatures[string(value)] = true
	}
	for feature, _ := range features {
		if _, ok := orgFeatures[feature]; !ok {
			return fmt.Errorf("invalid feature: %s", feature)
		}
	}

	return nil
}

func (s *service) updateOrgFeatures(ctx context.Context, org *app.Org, updateFeatures map[string]bool) error {
	o := app.Org{
		ID: org.ID,
	}

	if allValue, ok := updateFeatures["all"]; ok {
		for feature := range org.Features {
			updateFeatures[feature] = allValue
		}
	} else {
		// add features from org.Features not in features
		for feature, enabled := range org.Features {
			if _, ok := updateFeatures[feature]; !ok {
				updateFeatures[feature] = enabled
			}
		}
	}

	res := s.db.WithContext(ctx).Model(&o).Updates(app.Org{
		Features: updateFeatures,
	})

	if res.Error != nil {
		return fmt.Errorf("unable to update org: %w", res.Error)
	}

	return nil
}
