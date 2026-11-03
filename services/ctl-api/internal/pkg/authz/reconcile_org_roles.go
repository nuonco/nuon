package authz

import (
	"bytes"
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
	roles, err := standardOrgRoles(org.ID)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireRoleMutationIsolation(tx); err != nil {
			return err
		}
		if err := lockRoleOrg(tx, org.ID); err != nil {
			return err
		}
		if err := tx.Select("id", "created_by_id").Where(app.Org{ID: org.ID}).Take(&org).Error; err != nil {
			return err
		}
		var existingRoles []app.Role
		if err := tx.Unscoped().Preload("Policies").Where(app.Role{OrgID: generics.NewNullString(org.ID)}).
			Find(&existingRoles).Error; err != nil {
			return fmt.Errorf("unable to load roles for org %s: %w", org.ID, err)
		}
		for _, want := range roles {
			var existing *app.Role
			retired := false
			for i := range existingRoles {
				candidate := &existingRoles[i]
				if candidate.RoleType != want.RoleType {
					continue
				}
				if candidate.DeletedAt != 0 {
					retired = true
					continue
				}
				if existing != nil {
					return fmt.Errorf("duplicate catalog role %s requires offline repair", want.RoleType)
				}
				existing = candidate
			}
			if existing == nil {
				if retired {
					return fmt.Errorf("retired catalog role %s requires offline repair", want.RoleType)
				}
				want.CreatedByID = org.CreatedByID
				want.Policies[0].CreatedByID = org.CreatedByID
				if err := tx.Create(&want).Error; err != nil {
					return fmt.Errorf("unable to create role %s: %w", want.RoleType, err)
				}
				continue
			}
			policy, err := decodeScopedRole(*existing, org.ID)
			if err != nil {
				return fmt.Errorf("catalog role %s requires offline migration: %w", want.RoleType, err)
			}
			document, err := policy.MarshalJSON()
			if err != nil {
				return err
			}
			if !existing.Managed || existing.CatalogRevision != want.CatalogRevision ||
				existing.Policies[0].DeletedAt != 0 || !bytes.Equal(document, want.Policies[0].Document) {
				return fmt.Errorf("catalog role %s differs from revision %d; offline migration required", want.RoleType, want.CatalogRevision)
			}
			if roleMetadataMatches(*existing, want) {
				continue
			}
			if err := tx.Model(existing).
				Select("title", "description", "contexts", "managed").
				Updates(app.Role{
					Title:       want.Title,
					Description: want.Description,
					Contexts:    want.Contexts,
					Managed:     want.Managed,
				}).Error; err != nil {
				return fmt.Errorf("unable to update %s role metadata for org %s: %w", want.RoleType, org.ID, err)
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
