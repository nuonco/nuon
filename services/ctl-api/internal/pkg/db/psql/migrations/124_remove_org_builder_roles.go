package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Migration124RemoveOrgBuilderRoles retires the deprecated org_builder role: a
// prod audit (2026-08-07) found no live assignments beyond test artifacts, so
// every org's org_builder role, its policy, and any remaining assignments are
// soft-deleted.
func (m *Migrations) Migration124RemoveOrgBuilderRoles(ctx context.Context, db *gorm.DB) error {
	const batchSize = 50

	for {
		var roles []app.Role
		if err := db.WithContext(ctx).
			Where(app.Role{RoleType: app.RoleTypeOrgBuilder}).
			Limit(batchSize).
			Find(&roles).Error; err != nil {
			return fmt.Errorf("unable to fetch org_builder roles: %w", err)
		}
		if len(roles) == 0 {
			return nil
		}

		for _, role := range roles {
			role := role
			if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				if err := tx.
					Where(app.AccountRole{RoleID: role.ID}).
					Delete(&app.AccountRole{}).Error; err != nil {
					return err
				}
				if err := tx.
					Where(app.Policy{RoleID: role.ID}).
					Delete(&app.Policy{}).Error; err != nil {
					return err
				}
				return tx.Delete(&role).Error
			}); err != nil {
				return fmt.Errorf("unable to remove org_builder role %s: %w", role.ID, err)
			}
		}
	}
}
