package migrations

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (m *Migrations) migration051OrgRolesAndPolicies(ctx context.Context) error {
	var orgs []app.Org
	res := m.db.
		WithContext(ctx).
		Find(&orgs)
	if res.Error != nil {
		return fmt.Errorf("unable to get orgs: %w", res.Error)
	}

	for _, org := range orgs {
		ctx = context.WithValue(ctx, "account_id", org.CreatedByID)
		if err := m.authzClient.CreateOrgRoles(ctx, org.ID); err != nil {
			return err
		}
	}

	return nil
}
