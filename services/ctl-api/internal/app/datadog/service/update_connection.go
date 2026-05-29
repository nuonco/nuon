package service

import (
	"context"
	"errors"
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

// UpdateConnectionRequest mutates an existing DatadogConnection. Every
// field is optional; pass only what you want to change.
//
// Rotating APIKey re-runs the validation probe so a misconfigured rotation
// is caught the same way create-time is.
//
// Setting Status to "verified" while passing a new APIKey is the standard
// "re-enable a revoked connection after fixing creds" flow. Status alone
// can also flip between verified/revoked for manual ops without touching
// the keys.
type UpdateConnectionRequest struct {
	Name                 *string  `json:"name,omitempty" validate:"omitempty,min=1,max=128"`
	Site                 *string  `json:"site,omitempty"`
	APIKey               *string  `json:"api_key,omitempty" validate:"omitempty,min=10"`
	ApplicationKey       *string  `json:"application_key,omitempty"`
	Purpose              *string  `json:"purpose,omitempty" validate:"omitempty,oneof=internal customer"`
	Status               *string  `json:"status,omitempty" validate:"omitempty,oneof=verified revoked"`
	DefaultTags          []string `json:"default_tags,omitempty"`
	DefaultNotifyHandles []string `json:"default_notify_handles,omitempty"`
}

func (r *UpdateConnectionRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						UpdateDatadogConnection
// @Summary				Update a Datadog connection
// @Description			Mutates an existing DatadogConnection. Pass only the fields you want to change. Rotating `api_key` triggers a fresh validation probe against DD; a rejected key returns 400 so the rotation never silently leaves the connection in a broken state. The connection must belong to the calling org (ABAC enforced at the DB query level).
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id			path	string					true	"Org ID"
// @Param					connection_id	path	string					true	"Connection ID"
// @Param					req				body	UpdateConnectionRequest	true	"Input"
// @Success				200		{object}	app.DatadogConnection
// @Failure				400		{object}	stderr.ErrResponse
// @Failure				404		{object}	stderr.ErrResponse
// @Failure				500		{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/connections/{connection_id} [PATCH]
func (s *service) UpdateConnection(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	connectionID := ctx.Param("connection_id")

	req := UpdateConnectionRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	conn, err := s.updateConnection(ctx, org.ID, connectionID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update datadog connection: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, conn)
}

func (s *service) updateConnection(
	ctx context.Context,
	orgID, connectionID string,
	req *UpdateConnectionRequest,
) (*app.DatadogConnection, error) {
	var conn app.DatadogConnection
	if err := s.db.WithContext(ctx).
		Where(app.DatadogConnection{ID: connectionID, OrgID: orgID}).
		First(&conn).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("datadog connection %q not found in org %q", connectionID, orgID),
				Description: "Datadog connection not found",
			}
		}
		return nil, fmt.Errorf("lookup datadog connection: %w", err)
	}

	// Apply changes. Pointer fields drive "did the caller set it?";
	// slice fields are full replacements when present (empty slice
	// resets, nil leaves unchanged).
	if req.Name != nil {
		conn.Name = *req.Name
	}
	if req.Site != nil {
		conn.Site = *req.Site
	}
	if req.ApplicationKey != nil {
		conn.ApplicationKey = *req.ApplicationKey
	}
	if req.Purpose != nil {
		conn.Purpose = app.DatadogConnectionPurpose(*req.Purpose)
	}
	if req.Status != nil {
		conn.Status = app.DatadogConnectionStatus(*req.Status)
	}
	if req.DefaultTags != nil {
		conn.DefaultTags = req.DefaultTags
	}
	if req.DefaultNotifyHandles != nil {
		conn.DefaultNotifyHandles = req.DefaultNotifyHandles
	}

	// Re-validate before persisting an APIKey rotation. Run after we
	// applied the (possibly-new) Site so the probe hits the right
	// region.
	if req.APIKey != nil {
		conn.APIKey = *req.APIKey
		if err := s.validateAPIKey(ctx, conn.Site, conn.APIKey); err != nil {
			return nil, fmt.Errorf("datadog api key rejected: %w", err)
		}
	}

	if err := s.db.WithContext(ctx).Save(&conn).Error; err != nil {
		return nil, fmt.Errorf("save datadog connection: %w", err)
	}
	return &conn, nil
}
