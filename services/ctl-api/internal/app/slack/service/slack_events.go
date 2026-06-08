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

// slackEventEnvelope is the outer JSON wrapper Slack sends to the Events API.
// Reference: https://api.slack.com/apis/connections/events-api
type slackEventEnvelope struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge,omitempty"`
	TeamID    string          `json:"team_id,omitempty"`
	APIAppID  string          `json:"api_app_id,omitempty"`
	Event     slackInnerEvent `json:"event,omitempty"`
}

// slackInnerEvent is the inner `event` object on event_callback envelopes.
// We deliberately read only the fields we act on; Slack ships many more.
//
// Channel is overloaded by Slack: most channel.* events ship it as a string
// (just the ID), but channel_rename ships an object {id, name}. We capture
// it as RawMessage and decode lazily into the right shape per event type.
type slackInnerEvent struct {
	Type    string          `json:"type"`
	Channel json.RawMessage `json:"channel,omitempty"`
	// User is the Slack user id of the actor on member_joined_channel events.
	// We compare it against the install's BotUserID to detect when our own
	// bot was added to a channel (vs. a human teammate joining).
	User string `json:"user,omitempty"`
}

// slackChannelRef is the object shape Slack uses for channel_rename's
// `channel` field. Other channel events ship `channel` as a bare string id.
type slackChannelRef struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// channelIDFromEvent extracts the channel id from either Slack shape:
//
//   - channel_rename: object {id, name, ...} -> id is the channel id
//   - channel_archive / channel_left: bare string "C123" -> the channel id
//
// Returns empty string when the field is absent or malformed; callers treat
// that as a no-op (we 200 + log).
func (e slackInnerEvent) channelIDFromEvent() string {
	ref := e.parseChannelRef()
	return ref.ID
}

// parseChannelRef decodes the polymorphic channel field. Returns zero-value
// slackChannelRef when missing / malformed.
func (e slackInnerEvent) parseChannelRef() slackChannelRef {
	if len(e.Channel) == 0 {
		return slackChannelRef{}
	}
	// Try object shape first (channel_rename).
	var asObj slackChannelRef
	if err := json.Unmarshal(e.Channel, &asObj); err == nil && asObj.ID != "" {
		return asObj
	}
	// Fall back to bare string.
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
		// Return 200 so Slack does not retry malformed requests forever.
		ctx.Status(http.StatusOK)
		return
	}

	switch env.Type {
	case "url_verification":
		// One-time handshake sent when the Events Request URL is saved in
		// the Slack app config. The signing middleware still verifies this
		// request — we just echo the challenge back.
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

// handleSlackEventCallback dispatches on the inner event type. We act on:
//
//   - app_uninstalled / tokens_revoked: workspace install invalidated;
//     flip Status, revoke org-links, soft-delete subs (see
//     markWorkspaceUninstalled).
//   - channel_rename: keep SlackChannelSubscription.ChannelName fresh so
//     audit logs / pickers stay readable.
//   - channel_archive / channel_left: we can't post into the channel
//     anymore, so soft-delete every active sub for that channel.
//   - member_joined_channel: when the joining member is our own bot, post a
//     welcome message into the channel (see welcomeChannelOnBotJoin).
//
// Everything else is acked with 200 (Slack retries 4xx/5xx aggressively).
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
			// Still 200 — Slack would retry forever otherwise. We've logged
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
		// Slack fires this for every member that joins a channel the bot is
		// in; we only act when the joining member is our bot (i.e. the bot
		// was just added to the channel).
		//
		// Slack uses at-least-once delivery: a join we already processed can
		// be re-POSTed with X-Slack-Retry-Num set. We skip retries so a
		// single membership yields a single welcome, while a genuine re-add
		// (always a fresh delivery, no retry header) is welcomed again —
		// i.e. once-per-membership, no persistent state required.
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

// renameSubscriptionsForChannel updates SlackChannelSubscription.ChannelName
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

// softDeleteSubscriptionsForChannel soft-deletes every active sub targeting
// (team_id, channel_id). Used on channel_archive and channel_left — the bot
// can't post into the channel anymore, so silently dropping the rows is
// safer than leaving them and accumulating chat.postMessage failures.
//
// Org-scoped and install-scoped rows alike are removed; if the channel comes
// back (unarchive + re-add) the user resubscribes via /nuon subscribe.
func (s *service) softDeleteSubscriptionsForChannel(ctx *gin.Context, teamID, channelID string) error {
	return s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{TeamID: teamID, ChannelID: channelID}).
		Delete(&app.SlackChannelSubscription{}).Error
}

