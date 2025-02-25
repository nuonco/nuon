package migrations

import "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"

func (m *Migrations) All() []migrations.Migration {
	return []migrations.Migration{
		{
			Name: "01-create-internal-accounts",
			Fn:   m.migration01InternalAccounts,
		},
	}
}
