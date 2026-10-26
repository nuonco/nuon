package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

func (m *Migrations) Migration148BackfillPhoneHomeRoles(ctx context.Context, db *gorm.DB) error {
	const batchSize = 100

	client := account.New(account.Params{
		DB:          db,
		AuthzClient: authz.New(authz.Params{DB: db}),
	})
	lastID := ""
	for {
		var versions []app.InstallStackVersion
		// Forgetting an org or install soft-deletes it but leaves its stack versions,
		// and EnsurePhoneHomeServiceAccount cannot lock a soft-deleted owner. Those
		// versions have no live callback left to scope, so skip them.
		if err := db.WithContext(ctx).
			Select("install_stack_versions.id", "install_stack_versions.org_id", "install_stack_versions.install_stack_id").
			Joins("JOIN orgs ON orgs.id = install_stack_versions.org_id AND orgs.deleted_at = 0").
			Joins("JOIN installs ON installs.id = install_stack_versions.install_id AND installs.org_id = install_stack_versions.org_id AND installs.deleted_at = 0").
			Joins("JOIN install_stacks ON install_stacks.id = install_stack_versions.install_stack_id AND install_stacks.org_id = install_stack_versions.org_id AND install_stacks.deleted_at = 0").
			Where("install_stack_versions.phone_home_token_id <> ?", "").
			Where("install_stack_versions.id > ?", lastID).
			Order("install_stack_versions.id").
			Limit(batchSize).
			Find(&versions).Error; err != nil {
			return fmt.Errorf("load phone home stack versions: %w", err)
		}
		if len(versions) == 0 {
			return nil
		}

		for i := range versions {
			version := &versions[i]
			if _, err := client.EnsurePhoneHomeServiceAccount(
				ctx,
				version.OrgID,
				version.InstallStackID,
				version.ID,
			); err != nil {
				return fmt.Errorf("scope phone home account for stack version %s: %w", version.ID, err)
			}
		}
		lastID = versions[len(versions)-1].ID
	}
}
