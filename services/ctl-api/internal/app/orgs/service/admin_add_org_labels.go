package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AdminAddOrgLabelsRequest struct {
	Labels map[string]string `json:"labels" validate:"required"`
}

func (r *AdminAddOrgLabelsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						AdminAddOrgLabels
// @Summary				add labels to an org
// @Description			Merge the provided labels into the org's existing labels. Existing keys are overwritten.
// @Param					org_id	path	string						true	"org ID"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminAddOrgLabelsRequest	true	"Input"
// @Produce				json
// @Success				200	{object}	app.Org
// @Router					/v1/orgs/{org_id}/admin-labels [POST]
func (s *service) AdminAddOrgLabels(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminAddOrgLabelsRequest
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

	org.Labels.Merge(labels.Labels(req.Labels))

	if err := s.db.WithContext(ctx).Model(org).Select("labels").Updates(org).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update org labels: %w", err))
		return
	}

	org, err = s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}
