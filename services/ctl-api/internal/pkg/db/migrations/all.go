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
			Name:     "003-seed",
			Fn:       a.migration003Seed,
			Disabled: false,
		},
		{
			Name:     "004-fix-install-cascade-constraints",
			Fn:       a.migration004InstallsCascadeInputs,
			Disabled: true,
		},
		{
			Name:     "005-component-dependencies-primary-key",
			Fn:       a.migration005ComponentDependencyPrimaryKey,
			Disabled: true,
		},
		{
			Name:     "006-component-dependencies-deleted-at-constraint",
			Fn:       a.migration006ComponentDependencyDeletedAtConstraint,
			Disabled: true,
		},
		{
			Name:     "007-component-dependencies-cascading",
			Fn:       a.migration007ComponentDependencyCascade,
			Disabled: true,
		},
		{
			Name:     "008-install-deploy-types",
			Fn:       a.migration008InstallDeployType,
			Disabled: true,
		},
		{
			Name:     "009-add-app-runner-configs",
			Fn:       a.migration009AddAppRunnerConfigs,
			Disabled: true,
		},
		{
			Name:     "010-remove-org-health-check-names",
			Fn:       a.migration010RemoveHealthCheckName,
			Disabled: true,
		},
		{
			Name:     "011-remove-app-input-config",
			Fn:       a.migration011RemoveAppInputConfig,
			Disabled: true,
		},
		{
			Name:     "012-add-install-input-config-parents",
			Fn:       a.migration012AddInstallInputConfigParents,
			Disabled: true,
		},
		{
			Name:     "013-add-install-input-config-parent-not-null",
			Fn:       a.migration013InstallInputParentNotNull,
			Disabled: true,
		},
		{
			Name:     "014-app-input-display-name",
			Fn:       a.migration014AppInputDisplayName,
			Disabled: true,
		},
		{
			Name:     "015-app-input-display-name-not-nullable",
			Fn:       a.migration015DisplayNameNotNullable,
			Disabled: true,
		},
		{
			Name:     "016-input-cascades",
			Fn:       a.migration016InputCascades,
			Disabled: true,
		},
		{
			Name:     "017-add-org-types",
			Fn:       a.migration017AddOrgTypes,
			Disabled: true,
		},
		{
			Name:     "018-add-user-types",
			Fn:       a.migration018AddUserTypes,
			Disabled: true,
		},
		{
			Name:     "019-org-and-user-types-required",
			Fn:       a.migration019OrgAndUserTypesNotNullable,
			Disabled: true,
		},
		{
			Name:     "020-install-component-cascades",
			Fn:       a.migration020InstallComponentCascades,
			Disabled: true,
		},
		{
			Name:     "021-datadog-test-noop",
			Fn:       a.migration021NoopDatadogTest,
			Disabled: true,
		},
		{
			Name:     "022-remove-duplicate-user-tokens-v2",
			Fn:       a.migration022RemoveDuplicateUserTokens,
			Disabled: true,
		},
		{
			Name:     "023-user-tokens-unique",
			Fn:       a.migration023UserTokensUniqueConstraint,
			Disabled: true,
		},
		{
			Name:     "024-ensure-user-tokens-for-orgs",
			Fn:       a.migration024EnsureUserTokens,
			Disabled: true,
		},
		{
			Name:     "025-ensure-created-by-ids-and-org-ids",
			Fn:       a.migration025EnsureCreatedByIDs,
			Disabled: true,
		},
		{
			Name:     "027-delete-installs-with-deleted-orgs",
			Fn:       a.migration027DeleteInstallsWithDeletedOrgs,
			Disabled: true,
		},
		{
			Name:     "028-aws-ecr-image-configs",
			Fn:       a.migration028AWSECRConfigs,
			Disabled: true,
		},
		{
			Name:     "029-vcs-conns-cascade",
			Fn:       a.migration029VcsConnectionsConstraint,
			Disabled: true,
		},
		{
			Name:     "030-org-user-duplicate",
			Fn:       a.migration030OrgUserDuplicates,
			Disabled: true,
		},
		{
			Name:     "031-connected-config-cascade",
			Fn:       a.migration031ConnectedVCSConfigCascadeConstraint,
			Disabled: true,
		},
		{
			Name:     "032-sensitive-inputs",
			Fn:       a.migration033SensitiveInputs,
			Disabled: true,
		},
		{
			Name:     "033-sensitive-input",
			Fn:       a.migration033SensitiveInputs,
			Disabled: true,
		},
		{
			Name:     "033-install-events-cascade",
			Fn:       a.migration033InstallEventsCascade,
			Disabled: true,
		},
		{
			Name:     "034-app-sandbox-config",
			Fn:       a.migration034AppSandboxConfigAppID,
			Disabled: true,
		},
		{
			// REMOVED since the AppInstaller table was removed, breaking the code.
			Name:     "035-installers",
			Fn:       nil,
			Disabled: true,
		},
		{
			Name:     "036-component-var-names",
			Fn:       a.migration036ComponentVarNames,
			Disabled: true,
		},
		{
			Name:     "037-component-var-names-required",
			Fn:       a.migration037ComponentVarNameRequired,
			Disabled: true,
		},
		{
			Name:     "038-drop-app-installer-tables",
			Fn:       a.migration038DropAppInstallers,
			Disabled: true,
		},
		{
			Name:     "039-drop-component-var-name-required",
			Fn:       a.migration039DropComponentVarNameRequired,
			Disabled: true,
		},
		{
			Name:     "040-component-config-versions",
			Fn:       a.migration040ComponentConfigVersions,
			Disabled: true,
		},
		{
			Name: "041-app-config-view",
			Fn:   a.migration041AppConfigVersions,
		},
		{
			Name: "042-drop-component-config-connection-version",
			Fn:   a.migration042ComponentConfigConnectionsDropVersion,
		},
		{
			Name: "043-component-config-connections-view",
			Fn:   a.migration043ComponentConfigConnectionsView,
		},
		{
			Name: "044-installs-view",
			Fn:   a.migration044InstallsView,
		},
	}
}
