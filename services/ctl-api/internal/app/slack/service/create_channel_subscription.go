package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// CreateChannelSubscriptionRequest is the body for creating a per-channel
// routing rule from the dashboard. The Match field uses the same shape the
// slash-command modal builds — nil for org-wide, or any combination of the
// three TargetMatch slots (Installs / Components / Actions). Interests is
// optional; the handler defaults to AllEvents=true when omitted so the
// dashboard's minimal create-flow doesn't have to ship a full per-resource
// config to get a working sub.
type CreateChannelSubscriptionRequest struct {
	OrgLinkID   string                    `json:"org_link_id" validate:"required"`
	ChannelID   string                    `json:"channel_id" validate:"required"`
	ChannelName string                    `json:"channel_name"`
	Match       *labels.SubscriptionMatch `json:"match,omitempty" swaggertype:"object"`
	Interests   *interests.Interests      `json:"interests,omitempty" swaggertype:"object"`
}

func (r *CreateChannelSubscriptionRequest) Validate(v *validator.Validate) error {
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

// @ID						CreateSlackChannelSubscription
// @Summary				Create a Slack channel subscription
// @Description			Subscribes a Slack channel to events for the current org. The org_link_id must resolve to a verified SlackOrgLink belonging to the calling org; this is enforced at the DB query level (ABAC).
// @Tags					slack
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string								true	"Org ID"
// @Param					req		body	CreateChannelSubscriptionRequest	true	"Input"
// @Success				201	{object}	app.SlackChannelSubscription
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/slack/channel-subscriptions [POST]
func (s *service) CreateChannelSubscription(ctx *gin.Context) {
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

	req := CreateChannelSubscriptionRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	sub, err := s.createChannelSubscription(ctx, acct, org.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create slack channel subscription: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, sub)
}

// createChannelSubscription is the dashboard-side creator. It mirrors the
// slash-command modal's upsertModalSubscription path: trust-bind the
// org_link to the calling org, then insert a SlackChannelSubscription
// keyed on (team, channel, link, match_canonical). Re-creating with an
// identical Match upserts in place rather than 23505-ing — same semantics
// the modal relies on.
func (s *service) createChannelSubscription(
	ctx context.Context,
	acct *app.Account,
	orgID string,
	req *CreateChannelSubscriptionRequest,
) (*app.SlackChannelSubscription, error) {
	// Trust-bind: the link must be verified AND belong to the caller's org.
	// Re-derive both invariants from the DB rather than trusting the request
	// — the org_link_id is user-supplied.
	var link app.SlackOrgLink
	if err := s.db.WithContext(ctx).
		Where(app.SlackOrgLink{
			ID:     req.OrgLinkID,
			OrgID:  orgID,
			Status: app.SlackOrgLinkStatusVerified,
		}).
		First(&link).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("verified slack org link %q not found for org %q", req.OrgLinkID, orgID),
				Description: "Slack org link not found",
			}
		}
		return nil, fmt.Errorf("lookup slack org link: %w", err)
	}

	// Default Interests to AllEvents=true so a bare {org_link_id, channel_id}
	// request lands a working subscription. The modal already does the
	// same default; mirroring it here keeps the two creators consistent.
	in := interests.Interests{AllEvents: true}
	if req.Interests != nil {
		in = *req.Interests
	}

	sub := app.SlackChannelSubscription{
		OrgLinkID:          link.ID,
		OrgID:              link.OrgID,
		TeamID:             link.TeamID,
		ChannelID:          req.ChannelID,
		ChannelName:        req.ChannelName,
		Match:              req.Match,
		Interests:          in,
		CreatedByAccountID: &acct.ID,
		CreatedByID:        acct.ID,
	}

	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "team_id"},
			{Name: "channel_id"},
			{Name: "org_link_id"},
			{Name: "match_canonical"},
			{Name: "deleted_at"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"channel_name",
			"interests",
			"updated_at",
			"created_by_account_id",
		}),
	}).Create(&sub).Error; err != nil {
		return nil, fmt.Errorf("create slack channel subscription: %w", err)
	}
	return &sub, nil
}
