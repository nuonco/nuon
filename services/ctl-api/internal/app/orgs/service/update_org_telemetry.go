package service

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateOrgTelemetryRequest struct {
	Enabled       *bool   `json:"enabled,omitempty" extensions:"x-nullable"`
	RelayEndpoint *string `json:"relay_endpoint,omitempty" extensions:"x-nullable"`
}

// @ID UpdateOrgTelemetry
// @Summary Update current org telemetry settings
// @Description Omitted fields are unchanged. Set relay_endpoint to null or an empty string to use the deployment default. Relay endpoints must be HTTPS OTLP base URLs.
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
func (s *service) UpdateOrgTelemetry(ctx *gin.Context) (require.Response, error) {
	access, err := require.MutationAccessFromContext(ctx)
	if err != nil {
		return require.Response{}, err
	}
	var req UpdateOrgTelemetryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.Response{}, stderr.NewInvalidRequest(err)
	}
	if err := s.v.Struct(&req); err != nil {
		return require.Response{}, validatorPkg.FormatValidationError(err)
	}

	updates := map[string]interface{}{}
	if req.Enabled != nil {
		updates["telemetry_enabled"] = *req.Enabled
	}
	patch := cctx.PatcherFromContext(ctx)
	if req.Enabled == nil && patch != nil && slices.Contains(patch.SelectFields, "enabled") {
		return require.Response{}, stderr.NewInvalidRequest(fmt.Errorf("enabled must be true or false"))
	}
	if req.RelayEndpoint != nil || (patch != nil && slices.Contains(patch.SelectFields, "relay_endpoint")) {
		if req.RelayEndpoint != nil && *req.RelayEndpoint == "" {
			req.RelayEndpoint = nil
		}
		if req.RelayEndpoint != nil {
			if err := app.ValidateTelemetryRelayEndpoint(*req.RelayEndpoint); err != nil {
				return require.Response{}, stderr.NewInvalidRequest(err)
			}
		}
		updates["telemetry_relay_endpoint"] = req.RelayEndpoint
	}
	if len(updates) == 0 {
		return require.Response{}, stderr.NewInvalidRequest(fmt.Errorf("provide enabled or relay_endpoint"))
	}

	tx := access.DB()
	orgID := access.Target().OrgID
	res := tx.Model(&app.Org{}).Where(app.Org{ID: orgID}).Updates(updates)
	if res.Error != nil {
		return require.Response{}, fmt.Errorf("unable to update org telemetry: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return require.Response{}, gorm.ErrRecordNotFound
	}
	org, err := getOrg(tx, orgID)
	if err != nil {
		return require.Response{}, err
	}
	return require.Response{Status: http.StatusOK, Body: org}, nil
}
