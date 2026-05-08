package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// slackInteractionPayload is the outer envelope Slack POSTs to the
// interactivity endpoint. Slack sends a single form parameter `payload`
// whose value is JSON; we parse the subset of fields the dispatch + handler
// implementations actually consume.
//
// Reference: https://api.slack.com/interactivity/handling
type slackInteractionPayload struct {
	Type      string `json:"type"`
	TriggerID string `json:"trigger_id,omitempty"`
	Token     string `json:"token,omitempty"`

	// Team identifies the workspace the action originated in. We always
	// re-derive trust from this value (verified by the signing middleware
	// upstream) and never trust IDs the user picked in the modal alone.
	Team struct {
		ID string `json:"id,omitempty"`
	} `json:"team,omitempty"`

	User struct {
		ID string `json:"id,omitempty"`
	} `json:"user,omitempty"`

	// View carries the modal context for view_submission, view_closed, and
	// block_actions inside a modal. State.Values is keyed
	// block_id -> action_id -> raw element-state map (Slack's wire shape).
	View struct {
		ID              string `json:"id,omitempty"`
		Hash            string `json:"hash,omitempty"`
		CallbackID      string `json:"callback_id,omitempty"`
		PrivateMetadata string `json:"private_metadata,omitempty"`
		State           struct {
			Values map[string]map[string]map[string]any `json:"values,omitempty"`
		} `json:"state,omitempty"`
	} `json:"view,omitempty"`

	// Actions is set on block_actions payloads (button clicks etc.).
	Actions []struct {
		ActionID string `json:"action_id,omitempty"`
		BlockID  string `json:"block_id,omitempty"`
		Value    string `json:"value,omitempty"`
	} `json:"actions,omitempty"`

	// ActionID / BlockID / Value are top-level on block_suggestion payloads
	// (Slack's external_select dynamic-options handshake). Value is the
	// query string the user has typed so far.
	ActionID string `json:"action_id,omitempty"`
	BlockID  string `json:"block_id,omitempty"`
	Value    string `json:"value,omitempty"`
}

// SlackInteractions is the single request_url Slack invokes for every
// interactive surface (modals, buttons, select menus, dynamic options).
// It parses the envelope and dispatches on payload.Type:
//
//   - view_submission → subscribe modal submit (other callback_ids logged
//     and acked).
//   - block_actions → subscribe modal scope/notif radios (re-render via
//     views.update) and the unsubscribe modal's Remove buttons
//     (soft-delete + re-render).
//   - block_suggestion → subscribe modal install picker's external_select
//     dynamic options.
//   - view_closed / shortcut / message_action / unknown → acked with 200.
//
// Authenticated via the Slack signing-secret middleware on the route group.
// Returns 200 OK on every parseable envelope; Slack retries 4xx/5xx
// aggressively so we keep the negative-path noise to a minimum.
//
//	@ID						SlackInteractions
//	@Summary				Slack interactivity & shortcuts request URL
//	@Description			Receives interactive payloads (view_submission, block_actions, block_suggestion, shortcut). Authenticated via Slack signing-secret middleware (X-Slack-Signature + X-Slack-Request-Timestamp); not via API key. Dispatches subscribe/unsubscribe modal submissions, block_actions (scope/notif radios, Remove buttons), and the install picker's external_select block_suggestion handshake. Returns 200 on every parseable envelope so Slack does not retry; unhandled payload types are logged and acked.
//	@Tags					slack
//	@Accept					x-www-form-urlencoded
//	@Produce				json
//	@Param					payload	formData	string	true	"JSON-encoded interaction payload"
//	@Success				200
//	@Router					/slack/interactions [POST]
func (s *service) SlackInteractions(ctx *gin.Context) {
	rawPayload := ctx.PostForm("payload")
	if rawPayload == "" {
		s.l.Warn("slack interactions: missing payload form field")
		// Return 200 anyway — Slack treats non-2xx as a retry-trigger and
		// we have nothing actionable to add by failing here.
		ctx.Status(http.StatusOK)
		return
	}

	var payload slackInteractionPayload
	if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
		s.l.Warn("slack interactions: unmarshal payload failed",
			zap.Error(err),
			zap.String("payload_excerpt", excerpt(rawPayload, 256)))
		ctx.Status(http.StatusOK)
		return
	}

	logger := s.l.With(
		zap.String("interaction_type", payload.Type),
		zap.String("team_id", payload.Team.ID),
		zap.String("user_id", payload.User.ID),
		zap.String("view_callback_id", payload.View.CallbackID),
	)

	switch payload.Type {
	case "view_submission":
		// Modal submit: dispatch on callback_id. Subscribe modal closes
		// on success (empty 200) or returns response_action=errors to
		// keep the modal open with inline errors. Unsubscribe modal
		// lands later.
		switch payload.View.CallbackID {
		case subscribeModalCallbackID:
			body := s.handleSubscribeModalSubmission(ctx, payload)
			ctx.JSON(http.StatusOK, body)
			return
		default:
			logger.Debug("slack interactions: view_submission unknown callback_id")
		}
	case "view_closed":
		// Modal dismissed: nothing to do.
		logger.Debug("slack interactions: view_closed (no-op)")
	case "block_actions":
		// Block-level interactions inside a published view. Subscribe
		// modal scope/notif radios re-render via views.update; the
		// unsubscribe modal's Remove buttons soft-delete + re-render.
		switch payload.View.CallbackID {
		case subscribeModalCallbackID:
			s.handleSubscribeModalBlockActions(ctx, payload)
			ctx.Status(http.StatusOK)
			return
		case unsubscribeModalCallbackID:
			s.handleUnsubscribeModalBlockActions(ctx, payload)
			ctx.Status(http.StatusOK)
			return
		}
		logger.Debug("slack interactions: block_actions unhandled view")
	case "block_suggestion":
		// Dynamic options for external_select. The subscribe modal
		// drives four pickers (apps, then installs / components /
		// actions), each keyed off a distinct action_id; the modal's
		// resource-type select decides which entity picker is rendered,
		// and the app picker only appears for component / action kinds.
		// Returning an empty options list signals "no matches" to Slack.
		if payload.View.CallbackID == subscribeModalCallbackID {
			switch payload.ActionID {
			case subscribeAppActionID:
				body := s.handleSubscribeModalAppsBlockSuggestion(ctx, payload)
				ctx.JSON(http.StatusOK, body)
				return
			case subscribeEntitiesActionIDInstalls:
				body := s.handleSubscribeModalBlockSuggestion(ctx, payload)
				ctx.JSON(http.StatusOK, body)
				return
			case subscribeEntitiesActionIDComponents:
				body := s.handleSubscribeModalComponentsBlockSuggestion(ctx, payload)
				ctx.JSON(http.StatusOK, body)
				return
			case subscribeEntitiesActionIDActions:
				body := s.handleSubscribeModalActionsBlockSuggestion(ctx, payload)
				ctx.JSON(http.StatusOK, body)
				return
			}
		}
		ctx.JSON(http.StatusOK, gin.H{"options": []any{}})
		return
	case "shortcut", "message_action":
		// Global / message shortcuts — out of scope for this PR.
		logger.Debug("slack interactions: shortcut (out of scope)")
	default:
		logger.Debug("slack interactions: unhandled payload type")
	}

	ctx.Status(http.StatusOK)
}

// excerpt returns the first n bytes of s for safe inclusion in log lines.
// Used to surface a hint about malformed payloads without dumping whole
// request bodies.
func excerpt(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return fmt.Sprintf("%s…(%d more bytes)", s[:n], len(s)-n)
}
