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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
)

// slashResponseTypeEphemeral is the response_type Slack honors for
// "only the invoking user can see this message" replies.
const slashResponseTypeEphemeral = "ephemeral"

// slashHelpText is the canonical help shown for /nuon help and unknown
// subcommands. Kept as a single string (vs. block-kit) since the slash command
// surface is intentionally thin in v1; richer affordances live in the dashboard.
const slashHelpText = `*Nuon Slack commands*
` + "`/nuon subscribe`" + ` — subscribe this channel to Nuon events for the linked workspace
` + "`/nuon unsubscribe`" + ` — remove this channel's subscription
` + "`/nuon status`" + ` — show this workspace's installation, linked orgs, and this channel's subscription
` + "`/nuon help`" + ` — show this message

If your workspace is linked to multiple Nuon orgs, manage subscriptions from the dashboard.`

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
	text := strings.TrimSpace(ctx.PostForm("text"))

	if teamID == "" || channelID == "" || userID == "" {
		// Slack is malformed or replayed; respond OK with an ephemeral
		// hint so the user sees something rather than a Slack-side error.
		respondSlash(ctx, "Sorry — that command was missing required Slack metadata.")
		return
	}

	subcommand, _ := splitSubcommand(text)

	switch subcommand {
	case "", "help":
		respondSlash(ctx, slashHelpText)
	case "subscribe":
		s.handleSlashSubscribe(ctx, teamID, channelID, channelName, userID)
	case "unsubscribe":
		s.handleSlashUnsubscribe(ctx, teamID, channelID)
	case "status":
		s.handleSlashStatus(ctx, teamID, channelID)
	default:
		respondSlash(ctx, fmt.Sprintf("Unknown subcommand `%s`.\n\n%s", subcommand, slashHelpText))
	}
}

// handleSlashSubscribe creates a SlackChannelSubscription for the invoking
// channel, attributing creation to the Slack user (CreatedByAccountID nil —
// the CHECK constraint requires at least one of {slack_user_id, account_id}).
//
// The workspace must have an active SlackInstallation and exactly one verified
// SlackOrgLink. If multiple orgs are linked we punt to the dashboard rather
// than guessing.
func (s *service) handleSlashSubscribe(ctx *gin.Context, teamID, channelID, channelName, slackUserID string) {
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
	if len(links) > 1 {
		respondSlash(ctx, fmt.Sprintf(
			"This workspace is linked to %d Nuon orgs. Subscribing from Slack is ambiguous — please subscribe from the Nuon dashboard.",
			len(links),
		))
		return
	}

	link := links[0]
	if err := s.upsertSlashSubscription(ctx, link, channelID, channelName, slackUserID); err != nil {
		s.l.Error("slash subscribe: create subscription failed", zap.Error(err),
			zap.String("team_id", teamID), zap.String("channel_id", channelID))
		respondSlash(ctx, "Sorry — couldn't create the subscription. Please try again.")
		return
	}

	respondSlash(ctx, fmt.Sprintf("Subscribed <#%s> to Nuon events for this workspace.", channelID))
}

