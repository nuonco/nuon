package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AdminUpdateOrgFeaturesRequest struct {
	Features map[string]bool `json:"features" validate:"required"`
}

func (r *AdminUpdateOrgFeaturesRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
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
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.features.Enable(ctx, orgID, req.Features); err != nil {
		ctx.Error(errors.Wrap(err, "unable to enable org features"))
		return
	}

	org, err = s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}
