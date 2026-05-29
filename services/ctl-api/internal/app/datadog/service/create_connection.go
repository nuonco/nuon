package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// CreateConnectionRequest is the body for creating a per-org binding to a
// DD tenant.
//
// Site accepts either a known regional key (us1 / us3 / us5 / eu1 / ap1 /
// gov) or a full https URL for self-hosted DD. Validation lives on the
// model's BeforeSave so the same invariant is enforced regardless of the
// entry point.
//
// ApplicationKey is optional at create-time — the connection works for
// emit-only flows with just the API key. The one-click managed-monitor
// feature requires app_key; the dashboard surfaces a prompt to add one
// when it's missing.
//
// DefaultTags / DefaultNotifyHandles default to empty []string so the
// emit hot path doesn't have to nil-check.
type CreateConnectionRequest struct {
	Name                 string   `json:"name" validate:"required,min=1,max=128"`
	Site                 string   `json:"site" validate:"required"`
	APIKey               string   `json:"api_key" validate:"required,min=10"`
	ApplicationKey       string   `json:"application_key,omitempty"`
	Purpose              string   `json:"purpose,omitempty" validate:"omitempty,oneof=internal customer"`
	DefaultTags          []string `json:"default_tags,omitempty"`
	DefaultNotifyHandles []string `json:"default_notify_handles,omitempty"`
}

func (r *CreateConnectionRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateDatadogConnection
// @Summary				Create a Datadog connection
// @Description			Creates a per-org binding to a Datadog tenant. The connection's API key is verified against DD's /api/v1/validate endpoint before persistence — a bad key is rejected with a 400 so misconfigured connections never land in a "verified" state. Multiple connections per org are supported (typical layout: one for the vendor's own DD tenant + one per customer's DD tenant).
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string						true	"Org ID"
// @Param					req		body	CreateConnectionRequest		true	"Input"
// @Success				201	{object}	app.DatadogConnection
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/connections [POST]
func (s *service) CreateConnection(ctx *gin.Context) {
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	req := CreateConnectionRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	conn, err := s.createConnection(ctx, acct, org.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create datadog connection: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, conn)
}

// createConnection verifies the API key against DD, then persists a new
// DatadogConnection row. We intentionally validate BEFORE writing so a
// failed validation doesn't leave a half-configured row in the table that
// the lifecycle hook would skip but the dashboard would still surface.
func (s *service) createConnection(
	ctx context.Context,
	acct *app.Account,
	orgID string,
	req *CreateConnectionRequest,
) (*app.DatadogConnection, error) {
	// Probe the key before persisting. Network failures and 5xx bubble
	// up; a successful round-trip with valid=false is folded into an
	// invalid-request error.
	if err := s.validateAPIKey(ctx, req.Site, req.APIKey); err != nil {
		return nil, fmt.Errorf("datadog api key rejected: %w", err)
	}

	conn := app.DatadogConnection{
		OrgID:                orgID,
		Name:                 req.Name,
		Site:                 req.Site,
		APIKey:               req.APIKey,
		ApplicationKey:       req.ApplicationKey,
		Purpose:              app.DatadogConnectionPurpose(req.Purpose),
		DefaultTags:          req.DefaultTags,
		DefaultNotifyHandles: req.DefaultNotifyHandles,
		CreatedByID:          acct.ID,
	}

	if err := s.db.WithContext(ctx).Create(&conn).Error; err != nil {
		return nil, fmt.Errorf("create datadog connection: %w", err)
	}
	return &conn, nil
}
