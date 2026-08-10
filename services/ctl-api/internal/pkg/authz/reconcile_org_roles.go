package authz

import (
	"context"
	"errors"
	"fmt"

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
	for _, want := range standardOrgRoles(org.ID) {
		var existing app.Role
		err := db.WithContext(ctx).
			Where(app.Role{OrgID: generics.NewNullString(org.ID), RoleType: want.RoleType}).
			First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			want.CreatedByID = org.CreatedByID
			for i := range want.Policies {
				want.Policies[i].CreatedByID = org.CreatedByID
			}
			if err := db.WithContext(ctx).Create(&want).Error; err != nil {
				return fmt.Errorf("unable to create %s role for org %s: %w", want.RoleType, org.ID, err)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("unable to check %s role for org %s: %w", want.RoleType, org.ID, err)
		}

		res := db.WithContext(ctx).
			Model(&existing).
			Select("title", "description", "contexts", "managed").
			Updates(app.Role{
				Title:       want.Title,
				Description: want.Description,
				Contexts:    want.Contexts,
				Managed:     want.Managed,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update %s role metadata for org %s: %w", want.RoleType, org.ID, res.Error)
		}
	}

	return nil
}
