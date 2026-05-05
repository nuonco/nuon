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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// CreateChannelSubscriptionRequest is the body for creating a per-channel
// routing rule from the dashboard. CreatedBySlackUserID stays nil for
// dashboard-originated subs (the CHECK constraint requires at least one of
// the two creator fields; the account ID we set below satisfies it).
//
// Interests is optional. When omitted (zero value), the handler stamps in
// the AllEvents() default so the new sub forwards every supported lifecycle
// event until the user opts into a per-resource configuration.
type CreateChannelSubscriptionRequest struct {
	OrgLinkID   string              `json:"org_link_id" validate:"required"`
	ChannelID   string              `json:"channel_id" validate:"required"`
	ChannelName string              `json:"channel_name"`
	Interests   interests.Interests `json:"interests" swaggertype:"object"`
}

func (r *CreateChannelSubscriptionRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return interests.Validate(r.Interests)
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

func (s *service) createChannelSubscription(
	ctx context.Context,
	acct *app.Account,
	orgID string,
	req *CreateChannelSubscriptionRequest,
) (*app.SlackChannelSubscription, error) {
	// ABAC: OrgLinkID must resolve to a verified link in the caller's org.
	// Filtering by org id + status in the same WHERE means a caller cannot
	// create subscriptions against another org's link or against a revoked
	// one — the row simply isn't found.
	var link app.SlackOrgLink
	res := s.db.WithContext(ctx).
		Where(app.SlackOrgLink{
			ID:     req.OrgLinkID,
			OrgID:  orgID,
			Status: app.SlackOrgLinkStatusVerified,
		}).
		First(&link)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, stderr.ErrNotFound{Err: fmt.Errorf("slack org link %q not found", req.OrgLinkID)}
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// Pre-check for an existing live subscription on
	// (org_link_id, team_id, channel_id). Mirror createOrgLink — the unique
	// index would catch this at insert time but surfaces as a generic 500.
	var existing app.SlackChannelSubscription
	dupRes := s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{
			OrgLinkID: link.ID,
			TeamID:    link.TeamID,
			ChannelID: req.ChannelID,
		}).
		First(&existing)
	if dupRes.Error == nil {
		return nil, stderr.ErrConflict{
			Err:         fmt.Errorf("slack channel subscription already exists for channel %q on link %q", req.ChannelID, link.ID),
			Description: "this channel is already subscribed for this workspace",
		}
	}
	if !errors.Is(dupRes.Error, gorm.ErrRecordNotFound) {
		return nil, dupRes.Error
	}

	// Set the calling account on the context so SlackChannelSubscription's
	// BeforeCreate hook can resolve CreatedByID.
	ctx = cctx.SetAccountContext(ctx, acct)

	// Default new dashboard subs to AllEvents=true so users opt-in per
	// resource only when they actively configure the picker. Mirrors the
	// webhook default path.
	subInterests := req.Interests
	if subInterests.IsZero() {
		subInterests = interests.AllEvents()
	}

	acctID := acct.ID
	sub := &app.SlackChannelSubscription{
		OrgLinkID:          link.ID,
		TeamID:             link.TeamID,
		ChannelID:          req.ChannelID,
		ChannelName:        req.ChannelName,
		OrgID:              orgID,
		Interests:          subInterests,
		CreatedByAccountID: &acctID,
	}
	if err := s.db.WithContext(ctx).Create(sub).Error; err != nil {
		return nil, err
	}
	return sub, nil
}
