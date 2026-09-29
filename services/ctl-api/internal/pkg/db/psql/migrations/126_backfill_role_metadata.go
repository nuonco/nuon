package migrations

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

func (m *Migrations) Migration126BackfillRoleMetadata(ctx context.Context, db *gorm.DB) error {
	const batchSize = 20
	var offset int

	for {
		var orgs []app.Org
		if err := db.WithContext(ctx).Limit(batchSize).Offset(offset).Find(&orgs).Error; err != nil {
			return fmt.Errorf("unable to fetch orgs: %w", err)
		}
		if len(orgs) == 0 {
			return nil
		}

		for _, org := range orgs {
			if err := authz.ReconcileOrgRoles(ctx, db, org); err != nil {
				return fmt.Errorf("unable to reconcile roles for org %s: %w", org.ID, err)
			}
		}

		offset += batchSize
	}
}
