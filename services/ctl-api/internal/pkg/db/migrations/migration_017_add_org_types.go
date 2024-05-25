package migrations

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration017AddOrgTypes(ctx context.Context) error {
	var orgs []*app.Org
	res := a.db.WithContext(ctx).
		Find(&orgs)
	if res.Error != nil {
		return res.Error
	}

	for _, org := range orgs {
		orgTyp := app.OrgTypeReal
		if org.SandboxMode {
			orgTyp = app.OrgTypeSandbox
		}

		// update the pointer to point to the app input config parent
		a.l.Info("adding org type to org")
		res := a.db.WithContext(ctx).
			Model(&app.Org{
				ID: org.ID,
			}).
			Updates(app.Org{OrgType: orgTyp})
		if res.Error != nil {
			return fmt.Errorf("unable to add org type: %w", res.Error)
		}
	}

	return nil
}
