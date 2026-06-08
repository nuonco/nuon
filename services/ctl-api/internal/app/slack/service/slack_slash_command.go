package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// slashResponseTypeEphemeral is the response_type Slack honors for
// "only the invoking user can see this message" replies.
const slashResponseTypeEphemeral = "ephemeral"

// defaultSlashCommand is the fallback command name used when Slack's payload
// is missing the `command` field (malformed / replayed requests). Real slash
// POSTs always carry the actual invoked command.
const defaultSlashCommand = "/nuon"

// slashHelpText renders the canonical help shown for `<command> help` and
// unknown subcommands. command is the actual invoked slash command from the
// Slack payload (e.g. "/nuon", or a workspace-custom name like
// "/byoc-retool-dev"), so the examples match exactly what the user types.
// Kept as a single string (vs. block-kit) since the slash command surface is
// intentionally thin in v1; richer affordances live in the dashboard.
func slashHelpText(command string) string {
	return "*Nuon Slack commands*\n" +
		"`" + command + " subscribe`" + " — subscribe this channel to Nuon events (opens a dialog)\n" +
		"`" + command + " unsubscribe`" + " — remove this channel's subscription\n" +
		"`" + command + " status`" + " — show this workspace's installation, linked orgs, and this channel's subscription\n" +
		"`" + command + " help`" + " — show this message"
}

// slashResponse is the JSON envelope Slack expects from a slash command POST.
type slashResponse struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

// SlackSlashCommand handles POSTs from Slack for the /nuon slash command. The
// request is application/x-www-form-urlencoded, signed by Slack (verified by
// signing.Middleware on the route group), and ephemeral by default — we never
// echo into the channel without explicit user intent.
//
//	@ID						SlackSlashCommand
//	@Summary				Slack /nuon slash command webhook
//	@Description			Slack invokes this endpoint when a user runs `/nuon <subcommand>` in any channel of an installed workspace. Authenticated via the Slack signing-secret middleware (X-Slack-Signature + X-Slack-Request-Timestamp); not via API key. Subcommands: subscribe, unsubscribe, status, help. Responses are ephemeral.
//	@Tags					slack
//	@Accept					x-www-form-urlencoded
//	@Produce				json
//	@Param					team_id			formData	string	true	"Slack team (workspace) ID"
//	@Param					channel_id		formData	string	true	"Channel ID the command was invoked in"
//	@Param					channel_name	formData	string	false	"Channel name"
//	@Param					user_id			formData	string	true	"Slack user ID who invoked the command"
//	@Param					command			formData	string	true	"The slash command itself (e.g. /nuon)"
//	@Param					text			formData	string	false	"Subcommand text"
//	@Success				200	{object}	slashResponse
//	@Router					/slack/commands/nuon [POST]
func (s *service) SlackSlashCommand(ctx *gin.Context) {
	teamID := ctx.PostForm("team_id")
	channelID := ctx.PostForm("channel_id")
	channelName := ctx.PostForm("channel_name")
	userID := ctx.PostForm("user_id")
	triggerID := ctx.PostForm("trigger_id")
	text := strings.TrimSpace(ctx.PostForm("text"))

	// command is the actual invoked slash command (e.g. "/nuon" or a
	// workspace-custom name); we echo it back in help so examples match what
	// the user types. Falls back to the canonical name if Slack omits it.
	command := strings.TrimSpace(ctx.PostForm("command"))
	if command == "" {
		command = defaultSlashCommand
	}

	if teamID == "" || channelID == "" || userID == "" {
		// Slack is malformed or replayed; respond OK with an ephemeral
		// hint so the user sees something rather than a Slack-side error.
		respondSlash(ctx, "Sorry — that command was missing required Slack metadata.")
		return
	}

	subcommand, _ := splitSubcommand(text)

	switch subcommand {
	case "", "help":
		respondSlash(ctx, slashHelpText(command))
	case "subscribe":
		s.handleSlashSubscribe(ctx, teamID, channelID, channelName, userID, triggerID)
	case "unsubscribe":
		s.handleSlashUnsubscribe(ctx, teamID, channelID, channelName, triggerID)
	case "status":
		s.handleSlashStatus(ctx, teamID, channelID)
	default:
		respondSlash(ctx, fmt.Sprintf("Unknown subcommand `%s`.\n\n%s", subcommand, slashHelpText(command)))
	}
}

