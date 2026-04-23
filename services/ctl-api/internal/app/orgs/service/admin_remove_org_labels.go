package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AdminRemoveOrgLabelsRequest struct {
	Keys []string `json:"keys" validate:"required"`
}

func (r *AdminRemoveOrgLabelsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						AdminRemoveOrgLabels
// @Summary				remove labels from an org
// @Description			Remove the specified label keys from the org.
// @Param					org_id	path	string							true	"org ID"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminRemoveOrgLabelsRequest	true	"Input"
// @Produce				json
// @Success				200	{object}	app.Org
// @Router					/v1/orgs/{org_id}/admin-labels [DELETE]
func (s *service) AdminRemoveOrgLabels(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req AdminRemoveOrgLabelsRequest
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

	org.Labels.RemoveKeys(req.Keys)

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