// upsertSlashSubscription creates a sub or revives a previously soft-deleted
// one. Slash-command-originated subs set CreatedBySlackUserID (account id nil).
func (s *service) upsertSlashSubscription(
	ctx context.Context,
	link app.SlackOrgLink,
	channelID, channelName, slackUserID string,
) error {
	var existing app.SlackChannelSubscription
	res := s.db.WithContext(ctx).
		Unscoped().
		Where(app.SlackChannelSubscription{
			OrgLinkID: link.ID,
			TeamID:    link.TeamID,
			ChannelID: channelID,
		}).
		First(&existing)

	switch {
	case errors.Is(res.Error, gorm.ErrRecordNotFound):
		uid := slackUserID
		// Slash-command-originated subs have no Nuon account context;
		// resolve CreatedByID off InstalledByAccount via a synthetic
		// account-context lookup. We fetch the installation's installer
		// account so the BeforeCreate hook can stamp CreatedByID.
		ctxWithAcct, ctxErr := s.contextWithInstallerAccount(ctx, link.TeamID)
		if ctxErr != nil {
			return ctxErr
		}
		// Slash command subs default to AllEvents=true — same as the
		// dashboard create handler — since /nuon subscribe doesn't take
		// per-resource filters.
		sub := &app.SlackChannelSubscription{
			OrgLinkID:            link.ID,
			TeamID:               link.TeamID,
			ChannelID:            channelID,
			ChannelName:          channelName,
			OrgID:                link.OrgID,
			Interests:            interests.AllEvents(),
			CreatedBySlackUserID: &uid,
		}
		return s.db.WithContext(ctxWithAcct).Create(sub).Error
	case res.Error != nil:
		return res.Error
	default:
		uid := slackUserID
		existing.ChannelName = channelName
		existing.CreatedBySlackUserID = &uid
		existing.OrgID = link.OrgID
		existing.DeletedAt = 0
		// Backfill AllEvents for resurrected rows whose Interests column
		// is empty (e.g. legacy text[] rows after the column-drop
		// migration). New rows always have Interests populated, so this
		// is a no-op for them.
		if existing.Interests.IsZero() {
			existing.Interests = interests.AllEvents()
		}
		return s.db.WithContext(ctx).Unscoped().Save(&existing).Error
	}
}

// handleSlashUnsubscribe soft-deletes the active subscription for this
// (team_id, channel_id). Mirrors the subscribe ambiguity check: if the
// workspace is linked to multiple verified Nuon orgs, /nuon unsubscribe is
// ambiguous (it would silently delete subscriptions across every linked org)
// and we refuse, deferring to the dashboard. Idempotent when scoped to a
// single org-link.
func (s *service) handleSlashUnsubscribe(ctx *gin.Context, teamID, channelID string) {
	var links []app.SlackOrgLink
	if err := s.db.WithContext(ctx).
		Where(app.SlackOrgLink{TeamID: teamID, Status: app.SlackOrgLinkStatusVerified}).
		Find(&links).Error; err != nil {
		s.l.Error("slash unsubscribe: lookup org links failed", zap.Error(err),
			zap.String("team_id", teamID))
		respondSlash(ctx, "Sorry — something went wrong looking up linked orgs. Please try again.")
		return
	}

	switch len(links) {
	case 0:
		respondSlash(ctx, "This workspace isn't linked to any Nuon org.")
		return
	case 1:
		// fall through to scoped delete below
	default:
		respondSlash(ctx, fmt.Sprintf(
			"This workspace is linked to %d Nuon orgs. Unsubscribing from Slack is ambiguous — please unsubscribe from the Nuon dashboard.",
			len(links),
		))
		return
	}

	// Scope deletion to the single verified link's org so we never reach
	// across orgs. (The unique index permits the same channel to be
	// subscribed under multiple org_links for the same team_id.)
	link := links[0]
	res := s.db.WithContext(ctx).
		Where(app.SlackChannelSubscription{
			OrgLinkID: link.ID,
			TeamID:    teamID,
			ChannelID: channelID,
		}).
		Delete(&app.SlackChannelSubscription{})
	if res.Error != nil {
		s.l.Error("slash unsubscribe: delete failed", zap.Error(res.Error),
			zap.String("team_id", teamID), zap.String("channel_id", channelID))
		respondSlash(ctx, "Sorry — couldn't remove the subscription. Please try again.")
		return
	}
	if res.RowsAffected == 0 {
		respondSlash(ctx, fmt.Sprintf("<#%s> was not subscribed.", channelID))
		return
	}
	respondSlash(ctx, fmt.Sprintf("Unsubscribed <#%s> from Nuon events.", channelID))
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
			scope := "filtered"
			if sub.Interests.AllEvents {
				scope = "all events"
			}
			fmt.Fprintf(&b, "    – %s (%s)\n", name, scope)
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
