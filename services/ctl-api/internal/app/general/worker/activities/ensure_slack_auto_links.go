package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
)

// EnsureSlackAutoLinksRequest has no params — policy comes from a.cfg.SlackAutoLink*.
type EnsureSlackAutoLinksRequest struct{}

type EnsureSlackAutoLinksResult struct {
	OrgsConsidered     int
	LinksCreated       int
	LinksAlreadyActive int
	SubsSeeded         int
	SubsAlreadyExist   int
}

// @temporal-gen-v2 activity
//
// EnsureSlackAutoLinks reconciles SlackOrgLink (and optionally a default
// SlackChannelSubscription) for every org whose labels match the configured
// gate. Idempotent and additive: removals must go through the Slack /nuon
// flow, never this reconciler.
func (a *Activities) EnsureSlackAutoLinks(ctx context.Context, _ EnsureSlackAutoLinksRequest) (*EnsureSlackAutoLinksResult, error) {
	res := &EnsureSlackAutoLinksResult{}

	teamID := a.cfg.SlackAutoLinkTeamID
	labelKey := a.cfg.SlackAutoLinkOrgLabelKey
	channelID := a.cfg.SlackAutoLinkChannelID

	if teamID == "" || labelKey == "" {
		return res, nil
	}

	var install app.SlackInstallation
	switch err := a.db.WithContext(ctx).
		Where(app.SlackInstallation{TeamID: teamID, Status: app.SlackInstallationStatusActive}).
		First(&install).Error; {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return res, nil
	case err != nil:
		return nil, fmt.Errorf("lookup slack installation: %w", err)
	}

	// Empty value → "*" wildcard matches any value at the configured key.
	value := a.cfg.SlackAutoLinkOrgLabelValue
	if value == "" {
		value = "*"
	}
	gate := labels.Labels{labelKey: value}

	var orgs []app.Org
	if err := a.db.WithContext(ctx).
		Scopes(labels.WithLabels("labels", gate)).
		Select("id").
		Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("list matching orgs: %w", err)
	}
	res.OrgsConsidered = len(orgs)

	for _, org := range orgs {
		linkID, created, err := a.upsertAutoLink(ctx, install, org.ID)
		if err != nil {
			return nil, fmt.Errorf("upsert org link for %s: %w", org.ID, err)
		}
		if created {
			res.LinksCreated++
		} else {
			res.LinksAlreadyActive++
		}

		// Never re-seed: a deleted seed sub is intentional — leave it gone.
		if channelID == "" {
			continue
		}
		seeded, err := a.seedDefaultSubscription(ctx, install, linkID, org.ID, channelID)
		if err != nil {
			return nil, fmt.Errorf("seed default sub for %s: %w", org.ID, err)
		}
		if seeded {
			res.SubsSeeded++
		} else {
			res.SubsAlreadyExist++
		}
	}

	a.mw.Count("event_loop.general.slack_auto_link.links_created", int64(res.LinksCreated), nil)
	a.mw.Count("event_loop.general.slack_auto_link.subs_seeded", int64(res.SubsSeeded), nil)
	return res, nil
}

// upsertAutoLink mirrors the OAuth-callback link upsert in
// slack_oauth_callback.go: Unscoped() to revive soft-deleted rows,
// provenance attributed to the workspace installer.
func (a *Activities) upsertAutoLink(ctx context.Context, install app.SlackInstallation, orgID string) (string, bool, error) {
	var existing app.SlackOrgLink
	err := a.db.WithContext(ctx).
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
		if err := a.db.WithContext(ctx).Create(link).Error; err != nil {
			return "", false, fmt.Errorf("create org link: %w", err)
		}
		return link.ID, true, nil
	case err != nil:
		return "", false, fmt.Errorf("lookup existing org link: %w", err)
	}

	// Skip churning updated_at on an already-verified, live row.
	if existing.Status == app.SlackOrgLinkStatusVerified && existing.DeletedAt == 0 {
		return existing.ID, false, nil
	}
	existing.Status = app.SlackOrgLinkStatusVerified
	existing.DeletedAt = 0
	if err := a.db.WithContext(ctx).Unscoped().Save(&existing).Error; err != nil {
		return "", false, fmt.Errorf("revive org link: %w", err)
	}
	return existing.ID, true, nil
}

// seedDefaultSubscription inserts one nil-Match (org-wide), AllEvents-true
// sub when no live sub exists on this link. The "no existing sub" gate makes
// it one-shot: once an admin deletes or replaces the seed, we won't recreate.
// A failed channel-name lookup falls back to the channel ID — a Slack API
// blip must not leave orgs unlinked.
func (a *Activities) seedDefaultSubscription(ctx context.Context, install app.SlackInstallation, linkID, orgID, channelID string) (bool, error) {
	var count int64
	if err := a.db.WithContext(ctx).
		Model(&app.SlackChannelSubscription{}).
		Where(&app.SlackChannelSubscription{OrgLinkID: linkID, OrgID: orgID}).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("count existing subs: %w", err)
	}
	if count > 0 {
		return false, nil
	}

	channelName := channelID
	if info, err := a.slackClient.ConversationsInfo(ctx, install.BotAccessToken, channelID); err == nil && info.Channel.Name != "" {
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

	// Count gate above is racy; the unique index (team_id, channel_id,
	// org_link_id, match_canonical, deleted_at) is the actual guard.
	if err := a.db.WithContext(ctx).Clauses(clause.OnConflict{
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
