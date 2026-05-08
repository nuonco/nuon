package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

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

// UpdateChannelSubscriptionRequest mutates an existing per-channel routing
// rule from the dashboard. Every field is optional except by-ID identity in
// the URL — pass only the fields you want to change. Updating Match may
// cause a unique-index conflict if a row with the new canonical predicate
// already exists for the same (team, channel, link); the handler surfaces
// that as a 409 with a clear description so the dashboard can render a
// toast that mirrors the create flow's "Channel already subscribed".
type UpdateChannelSubscriptionRequest struct {
	ChannelID   *string `json:"channel_id,omitempty"`
	ChannelName *string `json:"channel_name,omitempty"`
	// Match uses a sentinel "set" via a wrapper so callers can express
	// "leave unchanged" (omit) vs "make org-wide" (explicitly null).
	// Slack-side rows treat nil Match as the org-wide subscription, so
	// without a sentinel the handler couldn't tell which case the
	// caller meant. updateMatch below keys off MatchSet.
	MatchSet  bool                      `json:"-"`
	Match     *labels.SubscriptionMatch `json:"match,omitempty" swaggertype:"object"`
	Interests *interests.Interests      `json:"interests,omitempty" swaggertype:"object"`
}

// UnmarshalJSON sets MatchSet whenever the body carries a `match` key,
// even if its value is JSON null. Without this, "match: null" and an
// omitted match are indistinguishable from the request struct.
func (r *UpdateChannelSubscriptionRequest) UnmarshalJSON(data []byte) error {
	// Two-pass decode: first into a raw map to detect key presence,
	// then into the struct shape via an alias to avoid recursion.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	type alias UpdateChannelSubscriptionRequest
	if err := json.Unmarshal(data, (*alias)(r)); err != nil {
		return err
	}
	_, r.MatchSet = raw["match"]
	return nil
}

func (r *UpdateChannelSubscriptionRequest) Validate(v *validator.Validate) error {
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

// @ID						UpdateSlackChannelSubscription
// @Summary				Update a Slack channel subscription
// @Description			Mutates a per-channel routing rule. Pass only the fields you want to change. Updating `match` may collide with the `(team_id, channel_id, org_link_id, match_canonical)` unique index — the API returns 409 with a clear description in that case so the dashboard can render the same toast it shows on a duplicate create. The subscription must belong to the calling org (ABAC enforced at the DB query level).
// @Tags					slack
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string								true	"Org ID"
// @Param					sub_id	path	string								true	"Subscription ID"
// @Param					req		body	UpdateChannelSubscriptionRequest	true	"Input"
// @Success				200	{object}	app.SlackChannelSubscription
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/slack/channel-subscriptions/{sub_id} [PATCH]
func (s *service) UpdateChannelSubscription(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	subID := strings.TrimSpace(ctx.Param("sub_id"))
	if subID == "" {
		ctx.Error(stderr.NewInvalidRequest(errors.New("sub_id is required")))
		return
	}

	req := UpdateChannelSubscriptionRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	sub, err := s.updateChannelSubscription(ctx, org.ID, subID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, sub)
}

// updateChannelSubscription trust-binds the row to the caller's org and
// applies the requested patches. The model's BeforeSave hook keeps
// MatchCanonical in lockstep with Match for both inserts and updates, so
// changing Match here recomputes the index column correctly.
func (s *service) updateChannelSubscription(
	ctx context.Context,
	orgID, subID string,
	req *UpdateChannelSubscriptionRequest,
) (*app.SlackChannelSubscription, error) {
	var sub app.SlackChannelSubscription
	if err := s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{ID: subID, OrgID: orgID}).
		First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("slack channel subscription %q not found for org %q", subID, orgID),
				Description: "Slack channel subscription not found",
			}
		}
		return nil, fmt.Errorf("lookup slack channel subscription: %w", err)
	}

	// Apply patches to the loaded row, then Save — using a map of updates
	// would skip the BeforeSave hook that recomputes MatchCanonical for
	// the unique index.
	if req.ChannelID != nil {
		sub.ChannelID = strings.TrimSpace(*req.ChannelID)
	}
	if req.ChannelName != nil {
		sub.ChannelName = *req.ChannelName
	}
	if req.MatchSet {
		sub.Match = req.Match
	}
	if req.Interests != nil {
		sub.Interests = *req.Interests
	}

	if err := s.db.WithContext(ctx).Save(&sub).Error; err != nil {
		// Postgres unique_violation = 23505. The unique index on
		// (team_id, channel_id, org_link_id, match_canonical, deleted_at)
		// fires when an edit collapses the new Match canonical onto an
		// existing row. Surface that as a 409 with a description that
		// mirrors the create flow's "Channel already subscribed" toast.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, stderr.ErrConflict{
				Err:         fmt.Errorf("slack channel subscription with this scope already exists: %w", err),
				Description: "Scope already subscribed to this channel",
			}
		}
		return nil, fmt.Errorf("update slack channel subscription: %w", err)
	}
	return &sub, nil
}
