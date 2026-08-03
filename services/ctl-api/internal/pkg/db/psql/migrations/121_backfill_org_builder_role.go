package migrations

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (m *Migrations) Migration121BackfillOrgBuilderRole(ctx context.Context, db *gorm.DB) error {
	const batchSize = 20
	var offset int

	for {
		var orgs []app.Org
		if err := db.WithContext(ctx).Limit(batchSize).Offset(offset).Find(&orgs).Error; err != nil {
			return fmt.Errorf("unable to fetch orgs: %w", err)
		}
		if len(orgs) == 0 {
			break
		}

		for _, org := range orgs {
			var existingRole app.Role
			err := db.WithContext(ctx).
				Where(app.Role{OrgID: generics.NewNullString(org.ID), RoleType: app.RoleTypeOrgBuilder}).
				First(&existingRole).Error
			if err == nil {
				continue
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("unable to check org_builder role for org %s: %w", org.ID, err)
			}

			role := app.Role{
				OrgID:       generics.NewNullString(org.ID),
				CreatedByID: org.CreatedByID,
				RoleType:    app.RoleTypeOrgBuilder,
				Policies: []app.Policy{
					{
						OrgID:       generics.NewNullString(org.ID),
						CreatedByID: org.CreatedByID,
						Name:        app.PolicyNameOrgBuilder,
						Permissions: pgtype.Hstore(map[string]*string{
							org.ID: permissions.PermissionRead.ToStrPtr(),
							permissions.ComponentBuildsObject(org.ID): permissions.PermissionCreate.ToStrPtr(),
						}),
					},
				},
			}

			if err := db.WithContext(ctx).Create(&role).Error; err != nil {
				return fmt.Errorf("unable to create org_builder role for org %s: %w", org.ID, err)
			}
		}

		offset += batchSize
	}

	return nil
}
