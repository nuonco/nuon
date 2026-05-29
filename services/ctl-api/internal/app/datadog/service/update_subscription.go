package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// UpdateEventSubscriptionRequest mutates an existing routing rule. Every
// field is optional — pass only the fields you want to change.
//
// Match uses the same MatchSet sentinel trick as the Slack update path:
// the dashboard needs to distinguish "leave Match unchanged" (omit) from
// "make this row org-wide" (explicit null). Without UnmarshalJSON, the
// two cases would collapse.
type UpdateEventSubscriptionRequest struct {
	MatchSet          bool                      `json:"-"`
	Match             *labels.SubscriptionMatch `json:"match,omitempty" swaggertype:"object"`
	Interests         *interests.Interests      `json:"interests,omitempty" swaggertype:"object"`
	AdditionalTags    *[]string                 `json:"additional_tags,omitempty"`
	AlertTypeOverride *string                   `json:"alert_type_override,omitempty" validate:"omitempty,oneof=info warning error success"`
	PriorityOverride  *string                   `json:"priority_override,omitempty" validate:"omitempty,oneof=normal low"`
}

func (r *UpdateEventSubscriptionRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	type alias UpdateEventSubscriptionRequest
	if err := json.Unmarshal(data, (*alias)(r)); err != nil {
		return err
	}
	_, r.MatchSet = raw["match"]
	return nil
}

func (r *UpdateEventSubscriptionRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	if r.Interests != nil {
		if err := interests.Validate(*r.Interests); err != nil {
			return err
		}
	}
	if r.MatchSet && r.Match != nil {
		if err := r.Match.Validate(); err != nil {
			return fmt.Errorf("invalid match: %w", err)
		}
	}
	return nil
}

// @ID						UpdateDatadogEventSubscription
// @Summary				Update a Datadog event subscription
// @Description			Mutates a routing rule. Pass only the fields you want to change. Updating `match` may collide with the `(connection_id, match_canonical)` unique index — the API returns 409 with a clear description in that case so the dashboard can render the same toast it shows on a duplicate create.
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string							true	"Org ID"
// @Param					sub_id	path	string							true	"Subscription ID"
// @Param					req		body	UpdateEventSubscriptionRequest	true	"Input"
// @Success				200	{object}	app.DatadogEventSubscription
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/event-subscriptions/{sub_id} [PATCH]
func (s *service) UpdateEventSubscription(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	subID := ctx.Param("sub_id")
	if subID == "" {
		ctx.Error(stderr.NewInvalidRequest(errors.New("sub_id is required")))
		return
	}

	req := UpdateEventSubscriptionRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	sub, err := s.updateEventSubscription(ctx, org.ID, subID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, sub)
}

func (s *service) updateEventSubscription(
	ctx context.Context,
	orgID, subID string,
	req *UpdateEventSubscriptionRequest,
) (*app.DatadogEventSubscription, error) {
	var sub app.DatadogEventSubscription
	if err := s.db.WithContext(ctx).
		Where(app.DatadogEventSubscription{ID: subID, OrgID: orgID}).
		First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("datadog event subscription %q not found for org %q", subID, orgID),
				Description: "Datadog event subscription not found",
			}
		}
		return nil, fmt.Errorf("lookup datadog event subscription: %w", err)
	}

	if req.MatchSet {
		sub.Match = req.Match
	}
	if req.Interests != nil {
		sub.Interests = *req.Interests
	}
	if req.AdditionalTags != nil {
		sub.AdditionalTags = *req.AdditionalTags
	}
	if req.AlertTypeOverride != nil {
		sub.AlertTypeOverride = *req.AlertTypeOverride
	}
	if req.PriorityOverride != nil {
		sub.PriorityOverride = *req.PriorityOverride
	}

	if err := s.db.WithContext(ctx).Save(&sub).Error; err != nil {
		// The BeforeSave hook recomputes MatchCanonical for the
		// unique index, so a Match change can collapse onto an
		// existing row. Surface that as 409 — same shape the create
		// flow uses on duplicate inserts.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, stderr.ErrConflict{
				Err:         fmt.Errorf("datadog event subscription with this scope already exists: %w", err),
				Description: "Scope already subscribed on this connection",
			}
		}
		return nil, fmt.Errorf("update datadog event subscription: %w", err)
	}
	return &sub, nil
}
