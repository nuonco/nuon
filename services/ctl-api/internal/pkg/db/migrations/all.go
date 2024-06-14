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
			Name: "049-create-accounts",
			Fn:   a.migration049CreateAccounts,
		},
		{
			Name: "050-created-by-constraints",
			Fn:   a.migration050UpdateCreatedByConstraints,
		},
		{
			Name: "051-create-org-roles-and-policies",
			Fn:   a.migration051OrgRolesAndPolicies,
		},
		{
			Name: "052-create-org-permissions",
			Fn:   a.migration052CreateOrgPermissions,
		},
		{
			Name: "053-drop-user-token-indexes",
			Fn:   a.migration053UserTokensLegacyIndexes,
		},
		{
			Name: "054-drop-sandbox-releases",
			Fn:   a.migration054DropSandboxReleases,
		},
		{
			Name: "055-user-token-account-ids",
			Fn:   a.migration055UserTokenAccountIDs,
		},
		{
			Name: "056-create-tokens-table",
			Fn:   a.migration056CreateTokensTable,
		},
		// TODO(jm): cleanup user_orgs, and remove email/subject from user_tokens
	}
}
