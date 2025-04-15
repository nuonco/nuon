package migrations

import "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"

func (m *Migrations) All() []migrations.Migration {
	return []migrations.Migration{
		{
			Name: "01-create-internal-accounts",
			Fn:   m.migration01InternalAccounts,
		},
		{
			Name: "002-drop-old-actions-run-index",
			Fn:   m.Migration002DropOldActionsRunIndex,
		},
		{
			Name: "086-runner-group-settings-backfill-groups",
			Fn:   m.Migration086RunnerGroupSettingsBackfillGroups,
		},
		{
			Name: "04",
			SQL:  `ALTER TABLE action_workflow_step_configs ALTER COLUMN command DROP NOT NULL;`,
		},
	}
}
