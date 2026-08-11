package authz

import (
	"context"
	"fmt"
	"slices"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ReconcileOrgRoles brings an org's managed roles in line with
// standardOrgRoles: missing roles are created (with their policies), and
// existing roles have their metadata (title, description, contexts, managed)
// updated to match the definition. Existing rows' policies are deliberately
// never modified.
func ReconcileOrgRoles(ctx context.Context, db *gorm.DB, org app.Org) error {
	var existingRoles []app.Role
	if err := db.WithContext(ctx).
		Where(app.Role{OrgID: generics.NewNullString(org.ID)}).
		Find(&existingRoles).Error; err != nil {
		return fmt.Errorf("unable to load roles for org %s: %w", org.ID, err)
	}

	byType := make(map[app.RoleType]app.Role, len(existingRoles))
	for _, role := range existingRoles {
		byType[role.RoleType] = role
	}

	var toCreate []app.Role
	var toUpdate []app.Role
	for _, want := range standardOrgRoles(org.ID) {
		existing, ok := byType[want.RoleType]
		if !ok {
			want.CreatedByID = org.CreatedByID
			for i := range want.Policies {
				want.Policies[i].CreatedByID = org.CreatedByID
			}
			toCreate = append(toCreate, want)
			continue
		}
		if roleMetadataMatches(existing, want) {
			continue
		}
		existing.Title = want.Title
		existing.Description = want.Description
		existing.Contexts = want.Contexts
		existing.Managed = want.Managed
		toUpdate = append(toUpdate, existing)
	}

	if len(toCreate) == 0 && len(toUpdate) == 0 {
		return nil
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(toCreate) > 0 {
			if err := tx.Create(&toCreate).Error; err != nil {
				return fmt.Errorf("unable to create roles for org %s: %w", org.ID, err)
			}
		}
		for i := range toUpdate {
			role := &toUpdate[i]
			if err := tx.Model(role).
				Select("title", "description", "contexts", "managed").
				Updates(app.Role{
					Title:       role.Title,
					Description: role.Description,
					Contexts:    role.Contexts,
					Managed:     role.Managed,
				}).Error; err != nil {
				return fmt.Errorf("unable to update %s role metadata for org %s: %w", role.RoleType, org.ID, err)
			}
		}
		return nil
	})
}

// roleMetadataMatches reports whether an existing role already carries the
// metadata a definition specifies, so reconcile can skip a no-op update.
func roleMetadataMatches(existing, want app.Role) bool {
	return existing.Title == want.Title &&
		existing.Description == want.Description &&
		existing.Managed == want.Managed &&
		slices.Equal(existing.Contexts, want.Contexts)
}
