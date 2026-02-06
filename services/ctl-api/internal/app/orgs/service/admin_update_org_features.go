package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminUpdateOrgFeaturesRequest struct {
	Features map[string]bool `json:"features" validate:"required"`
}

// @ID						AdminUpdateOrgFeatures
// @Summary				update org features for a single org
// @Description.markdown	admin_update_org_features.md
// @Param					org_id	path	string	true	"org ID"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminUpdateOrgFeaturesRequest	true	"Input"
// @Produce				json
// @Success				200	{object}	app.Org
// @Router					/v1/orgs/{org_id}/admin-features  [PATCH]
func (s *service) AdminUpdateOrgFeatures(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminUpdateOrgFeaturesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %v", err),
		})
		return
	}

	// Validate required fields
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %v", err),
		})
		return
	}

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable update org: %w", err),
			Description: fmt.Sprintf("unable update org: %v", err),
		})
		return
	}

	if err := s.features.Enable(ctx, orgID, req.Features); err != nil {
		ctx.Error(errors.Wrap(err, "unable to enable org features"))
		return
	}

	org, err = s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable update org: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, org)
}
