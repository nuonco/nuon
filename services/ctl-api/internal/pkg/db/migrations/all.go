package migrations

import "context"

type Migration struct {
	Name     string
	Fn       func(context.Context) error
	Disabled bool
}

func (a *Migrations) GetAll() []Migration {
	return []Migration{
		{
			Name:     "001-sql-example",
			Fn:       a.migration001ExampleSQL,
			Disabled: true,
		},
		{
			Name:     "002-model-migration",
			Fn:       a.migration002ExampleModel,
			Disabled: true,
		},
		{
			Name: "041-app-config-view",
			Fn:   a.migration041AppConfigVersions,
		},
		{
			Name: "043-component-config-connections-view",
			Fn:   a.migration043ComponentConfigConnectionsView,
		},
		{
			Name: "044-installs-view",
			Fn:   a.migration044InstallsView,
		},
		{
			Name:     "049-create-accounts",
			Fn:       a.migration049CreateAccounts,
			Disabled: true,
		},
		{
			Name:     "050-created-by-constraints",
			Fn:       a.migration050UpdateCreatedByConstraints,
			Disabled: true,
		},
		{
			Name:     "051-create-org-roles-and-policies",
			Fn:       a.migration051OrgRolesAndPolicies,
			Disabled: true,
		},
		{
			Name:     "052-create-org-permissions",
			Fn:       a.migration052CreateOrgPermissions,
			Disabled: true,
		},
		{
			Name:     "053-drop-user-token-indexes",
			Fn:       a.migration053UserTokensLegacyIndexes,
			Disabled: true,
		},
		{
			Name:     "054-drop-sandbox-releases",
			Fn:       a.migration054DropSandboxReleases,
			Disabled: true,
		},
		{
			Name:     "055-user-token-account-ids",
			Fn:       a.migration055UserTokenAccountIDs,
			Disabled: true,
		},
		{
			Name:     "056-create-tokens-table",
			Fn:       a.migration056CreateTokensTable,
			Disabled: true,
		},
		{
			Name: "057-authz-cleanup",
			Fn:   a.migration057CleanupOldAuthz,
		},
		{
			Name: "058-aws-region-types",
			Fn:   a.migration058AWSRegionTypes,
		},
	}
}