// handleSlashSubscribe always opens the subscribe modal so the user can pick
// the org, scope, and per-resource filters — symmetric with /nuon unsubscribe.
// We still verify the workspace has an active installation and at least one
// verified org link before opening the modal, so the user gets a useful
// ephemeral message rather than an empty modal in the misconfigured case.
func (s *service) handleSlashSubscribe(
	ctx *gin.Context,
	teamID, channelID, channelName, slackUserID, triggerID string,
) {
	var install app.SlackInstallation
	res := s.db.WithContext(ctx).
		Where(app.SlackInstallation{TeamID: teamID, Status: app.SlackInstallationStatusActive}).
		First(&install)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		respondSlash(ctx, "This workspace doesn't have an active Nuon install. Please re-install from the Nuon dashboard.")
		return
	}
	if res.Error != nil {
		s.l.Error("slash subscribe: lookup installation failed", zap.Error(res.Error), zap.String("team_id", teamID))
		respondSlash(ctx, "Sorry — something went wrong looking up your workspace. Please try again.")
		return
	}

	var links []app.SlackOrgLink
	if err := s.db.WithContext(ctx).
		Where(app.SlackOrgLink{TeamID: teamID, Status: app.SlackOrgLinkStatusVerified}).
		Find(&links).Error; err != nil {
		s.l.Error("slash subscribe: lookup org links failed", zap.Error(err), zap.String("team_id", teamID))
		respondSlash(ctx, "Sorry — something went wrong looking up linked orgs. Please try again.")
		return
	}

	if len(links) == 0 {
		respondSlash(ctx, "This workspace isn't linked to any Nuon org yet. Open the Nuon dashboard to link an org.")
		return
	}

	if triggerID == "" {
		// No trigger_id from Slack means we can't open a modal. This
		// should never happen with a real slash command POST, but
		// guard so we return something useful instead of opaque-failing.
		respondSlash(ctx, "Sorry — Slack didn't send a trigger id. Try the command again.")
		return
	}
	// Pre-populate the modal from this channel's most recently updated
	// existing subscription (if any) so re-running /nuon subscribe doesn't
	// throw away the user's prior org / scope / install / event-filter
	// selections. With no existing sub this returns the zero value and the
	// modal renders its default empty state.
	preselect := s.preselectSubscribeRenderStateForChannel(ctx, teamID, channelID)
	if err := s.openSubscribeModalForSlash(ctx, triggerID, teamID, channelID, channelName, slackUserID, preselect); err != nil {
		s.l.Error("slash subscribe: open modal failed", zap.Error(err),
			zap.String("team_id", teamID), zap.String("channel_id", channelID))
		respondSlash(ctx, "Sorry — couldn't open the subscribe dialog. Please try again.")
		return
	}
	// Modal opens; reply with an empty 200 so Slack doesn't render an
	// extra ephemeral message in the channel.
	ctx.Status(http.StatusOK)
}

// handleSlashUnsubscribe always opens the unsubscribe modal so the user can
// see (and selectively remove) every active subscription targeting this
// channel — across orgs and scopes.
//
// The modal renders an empty-state message when nothing is subscribed, so
// users still get visual confirmation that the command worked.
func (s *service) handleSlashUnsubscribe(ctx *gin.Context, teamID, channelID, channelName, triggerID string) {
	if triggerID == "" {
		// Defensive: real Slack POSTs always include trigger_id. Tell the
		// user to retry instead of silently failing.
		respondSlash(ctx, "Sorry — Slack didn't send a trigger id. Try the command again.")
		return
	}
	if err := s.openUnsubscribeModalForSlash(ctx, triggerID, teamID, channelID, channelName); err != nil {
		s.l.Error("slash unsubscribe: open modal failed", zap.Error(err),
			zap.String("team_id", teamID), zap.String("channel_id", channelID))
		respondSlash(ctx, "Sorry — couldn't open the unsubscribe dialog. Please try again.")
		return
	}
	// Modal opens; reply with empty 200 so Slack doesn't render an extra
	// ephemeral message in-channel.
	ctx.Status(http.StatusOK)
}

