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

const slashResponseTypeEphemeral = "ephemeral"

const defaultSlashCommand = "/nuon"

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

	command := strings.TrimSpace(ctx.PostForm("command"))
	if command == "" {
		command = defaultSlashCommand
	}

	if teamID == "" || channelID == "" || userID == "" {
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
		respondSlash(ctx, "Sorry — Slack didn't send a trigger id. Try the command again.")
		return
	}
	preselect := s.preselectSubscribeRenderStateForChannel(ctx, teamID, channelID)
	if err := s.openSubscribeModalForSlash(ctx, triggerID, teamID, channelID, channelName, slackUserID, preselect); err != nil {
		s.l.Error("slash subscribe: open modal failed", zap.Error(err),
			zap.String("team_id", teamID), zap.String("channel_id", channelID))
		respondSlash(ctx, "Sorry — couldn't open the subscribe dialog. Please try again.")
		return
	}
	ctx.Status(http.StatusOK)
}

func (s *service) handleSlashUnsubscribe(ctx *gin.Context, teamID, channelID, channelName, triggerID string) {
	if triggerID == "" {
		respondSlash(ctx, "Sorry — Slack didn't send a trigger id. Try the command again.")
		return
	}
	if err := s.openUnsubscribeModalForSlash(ctx, triggerID, teamID, channelID, channelName); err != nil {
		s.l.Error("slash unsubscribe: open modal failed", zap.Error(err),
			zap.String("team_id", teamID), zap.String("channel_id", channelID))
		respondSlash(ctx, "Sorry — couldn't open the unsubscribe dialog. Please try again.")
		return
	}
	ctx.Status(http.StatusOK)
}

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

func respondSlash(ctx *gin.Context, text string) {
	ctx.JSON(http.StatusOK, slashResponse{
		ResponseType: slashResponseTypeEphemeral,
		Text:         text,
	})
}
