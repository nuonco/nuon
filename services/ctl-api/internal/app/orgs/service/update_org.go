package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateOrgRequest struct {
	Name      string                     `json:"name" validate:"required_without=Telemetry"`
	Telemetry *UpdateOrgTelemetryRequest `json:"telemetry,omitempty"`
}

type UpdateOrgTelemetryRequest struct {
	Enabled *bool `json:"enabled" validate:"required"`
}

func (c *UpdateOrgRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						UpdateOrg
// @Summary				Update current org
// @Description.markdown	update_org.md
// @Param					req	body	UpdateOrgRequest	true	"Input"
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Org
// @Router					/v1/orgs/current [PATCH]
func (s *service) UpdateOrg(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req UpdateOrgRequest
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

	if req.Telemetry != nil {
		caller, err := cctx.AccountFromGinContext(ctx)
		if err != nil {
			ctx.Error(err)
			return
		}
		if !s.isOrgAdmin(caller, org.ID) {
			ctx.Error(stderr.ErrAuthorization{
				Err:         fmt.Errorf("only org admins can change the telemetry default"),
				Description: "only org admins can change the telemetry default",
			})
			return
		}
	}

	org, err = s.updateOrg(ctx, org.ID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}

func (s *service) updateOrg(ctx context.Context, orgID string, req *UpdateOrgRequest) (*app.Org, error) {
	updates := map[string]any{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Telemetry != nil {
		updates["telemetry_enabled"] = *req.Telemetry.Enabled
	}
	filter := app.Org{ID: orgID}
	res := s.db.WithContext(ctx).Model(&app.Org{}).Where(filter).Updates(updates)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("org not found: %w", gorm.ErrRecordNotFound)
	}
	var org app.Org
	if err := s.db.WithContext(ctx).Where(filter).First(&org).Error; err != nil {
		return nil, fmt.Errorf("unable to reload org: %w", err)
	}

	return &org, nil
}
