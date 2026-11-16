package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (m *Migrations) Migration153BackfillServiceAccountAPI(ctx context.Context, db *gorm.DB) error {
	apiForRole := map[app.RoleType]permissions.API{
		app.RoleTypeRunner:             permissions.APIRunner,
		app.RoleTypeStack:              permissions.APIStack,
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
		if err := db.WithContext(ctx).
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
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current app.Role
		res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Role{ID: role.ID}).Find(&current)
		if res.Error != nil {
			return fmt.Errorf("lock role %s: %w", role.ID, res.Error)
		}
		if res.RowsAffected == 0 {
			return nil
		}
		var policies []app.Policy
		if err := tx.Model(&app.Policy{}).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Policy{RoleID: role.ID}).Order("id").Scan(&policies).Error; err != nil {
			return fmt.Errorf("load policies for role %s: %w", role.ID, err)
		}
		if len(policies) == 0 {
			return nil
		}
		for _, policy := range policies {
			if policy.Document != nil {
				return nil
			}
		}
		for i := range policies {
			policy := &policies[i]
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
				if err := tx.Model(&app.Policy{}).Where(app.Policy{ID: policy.ID}).Update("permissions", perms).Error; err != nil {
					return fmt.Errorf("set api on policy %s: %w", policy.ID, err)
				}
			}
		}
		if current.Version < app.RoleVersionV2 {
			if err := tx.Model(&app.Role{}).Where(app.Role{ID: role.ID}).Update("version", app.RoleVersionV2).Error; err != nil {
				return fmt.Errorf("set version on role %s: %w", role.ID, err)
			}
		}
		return nil
	})
}
