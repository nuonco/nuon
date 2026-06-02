// Package autolink ensures a SlackOrgLink and optional default channel
// subscription for a single org under the same gating policy the
// reconciler applies, so org creation and the periodic sweep share one
// code path.
package autolink

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

type Helper struct {
	cfg         *internal.Config
	db          *gorm.DB
	slackClient *slackclient.Client
	mw          metrics.Writer
}

type Params struct {
	fx.In

	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	SlackClient *slackclient.Client
	MW          metrics.Writer
}

func New(p Params) *Helper {
	return &Helper{cfg: p.Cfg, db: p.DB, slackClient: p.SlackClient, mw: p.MW}
}

type Result struct {
	LinkCreated   bool
	LinkRevived   bool
	SubSeeded     bool
	SkippedReason string
}

// EnsureForOrg upserts the SlackOrgLink and seeds the default channel
// subscription for one org. Skips silently when the policy is
// unconfigured, the workspace install is missing, or the org is not
// labeled for auto-link.
func (h *Helper) EnsureForOrg(ctx context.Context, orgID string) (*Result, error) {
	res := &Result{}

	teamID := h.cfg.SlackAutoLinkTeamID
	labelKey := h.cfg.SlackAutoLinkOrgLabelKey
	channelID := h.cfg.SlackAutoLinkChannelID
	if teamID == "" || labelKey == "" {
		res.SkippedReason = "policy_unconfigured"
		return res, nil
	}

	var org app.Org
	if err := h.db.WithContext(ctx).Select("id", "labels").Where(&app.Org{ID: orgID}).First(&org).Error; err != nil {
		return nil, fmt.Errorf("lookup org: %w", err)
	}
	if !labelMatches(org.Labels, labelKey, h.cfg.SlackAutoLinkOrgLabelValue) {
		res.SkippedReason = "label_gate_unmatched"
		return res, nil
	}

	var install app.SlackInstallation
	switch err := h.db.WithContext(ctx).
		Where(app.SlackInstallation{TeamID: teamID, Status: app.SlackInstallationStatusActive}).
		First(&install).Error; {
	case errors.Is(err, gorm.ErrRecordNotFound):
		res.SkippedReason = "no_active_installation"
		return res, nil
	case err != nil:
		return nil, fmt.Errorf("lookup slack installation: %w", err)
	}

	linkID, created, err := h.upsertAutoLink(ctx, install, orgID)
	if err != nil {
		return nil, fmt.Errorf("upsert org link: %w", err)
	}
	res.LinkCreated = created
	res.LinkRevived = !created

	if channelID != "" {
		seeded, err := h.seedDefaultSubscription(ctx, install, linkID, orgID, channelID)
		if err != nil {
			return nil, fmt.Errorf("seed default subscription: %w", err)
		}
		res.SubSeeded = seeded
	}

	if h.mw != nil {
		if res.LinkCreated {
			h.mw.Count("general.slack_auto_link.links_created", 1, nil)
		}
		if res.SubSeeded {
			h.mw.Count("general.slack_auto_link.subs_seeded", 1, nil)
		}
	}
	return res, nil
}

// labelMatches treats an empty configured value as a wildcard.
func labelMatches(have labels.Labels, key, want string) bool {
	got, ok := have[key]
	if !ok {
		return false
	}
	if want == "" {
		return true
	}
	return got == want
}

// upsertAutoLink revives soft-deleted rows so a re-link reuses the
// original link id, and attributes the row to the workspace installer.
func (h *Helper) upsertAutoLink(ctx context.Context, install app.SlackInstallation, orgID string) (string, bool, error) {
	var existing app.SlackOrgLink
	err := h.db.WithContext(ctx).
		Unscoped().
		Where(app.SlackOrgLink{TeamID: install.TeamID, OrgID: orgID}).
		First(&existing).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		link := &app.SlackOrgLink{
			TeamID:            install.TeamID,
			OrgID:             orgID,
			Status:            app.SlackOrgLinkStatusVerified,
			LinkedByAccountID: install.InstalledByAccountID,
			CreatedByID:       install.InstalledByAccountID,
		}
		if err := h.db.WithContext(ctx).Create(link).Error; err != nil {
			return "", false, fmt.Errorf("create org link: %w", err)
		}
		return link.ID, true, nil
	case err != nil:
		return "", false, fmt.Errorf("lookup existing org link: %w", err)
	}

	if existing.Status == app.SlackOrgLinkStatusVerified && existing.DeletedAt == 0 {
		return existing.ID, false, nil
	}
	existing.Status = app.SlackOrgLinkStatusVerified
	existing.DeletedAt = 0
	if err := h.db.WithContext(ctx).Unscoped().Save(&existing).Error; err != nil {
		return "", false, fmt.Errorf("revive org link: %w", err)
	}
	return existing.ID, true, nil
}

// seedDefaultSubscription inserts one org-wide AllEvents sub iff none
// exists on this link, so removing or replacing the seed is permanent.
// Falls back to the channel id when the name lookup fails, so a Slack
// API blip doesn't block the link.
func (h *Helper) seedDefaultSubscription(ctx context.Context, install app.SlackInstallation, linkID, orgID, channelID string) (bool, error) {
	var count int64
	if err := h.db.WithContext(ctx).
		Model(&app.SlackChannelSubscription{}).
		Where(&app.SlackChannelSubscription{OrgLinkID: linkID, OrgID: orgID}).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("count existing subs: %w", err)
	}
	if count > 0 {
		return false, nil
	}

	channelName := channelID
	if info, err := h.slackClient.ConversationsInfo(ctx, install.BotAccessToken, channelID); err == nil && info.Channel.Name != "" {
		channelName = info.Channel.Name
	}

	accountID := install.InstalledByAccountID
	sub := app.SlackChannelSubscription{
		OrgLinkID:          linkID,
		OrgID:              orgID,
		TeamID:             install.TeamID,
		ChannelID:          channelID,
		ChannelName:        channelName,
		Interests:          interests.AllEvents(),
		CreatedByAccountID: &accountID,
		CreatedByID:        accountID,
	}

	if err := h.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "team_id"},
			{Name: "channel_id"},
			{Name: "org_link_id"},
			{Name: "match_canonical"},
			{Name: "deleted_at"},
		},
		DoNothing: true,
	}).Create(&sub).Error; err != nil {
		return false, fmt.Errorf("create seed subscription: %w", err)
	}
	return true, nil
}
