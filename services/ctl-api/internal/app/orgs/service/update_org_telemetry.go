package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
	"gorm.io/gorm"
)

type UpdateOrgTelemetryRequest struct {
	Enabled *bool `json:"enabled" validate:"required"`
}

// @ID UpdateOrgTelemetry
// @Summary Update current org telemetry settings
// @Tags orgs
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param req body UpdateOrgTelemetryRequest true "Input"
// @Failure 400 {object} stderr.ErrResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Success 200 {object} app.Org
// @Router /v1/orgs/current/telemetry [PATCH]
func (s *service) UpdateOrgTelemetry(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
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

	var req UpdateOrgTelemetryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(validatorPkg.FormatValidationError(err))
		return
	}

	res := s.db.WithContext(ctx).Model(&app.Org{}).
		Where(app.Org{ID: org.ID}).Update("telemetry_enabled", *req.Enabled)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to update org telemetry: %w", res.Error))
		return
	}
	if res.RowsAffected != 1 {
		ctx.Error(fmt.Errorf("org not found: %w", gorm.ErrRecordNotFound))
		return
	}
	org, err = s.getOrg(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, org)
}
