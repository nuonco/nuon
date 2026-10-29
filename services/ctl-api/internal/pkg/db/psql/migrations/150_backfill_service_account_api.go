package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (m *Migrations) Migration150BackfillServiceAccountAPI(ctx context.Context, db *gorm.DB) error {
	apiForRole := map[app.RoleType]permissions.API{
		app.RoleTypeRunner:             permissions.APIRunner,
		app.RoleTypeStack:              permissions.APIStack,
		app.RoleTypePhoneHome:          permissions.APIStack,
		app.RoleTypeTelemetryCollector: permissions.APIRunner,
	}
	roleTypes := make([]app.RoleType, 0, len(apiForRole))
	for roleType := range apiForRole {
		roleTypes = append(roleTypes, roleType)
	}

	const batchSize = 100
	lastID := ""
	for {
		var roles []app.Role
		if err := db.WithContext(ctx).Preload("Policies").
			Where("role_type IN ?", roleTypes).
			Where("id > ?", lastID).
			Order("id").
			Limit(batchSize).
			Find(&roles).Error; err != nil {
			return fmt.Errorf("load service account roles: %w", err)
		}
		if len(roles) == 0 {
			return nil
		}
		for i := range roles {
			role := &roles[i]
			if err := backfillServiceAccountAPI(ctx, db, role, apiForRole[role.RoleType]); err != nil {
				return err
			}
		}
		lastID = roles[len(roles)-1].ID
	}
}

func backfillServiceAccountAPI(ctx context.Context, db *gorm.DB, role *app.Role, api permissions.API) error {
	if len(role.Policies) == 0 {
		return nil
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range role.Policies {
			policy := &role.Policies[i]
			perms := policy.Permissions
			if perms == nil {
				perms = pgtype.Hstore{}
			}
			if raw, ok := perms[permissions.APIPermissionKey]; ok && raw != nil && *raw != "" {
				if *raw != string(api) {
					return fmt.Errorf("role %s has api %s, want %s", role.ID, *raw, api)
				}
			} else {
				perms[permissions.APIPermissionKey] = generics.ToPtr(string(api))
				if err := tx.Model(&app.Policy{}).Where("id = ?", policy.ID).Update("permissions", perms).Error; err != nil {
					return fmt.Errorf("set api on policy %s: %w", policy.ID, err)
				}
			}
		}
		if role.Version < app.RoleVersionV2 {
			if err := tx.Model(&app.Role{}).Where("id = ?", role.ID).Update("version", app.RoleVersionV2).Error; err != nil {
				return fmt.Errorf("set version on role %s: %w", role.ID, err)
			}
		}
		return nil
	})
}
