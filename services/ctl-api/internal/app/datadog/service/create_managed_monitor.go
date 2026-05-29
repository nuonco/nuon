package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	ddclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// CreateManagedMonitorRequest backs the dashboard's one-click
// "Alert in Datadog" button. TargetType + TargetID identify the Nuon
// resource to watch; Preset chooses a query template
// (failure / drift). NotifyHandles, when set, override the connection's
// DefaultNotifyHandles for this specific monitor — the common case is
// "use the connection default", but a per-install override is sometimes
// useful for high-criticality customers.
//
// DisplayName is optional and surfaces in the monitor's name + body in
// DD. Without it we fall back to the raw target ID, which still works
// but reads worse in DD's UI.
type CreateManagedMonitorRequest struct {
	ConnectionID  string                              `json:"connection_id" validate:"required"`
	TargetType    app.DatadogManagedMonitorTargetType `json:"target_type" validate:"required,oneof=action install component workflow"`
	TargetID      string                              `json:"target_id" validate:"required"`
	Preset        app.DatadogManagedMonitorPreset     `json:"preset" validate:"required,oneof=failure drift"`
	DisplayName   string                              `json:"display_name,omitempty"`
	NotifyHandles []string                            `json:"notify_handles,omitempty"`
}

func (r *CreateManagedMonitorRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateDatadogManagedMonitor
// @Summary				Create a one-click Datadog managed monitor
// @Description			Creates a DD monitor that fires on Nuon events for the given target. The target's parent connection must belong to the calling org and must have an application key (the DD Monitors API requires both keys). The (connection, target, preset) tuple is unique — clicking the button twice returns the existing row instead of creating a duplicate.
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string							true	"Org ID"
// @Param					req		body	CreateManagedMonitorRequest		true	"Input"
// @Success				201	{object}	app.DatadogManagedMonitor
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/managed-monitors [POST]
func (s *service) CreateManagedMonitor(ctx *gin.Context) {
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

	req := CreateManagedMonitorRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	monitor, err := s.createManagedMonitor(ctx, acct, org.ID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, monitor)
}

func (s *service) createManagedMonitor(
	ctx context.Context,
	acct *app.Account,
	orgID string,
	req *CreateManagedMonitorRequest,
) (*app.DatadogManagedMonitor, error) {
	var conn app.DatadogConnection
	if err := s.db.WithContext(ctx).
		Where(app.DatadogConnection{ID: req.ConnectionID, OrgID: orgID}).
		First(&conn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("datadog connection %q not found in org %q", req.ConnectionID, orgID),
				Description: "Datadog connection not found",
			}
		}
		return nil, fmt.Errorf("lookup datadog connection: %w", err)
	}
	if conn.Status != app.DatadogConnectionStatusVerified {
		return nil, stderr.NewInvalidRequest(fmt.Errorf("connection %q is not verified", conn.ID))
	}
	if conn.ApplicationKey == "" {
		// Surface this as a 400 with a clear description — the
		// dashboard's one-click button checks the same field before
		// rendering the action, so reaching this branch usually means
		// a race between key removal and the click.
		return nil, stderr.NewInvalidRequest(fmt.Errorf(
			"connection %q has no application key; the DD Monitors API requires both api_key and application_key",
			conn.ID,
		))
	}

	// Idempotency: if a row already exists for (connection, target, preset)
	// return it instead of failing — the button is designed to be
	// "create or no-op", not "create or 409". This mirrors how the Slack
	// modal handles "channel already subscribed".
	var existing app.DatadogManagedMonitor
	if err := s.db.WithContext(ctx).
		Where(app.DatadogManagedMonitor{
			ConnectionID: conn.ID,
			TargetType:   req.TargetType,
			TargetID:     req.TargetID,
			Preset:       req.Preset,
		}).
		First(&existing).Error; err == nil {
		return &existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("check existing managed monitor: %w", err)
	}

	// Pick effective notify handles: per-request override wins, fall
	// back to the connection default. Empty is allowed — DD will just
	// not fan out, which the user might intentionally want if they
	// manage notification routing entirely inside DD.
	handles := req.NotifyHandles
	if len(handles) == 0 {
		handles = []string(conn.DefaultNotifyHandles)
	}

	ddReq, err := buildMonitorRequest(req.TargetType, req.TargetID, req.Preset, handles, req.DisplayName)
	if err != nil {
		return nil, stderr.NewInvalidRequest(fmt.Errorf("build monitor request: %w", err))
	}

	baseURL := ddclient.ResolveSiteURL(conn.Site)
	ddMonitor, err := s.ddClient.CreateMonitor(ctx, baseURL, conn.APIKey, conn.ApplicationKey, ddReq)
	if err != nil {
		// 401/403 from DD = stale credentials. Same auth-error
		// detection used by the lifecycle hook (mirrors the contract
		// from the slack hook).
		var apiErr *ddclient.APIError
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 401 || apiErr.StatusCode == 403) {
			return nil, stderr.NewInvalidRequest(fmt.Errorf(
				"datadog rejected the application key for connection %q; mark connection revoked and re-add keys",
				conn.ID,
			))
		}
		return nil, fmt.Errorf("create datadog monitor: %w", err)
	}

	row := app.DatadogManagedMonitor{
		ConnectionID:  conn.ID,
		OrgID:         orgID,
		TargetType:    req.TargetType,
		TargetID:      req.TargetID,
		Preset:        req.Preset,
		DDMonitorID:   ddMonitor.ID,
		NotifyHandles: cleanHandles(handles),
		CreatedByID:   acct.ID,
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		// Lost-race fallback: another goroutine inserted between our
		// existence check and Create. Return the existing row so the
		// caller sees idempotent behavior end-to-end.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			var racer app.DatadogManagedMonitor
			if rErr := s.db.WithContext(ctx).
				Where(app.DatadogManagedMonitor{
					ConnectionID: conn.ID,
					TargetType:   req.TargetType,
					TargetID:     req.TargetID,
					Preset:       req.Preset,
				}).
				First(&racer).Error; rErr == nil {
				return &racer, nil
			}
		}
		return nil, fmt.Errorf("create managed monitor row: %w", err)
	}
	return &row, nil
}
