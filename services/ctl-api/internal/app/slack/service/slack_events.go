package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
type slackInnerEvent struct {
	Type string `json:"type"`
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

// handleSlackEventCallback dispatches on the inner event type. We currently
// only act on lifecycle events that invalidate the workspace install:
// app_uninstalled and tokens_revoked. Everything else is acked with 200.
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
	default:
		s.l.Debug("slack events: ignoring unhandled inner event",
			zap.String("event_type", env.Event.Type), zap.String("team_id", env.TeamID))
		ctx.Status(http.StatusOK)
	}
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