// handleSlashStatus reports — ephemerally — what Nuon knows about this
// workspace and channel: whether the installation is active, which Nuon orgs
// the workspace is linked to, and whether this channel has any active
// subscriptions. Read-only and safe to invoke from any channel.
func (s *service) handleSlashStatus(ctx *gin.Context, teamID, channelID string) {
	var install app.SlackInstallation
	res := s.db.WithContext(ctx).
		Where(app.SlackInstallation{TeamID: teamID, Status: app.SlackInstallationStatusActive}).
		First(&install)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		respondSlash(ctx, "This workspace doesn't have an active Nuon install. Please re-install from the Nuon dashboard.")
		return
	}
	if res.Error != nil {
		s.l.Error("slash status: lookup installation failed", zap.Error(res.Error), zap.String("team_id", teamID))
		respondSlash(ctx, "Sorry — something went wrong looking up your workspace. Please try again.")
		return
	}

	var links []app.SlackOrgLink
	if err := s.db.WithContext(ctx).
		Preload("Org").
		Where(app.SlackOrgLink{TeamID: teamID, Status: app.SlackOrgLinkStatusVerified}).
		Find(&links).Error; err != nil {
		s.l.Error("slash status: lookup org links failed", zap.Error(err), zap.String("team_id", teamID))
		respondSlash(ctx, "Sorry — something went wrong looking up linked orgs. Please try again.")
		return
	}

	var subs []app.SlackChannelSubscription
	if err := s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{TeamID: teamID, ChannelID: channelID}).
		Find(&subs).Error; err != nil {
		s.l.Error("slash status: lookup subscriptions failed", zap.Error(err),
			zap.String("team_id", teamID), zap.String("channel_id", channelID))
		respondSlash(ctx, "Sorry — something went wrong looking up subscriptions. Please try again.")
		return
	}

	var b strings.Builder
	fmt.Fprintf(&b, "*Nuon status for this workspace*\n")
	fmt.Fprintf(&b, "• Installation: active\n")

	if len(links) == 0 {
		b.WriteString("• Linked Nuon orgs: none — open the Nuon dashboard to link an org.\n")
	} else {
		fmt.Fprintf(&b, "• Linked Nuon orgs (%d):\n", len(links))
		for _, l := range links {
			name := l.Org.Name
			if name == "" {
				name = l.OrgID
			}
			fmt.Fprintf(&b, "    – %s\n", name)
		}
	}

	if len(subs) == 0 {
		fmt.Fprintf(&b, "• <#%s> subscription: none\n", channelID)
	} else {
		fmt.Fprintf(&b, "• <#%s> subscriptions (%d):\n", channelID, len(subs))
		// Index linked orgs for name lookup; subs may reference an
		// org_link that's no longer verified, in which case we fall back
		// to the org id stored on the sub itself.
		orgNameByLinkID := make(map[string]string, len(links))
		for _, l := range links {
			if l.Org.Name != "" {
				orgNameByLinkID[l.ID] = l.Org.Name
			} else {
				orgNameByLinkID[l.ID] = l.OrgID
			}
		}
		for _, sub := range subs {
			name, ok := orgNameByLinkID[sub.OrgLinkID]
			if !ok {
				name = sub.OrgID
			}
			// describeMatch lives in subscribe_modal.go; the unsubscribe
			// modal uses the same helper so the two surfaces show
			// identical scope tags.
			scopeTag := describeMatch(sub.Match)
			filter := "specific events"
			if sub.Interests.AllEvents {
				filter = "all events"
			}
			fmt.Fprintf(&b, "    – %s · %s · %s\n", name, scopeTag, filter)
		}
	}

	respondSlash(ctx, strings.TrimRight(b.String(), "\n"))
}

// contextWithInstallerAccount resolves the SlackInstallation's installer
// account and stamps it on the context so BeforeCreate hooks can populate
// CreatedByID for slash-command-originated writes (which have no
// dashboard-authenticated account).
func (s *service) contextWithInstallerAccount(ctx context.Context, teamID string) (context.Context, error) {
	var install app.SlackInstallation
	if err := s.db.WithContext(ctx).
		Where(app.SlackInstallation{TeamID: teamID, Status: app.SlackInstallationStatusActive}).
		First(&install).Error; err != nil {
		return ctx, fmt.Errorf("lookup installation for created-by stamp: %w", err)
	}
	var acct app.Account
	if err := s.db.WithContext(ctx).
		Where(app.Account{ID: install.InstalledByAccountID}).
		First(&acct).Error; err != nil {
		return ctx, fmt.Errorf("lookup installer account %q: %w", install.InstalledByAccountID, err)
	}
	return cctx.SetAccountContext(ctx, &acct), nil
}

// splitSubcommand parses the leading subcommand token from `text`. Slack
// passes everything after the slash command as a single string; we split on
// whitespace so `subscribe foo bar` returns ("subscribe", "foo bar").
func splitSubcommand(text string) (string, string) {
	t := strings.TrimSpace(text)
	if t == "" {
		return "", ""
	}
	parts := strings.SplitN(t, " ", 2)
	cmd := strings.ToLower(strings.TrimSpace(parts[0]))
	if len(parts) == 2 {
		return cmd, strings.TrimSpace(parts[1])
	}
	return cmd, ""
}

// respondSlash writes an ephemeral Slack slash-command response.
func respondSlash(ctx *gin.Context, text string) {
	ctx.JSON(http.StatusOK, slashResponse{
		ResponseType: slashResponseTypeEphemeral,
		Text:         text,
	})
}
