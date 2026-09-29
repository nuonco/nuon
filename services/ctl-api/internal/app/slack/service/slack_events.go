package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

type slackEventEnvelope struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge,omitempty"`
	TeamID    string          `json:"team_id,omitempty"`
	APIAppID  string          `json:"api_app_id,omitempty"`
	Event     slackInnerEvent `json:"event,omitempty"`
}

type slackInnerEvent struct {
	Type    string          `json:"type"`
	Channel json.RawMessage `json:"channel,omitempty"`
	User    string          `json:"user,omitempty"`
}

type slackChannelRef struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (e slackInnerEvent) channelIDFromEvent() string {
	ref := e.parseChannelRef()
	return ref.ID
}

func (e slackInnerEvent) parseChannelRef() slackChannelRef {
	if len(e.Channel) == 0 {
		return slackChannelRef{}
	}
	var asObj slackChannelRef
	if err := json.Unmarshal(e.Channel, &asObj); err == nil && asObj.ID != "" {
		return asObj
	}
	var asStr string
	if err := json.Unmarshal(e.Channel, &asStr); err == nil {
		return slackChannelRef{ID: asStr}
	}
	return slackChannelRef{}
}

// slackChallengeResponse is what Slack expects back during the URL
// verification handshake (sent once when wiring up the Events API
// subscription URL in the Slack app config).
type slackChallengeResponse struct {
	Challenge string `json:"challenge"`
}

// SlackEvents handles POSTs from Slack's Events API, which fires for
// app_uninstalled, tokens_revoked, and the initial url_verification handshake.
// Authenticated via the Slack signing-secret middleware on the route group;
// 200 OK on every handled event (Slack retries 4xx/5xx aggressively).
//
//	@ID						SlackEvents
//	@Summary				Slack Events API webhook
//	@Description			Receives lifecycle events from Slack: url_verification (handshake), app_uninstalled (workspace removed Nuon), tokens_revoked (bot token invalidated). Authenticated via Slack signing-secret middleware (X-Slack-Signature + X-Slack-Request-Timestamp); not via API key. Returns 200 even for unhandled event types so Slack does not retry.
//	@Tags					slack
//	@Accept					json
//	@Produce				json
//	@Param					body	body	object	true	"Slack event envelope"
//	@Success				200	{object}	slackChallengeResponse	"For url_verification: returns challenge. Otherwise empty body."
//	@Router					/slack/events [POST]
func (s *service) SlackEvents(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		s.l.Warn("slack events: read body failed", zap.Error(err))
		ctx.Status(http.StatusBadRequest)
		return
	}

	var env slackEventEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		s.l.Warn("slack events: decode body failed", zap.Error(err))
		ctx.Status(http.StatusOK)
		return
	}

	switch env.Type {
	case "url_verification":
		ctx.JSON(http.StatusOK, slackChallengeResponse{Challenge: env.Challenge})
		return
	case "event_callback":
		s.handleSlackEventCallback(ctx, env)
		return
	default:
		s.l.Debug("slack events: ignoring unhandled envelope type",
			zap.String("type", env.Type), zap.String("team_id", env.TeamID))
		ctx.Status(http.StatusOK)
		return
	}
}

func (s *service) handleSlackEventCallback(ctx *gin.Context, env slackEventEnvelope) {
	switch env.Event.Type {
	case "app_uninstalled", "tokens_revoked":
		if env.TeamID == "" {
			s.l.Warn("slack events: lifecycle event missing team_id",
				zap.String("event_type", env.Event.Type))
			ctx.Status(http.StatusOK)
			return
		}
		if err := s.markWorkspaceUninstalled(ctx, env.TeamID, env.Event.Type); err != nil {
			s.l.Error("slack events: mark uninstalled failed",
				zap.Error(err), zap.String("team_id", env.TeamID),
				zap.String("event_type", env.Event.Type))
			// why: Still 200 — Slack would retry forever otherwise. We've logged
			// for ops follow-up.
			ctx.Status(http.StatusOK)
			return
		}
		s.l.Info("slack events: workspace uninstalled",
			zap.String("team_id", env.TeamID),
			zap.String("event_type", env.Event.Type))
		ctx.Status(http.StatusOK)
	case "channel_rename":
		ref := env.Event.parseChannelRef()
		if env.TeamID == "" || ref.ID == "" {
			s.l.Warn("slack events: channel_rename missing team_id/channel",
				zap.String("team_id", env.TeamID))
			ctx.Status(http.StatusOK)
			return
		}
		if err := s.renameSubscriptionsForChannel(ctx, env.TeamID, ref.ID, ref.Name); err != nil {
			s.l.Error("slack events: rename channel subs failed",
				zap.Error(err),
				zap.String("team_id", env.TeamID),
				zap.String("channel_id", ref.ID))
			ctx.Status(http.StatusOK)
			return
		}
		s.l.Info("slack events: channel renamed",
			zap.String("team_id", env.TeamID),
			zap.String("channel_id", ref.ID),
			zap.String("new_name", ref.Name))
		ctx.Status(http.StatusOK)
	case "channel_archive", "channel_left":
		channelID := env.Event.channelIDFromEvent()
		if env.TeamID == "" || channelID == "" {
			s.l.Warn("slack events: channel event missing team_id/channel",
				zap.String("team_id", env.TeamID),
				zap.String("event_type", env.Event.Type))
			ctx.Status(http.StatusOK)
			return
		}
		if err := s.softDeleteSubscriptionsForChannel(ctx, env.TeamID, channelID); err != nil {
			s.l.Error("slack events: soft-delete channel subs failed",
				zap.Error(err),
				zap.String("team_id", env.TeamID),
				zap.String("channel_id", channelID),
				zap.String("event_type", env.Event.Type))
			ctx.Status(http.StatusOK)
			return
		}
		s.l.Info("slack events: channel subs cleaned up",
			zap.String("team_id", env.TeamID),
			zap.String("channel_id", channelID),
			zap.String("event_type", env.Event.Type))
		ctx.Status(http.StatusOK)
	case "member_joined_channel":
		if ctx.GetHeader("X-Slack-Retry-Num") != "" {
			ctx.Status(http.StatusOK)
			return
		}
		channelID := env.Event.channelIDFromEvent()
		if env.TeamID == "" || channelID == "" || env.Event.User == "" {
			s.l.Warn("slack events: member_joined_channel missing team_id/channel/user",
				zap.String("team_id", env.TeamID))
			ctx.Status(http.StatusOK)
			return
		}
		if err := s.welcomeChannelOnBotJoin(ctx, env.TeamID, channelID, env.Event.User); err != nil {
			s.l.Error("slack events: welcome on bot join failed",
				zap.Error(err),
				zap.String("team_id", env.TeamID),
				zap.String("channel_id", channelID))
			ctx.Status(http.StatusOK)
			return
		}
		ctx.Status(http.StatusOK)
	default:
		s.l.Debug("slack events: ignoring unhandled inner event",
			zap.String("event_type", env.Event.Type), zap.String("team_id", env.TeamID))
		ctx.Status(http.StatusOK)
	}
}

