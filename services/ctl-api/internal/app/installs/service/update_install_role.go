package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateInstallRoleRequest struct {
	Enabled *bool `json:"enabled" validate:"required"`
}

func (c *UpdateInstallRoleRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						UpdateInstallRole
// @Summary				update an install role
// @Description			enable or disable an install role
// @Tags					installs
// @Param					install_id	path	string	true	"install ID"
// @Param					role_id		path	string	true	"install role ID"
// @Accept					json
// @Param					req	body	UpdateInstallRoleRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallRoles
// @Router					/v1/installs/{install_id}/roles/{role_id} [patch]
func (s *service) UpdateInstallRole(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	roleID := ctx.Param("role_id")

	var req UpdateInstallRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	role, err := s.updateInstallRole(ctx, org.ID, installID, roleID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update install role: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, role)
}

func (s *service) updateInstallRole(ctx *gin.Context, orgID, installID, roleID string, req *UpdateInstallRoleRequest) (*app.InstallRoles, error) {
	var role app.InstallRoles
	res := s.db.WithContext(ctx).
		Where("id = ? AND install_id = ? AND org_id = ?", roleID, installID, orgID).
		First(&role)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find install role: %w", res.Error)
	}

	res = s.db.WithContext(ctx).
		Model(&role).
		Update("enabled", *req.Enabled)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update install role: %w", res.Error)
	}

	return &role, nil
}
