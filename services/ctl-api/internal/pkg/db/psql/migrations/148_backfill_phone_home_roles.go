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
		if err := db.WithContext(ctx).
			Select("id", "org_id", "install_stack_id").
			Where("phone_home_token_id <> ?", "").
			Where("id > ?", lastID).
			Order("id").
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