// why: renameSubscriptionsForChannel updates SlackChannelSubscription.ChannelName
// for every active sub that references the renamed channel. Scoped to the
// signed (team_id, channel_id) so cross-workspace bleed isn't possible.
// No-ops gracefully when the new name is empty (Slack should never send that
// but we don't want to clobber existing names with "").
func (s *service) renameSubscriptionsForChannel(ctx *gin.Context, teamID, channelID, newName string) error {
	if newName == "" {
		return nil
	}
	return s.db.WithContext(ctx).
		Model(&app.SlackChannelSubscription{}).
		Where(app.SlackChannelSubscription{TeamID: teamID, ChannelID: channelID}).
		Updates(map[string]any{"channel_name": newName}).Error
}

func (s *service) softDeleteSubscriptionsForChannel(ctx *gin.Context, teamID, channelID string) error {
	return s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{TeamID: teamID, ChannelID: channelID}).
		Delete(&app.SlackChannelSubscription{}).Error
}

const slackWelcomeTextFmt = ":wave: *Thanks for adding %s!*\n" +
	"I post deployment lifecycle events from your installs into the channels you choose.\n\n" +
	"Run `/nuon subscribe` to pick which org and events this channel should receive, or `/nuon help` to see everything I can do."

const defaultBotDisplayName = "Nuon"

func (s *service) welcomeChannelOnBotJoin(ctx *gin.Context, teamID, channelID, joinedUserID string) error {
	var install app.SlackInstallation
	res := s.db.WithContext(ctx).
		Where(app.SlackInstallation{TeamID: teamID, Status: app.SlackInstallationStatusActive}).
		First(&install)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("lookup installation: %w", res.Error)
	}

	if joinedUserID != install.BotUserID {
		return nil
	}

	botName := defaultBotDisplayName
	if at, err := s.slackClient.AuthTest(ctx, install.BotAccessToken); err != nil {
		s.l.Warn("slack welcome: auth.test failed, using default bot name",
			zap.String("team_id", teamID), zap.Error(err))
	} else if at.User != "" {
		botName = at.User
	}

	if _, err := s.slackClient.PostMessage(ctx, install.BotAccessToken, slackclient.PostMessageRequest{
		Channel: channelID,
		Text:    fmt.Sprintf(slackWelcomeTextFmt, botName),
	}); err != nil {
		return fmt.Errorf("post welcome message: %w", err)
	}
	return nil
}

func (s *service) markWorkspaceUninstalled(ctx *gin.Context, teamID, reason string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// why: 1. Flip installation status. If no row matches, treat as no-op.
		// Intentionally no Status filter: we want this update to be idempotent
		// for repeat app_uninstalled / tokens_revoked deliveries (Slack retries
		// aggressively). GORM's default scope already hides soft-deleted
		// tombstones, so we never overwrite a tombstoned re-install row.
		if err := tx.Model(&app.SlackInstallation{}).
			Where(app.SlackInstallation{TeamID: teamID}).
			Updates(map[string]any{
				"status": app.SlackInstallationStatusUninstalled,
			}).Error; err != nil {
			return fmt.Errorf("update installation status: %w", err)
		}

		if err := tx.Model(&app.SlackOrgLink{}).
			Where(app.SlackOrgLink{TeamID: teamID, Status: app.SlackOrgLinkStatusVerified}).
			Updates(map[string]any{
				"status": app.SlackOrgLinkStatusRevoked,
			}).Error; err != nil {
			return fmt.Errorf("revoke org links: %w", err)
		}

		if err := tx.Where(app.SlackChannelSubscription{TeamID: teamID}).
			Delete(&app.SlackChannelSubscription{}).Error; err != nil {
			return fmt.Errorf("soft-delete channel subscriptions: %w", err)
		}
		return nil
	})
}
