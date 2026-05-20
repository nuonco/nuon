package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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

	switch err := a.db.WithContext(ctx).
		Where(app.SlackInstallation{TeamID: teamID, Status: app.SlackInstallationStatusActive}).
		First(&app.SlackInstallation{}).Error; {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return res, nil
	case err != nil:
		return nil, fmt.Errorf("lookup slack installation: %w", err)
	}

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
		r, err := a.autoLinkHelper.EnsureForOrg(ctx, org.ID)
		if err != nil {
			return nil, fmt.Errorf("ensure for %s: %w", org.ID, err)
		}
		if r.LinkCreated || r.LinkRevived {
			res.LinksCreated++
		} else {
			res.LinksAlreadyActive++
		}
		if channelID != "" {
			if r.SubSeeded {
				res.SubsSeeded++
			} else {
				res.SubsAlreadyExist++
			}
		}
	}

	return res, nil
}