// slackWelcomeTextFmt is posted into a channel the first time our bot is added
// to it. Mirrors the slash-command help phrasing so the onboarding is
// consistent. Kept as plain text (vs. Block Kit) for the same reasons as
// slashHelpText — the v1 surface is intentionally thin. The %s is the bot's
// own display name (resolved per-installation, see welcomeChannelOnBotJoin) so
// the greeting matches whatever the app is actually named in this workspace.
const slackWelcomeTextFmt = ":wave: *Thanks for adding %s!*\n" +
	"I post deployment lifecycle events from your installs into the channels you choose.\n\n" +
	"Run `/nuon subscribe` to pick which org and events this channel should receive, or `/nuon help` to see everything I can do."

// defaultBotDisplayName is the greeting fallback when we can't resolve the
// bot's actual display name from Slack (e.g. auth.test fails). Never block the
// welcome on a name lookup.
const defaultBotDisplayName = "Nuon"

// welcomeChannelOnBotJoin posts the welcome message when our bot is the member
// that just joined channelID. No-ops (returns nil) when there's no active
// installation for the workspace or when the joining member is someone other
// than our bot — both are expected, non-error cases. Retry/dedup is handled by
// the caller (member_joined_channel skips Slack re-deliveries), so this always
// posts on a genuine join.
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

	// Only greet when the joining member is our bot — Slack fires
	// member_joined_channel for every member, including human teammates.
	if joinedUserID != install.BotUserID {
		return nil
	}

	// Resolve the bot's own display name so the greeting matches whatever the
	// app is named in this workspace (e.g. "nuon-stage"). auth.test needs no
	// extra scope and returns the bot user's handle in User. Best-effort: fall
	// back to the brand name rather than skip the welcome on a lookup failure.
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

// markWorkspaceUninstalled flips the installation Status to uninstalled and
// revokes any verified org-links + active subscriptions for this workspace.
// We do NOT soft-delete the installation row itself; we keep it so audit /
// operator queries can still see history. A subsequent re-install via the
// OAuth callback will flip Status back to active and refresh the token.
func (s *service) markWorkspaceUninstalled(ctx *gin.Context, teamID, reason string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Flip installation status. If no row matches, treat as no-op.
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

		// 2. Revoke any verified org-links so subsequent message routing
		// fails the trust check (a, b, c invariant in the model docs).
		if err := tx.Model(&app.SlackOrgLink{}).
			Where(app.SlackOrgLink{TeamID: teamID, Status: app.SlackOrgLinkStatusVerified}).
			Updates(map[string]any{
				"status": app.SlackOrgLinkStatusRevoked,
			}).Error; err != nil {
			return fmt.Errorf("revoke org links: %w", err)
		}

		// 3. Soft-delete every channel sub for this workspace. The PG
		// CASCADE on slack_channel_subscriptions.org_link_id only fires on
		// hard deletes, so we mirror it for the soft-delete path.
		if err := tx.Where(app.SlackChannelSubscription{TeamID: teamID}).
			Delete(&app.SlackChannelSubscription{}).Error; err != nil {
			return fmt.Errorf("soft-delete channel subscriptions: %w", err)
		}
		return nil
	})
}
