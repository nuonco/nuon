package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Migration123SlackInstallationOwnerOrgBackfill backfills owner_org_id: single-link
// workspaces get their sole org; multi-link ones (auto-link fan-out) can't be inferred.
func (m *Migrations) Migration123SlackInstallationOwnerOrgBackfill(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).AutoMigrate(&app.SlackInstallation{}); err != nil {
		return fmt.Errorf("automigrate slack_installations: %w", err)
	}

	var installs []app.SlackInstallation
	if err := db.WithContext(ctx).
		Where(app.SlackInstallation{Status: app.SlackInstallationStatusActive}).
		Where("owner_org_id IS NULL OR owner_org_id = ''").
		Find(&installs).Error; err != nil {
		return fmt.Errorf("list installs: %w", err)
	}

	for _, install := range installs {
		var links []app.SlackOrgLink
		if err := db.WithContext(ctx).
			Where(app.SlackOrgLink{TeamID: install.TeamID, Status: app.SlackOrgLinkStatusVerified}).
			Find(&links).Error; err != nil {
			return fmt.Errorf("list links for team %s: %w", install.TeamID, err)
		}
		if len(links) != 1 {
			continue
		}
		if err := db.WithContext(ctx).
			Model(&app.SlackInstallation{}).
			Where(app.SlackInstallation{ID: install.ID}).
			Update("owner_org_id", links[0].OrgID).Error; err != nil {
			return fmt.Errorf("set owner for install %s: %w", install.ID, err)
		}
	}

	// Explicit owner for the nuon workspace (auto-link fan-out hides it above); no-op elsewhere.
	const nuonTeamID, nuonOwnerOrgID = "T02H4BYC54P", "org4f5hq4tyo44legra6r4nm18"
	if err := db.WithContext(ctx).
		Model(&app.SlackInstallation{}).
		Where("team_id = ? AND (owner_org_id IS NULL OR owner_org_id = '')", nuonTeamID).
		Where("EXISTS (?)", db.Model(&app.SlackOrgLink{}).
			Select("1").
			Where(app.SlackOrgLink{TeamID: nuonTeamID, OrgID: nuonOwnerOrgID, Status: app.SlackOrgLinkStatusVerified})).
		Update("owner_org_id", nuonOwnerOrgID).Error; err != nil {
		return fmt.Errorf("set owner for nuon workspace: %w", err)
	}

	return nil
}
