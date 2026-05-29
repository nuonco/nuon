package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// CreateEventSubscriptionRequest mirrors the Slack channel subscription
// create shape one-to-one (Match + Interests) since the routing predicate
// is identical. AdditionalTags / overrides are DD-specific.
//
// Match nil → org-wide subscription on this connection.
// Interests nil → defaults to AllEvents=true so a bare {connection_id}
// request lands a working subscription; matches the Slack create flow's
// behavior so the two surfaces stay symmetric.
type CreateEventSubscriptionRequest struct {
	ConnectionID      string                    `json:"connection_id" validate:"required"`
	Match             *labels.SubscriptionMatch `json:"match,omitempty" swaggertype:"object"`
	Interests         *interests.Interests      `json:"interests,omitempty" swaggertype:"object"`
	AdditionalTags    []string                  `json:"additional_tags,omitempty"`
	AlertTypeOverride string                    `json:"alert_type_override,omitempty" validate:"omitempty,oneof=info warning error success"`
	PriorityOverride  string                    `json:"priority_override,omitempty" validate:"omitempty,oneof=normal low"`
}

func (r *CreateEventSubscriptionRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	if r.Interests != nil {
		if err := interests.Validate(*r.Interests); err != nil {
			return err
		}
	}
	if r.Match != nil {
		if err := r.Match.Validate(); err != nil {
			return fmt.Errorf("invalid match: %w", err)
		}
	}
	return nil
}

// @ID						CreateDatadogEventSubscription
// @Summary				Create a Datadog event subscription
// @Description			Subscribes a Datadog connection's event stream to events for the current org. The connection must belong to the calling org (ABAC verified before write).
// @Tags					datadog
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string							true	"Org ID"
// @Param					req		body	CreateEventSubscriptionRequest	true	"Input"
// @Success				201	{object}	app.DatadogEventSubscription
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/datadog/event-subscriptions [POST]
func (s *service) CreateEventSubscription(ctx *gin.Context) {
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

	req := CreateEventSubscriptionRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	sub, err := s.createEventSubscription(ctx, acct, org.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create datadog event subscription: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, sub)
}

// createEventSubscription confirms the parent connection belongs to the
// caller's org, then upserts the row. The unique index on
// (connection_id, match_canonical, deleted_at) makes re-creating with an
// identical Match an idempotent update of Interests/tags/overrides — same
// semantics the Slack create flow relies on.
func (s *service) createEventSubscription(
	ctx context.Context,
	acct *app.Account,
	orgID string,
	req *CreateEventSubscriptionRequest,
) (*app.DatadogEventSubscription, error) {
	if err := s.orgConnectionExists(ctx, orgID, req.ConnectionID); err != nil {
		return nil, err
	}

	in := interests.Interests{AllEvents: true}
	if req.Interests != nil {
		in = *req.Interests
	}

	sub := app.DatadogEventSubscription{
		ConnectionID:      req.ConnectionID,
		OrgID:             orgID,
		Match:             req.Match,
		Interests:         in,
		AdditionalTags:    req.AdditionalTags,
		AlertTypeOverride: req.AlertTypeOverride,
		PriorityOverride:  req.PriorityOverride,
		CreatedByID:       acct.ID,
	}

	if err := s.db.WithContext(ctx).Create(&sub).Error; err != nil {
		return nil, fmt.Errorf("create datadog event subscription: %w", err)
	}
	return &sub, nil
}
