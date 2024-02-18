package migrations

import "context"

type Migration struct {
	Name string
	Fn   func(context.Context) error
}

func (a *Migrations) GetAll() []Migration {
	return []Migration{
		{
			Name: "001-sql-example",
			Fn:   a.migration001ExampleSQL,
		},
		{
			Name: "002-model-migration",
			Fn:   a.migration002ExampleModel,
		},
		{
			Name: "003-seed",
			Fn:   a.migration003Seed,
		},
		{
			Name: "004-fix-install-cascade-constraints",
			Fn:   a.migration004InstallsCascadeInputs,
		},
		{
			Name: "005-component-dependencies-primary-key",
			Fn:   a.migration005ComponentDependencyPrimaryKey,
		},
		{
			Name: "006-component-dependencies-deleted-at-constraint",
			Fn:   a.migration006ComponentDependencyDeletedAtConstraint,
		},
		{
			Name: "007-component-dependencies-cascading",
			Fn:   a.migration007ComponentDependencyCascade,
		},
		{
			Name: "008-install-deploy-types",
			Fn:   a.migration008InstallDeployType,
		},
		{
			Name: "009-add-app-runner-configs",
			Fn:   a.migration009AddAppRunnerConfigs,
		},
		{
			Name: "010-remove-org-health-check-names",
			Fn:   a.migration010RemoveHealthCheckName,
		},
		{
			Name: "011-remove-app-input-config",
			Fn:   a.migration011RemoveAppInputConfig,
		},
		{
			Name: "012-add-install-input-config-parents",
			Fn:   a.migration012AddInstallInputConfigParents,
		},
		{
			Name: "013-add-install-input-config-parent-not-null",
			Fn:   a.migration013InstallInputParentNotNull,
		},
		{
			Name: "014-app-input-display-name",
			Fn:   a.migration014AppInputDisplayName,
		},
		{
			Name: "015-app-input-display-name-not-nullable",
			Fn:   a.migration015DisplayNameNotNullable,
		},
		{
			Name: "016-input-cascades",
			Fn:   a.migration016InputCascades,
		},
		{
			Name: "017-add-org-types",
			Fn:   a.migration017AddOrgTypes,
		},
		{
			Name: "018-add-user-types",
			Fn:   a.migration018AddUserTypes,
		},
		{
			Name: "019-org-and-user-types-required",
			Fn:   a.migration019OrgAndUserTypesNotNullable,
		},
		{
			Name: "020-install-component-cascades",
			Fn:   a.migration020InstallComponentCascades,
		},
		{
			Name: "021-datadog-test-noop",
			Fn:   a.migration021NoopDatadogTest,
		},
		{
			Name: "022-remove-duplicate-user-tokens-v2",
			Fn:   a.migration022RemoveDuplicateUserTokens,
		},
		{
			Name: "023-user-tokens-unique",
			Fn:   a.migration023UserTokensUniqueConstraint,
		},
		{
			Name: "024-ensure-user-tokens-for-orgs",
			Fn:   a.migration024EnsureUserTokens,
		},
	}
}
