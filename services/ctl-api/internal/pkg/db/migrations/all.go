package migrations

import "context"

type Migration struct {
	Name      string
	Fn        func(context.Context) error
	Disabled  bool
	AlwaysRun bool
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
			Name: "041-app-config-view-version",
			Fn:   a.migration041AppConfigVersions,
		},
		{
			Name: "043-component-config-connections-view-version",
			Fn:   a.migration043ComponentConfigConnectionsView,
		},
		{
			Name: "044-installs-view-version",
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
		{
			Name: "059-install-status",
			Fn:   a.migration059InstallStatus,
		},
		{
			Name: "060-installs-view-v2",
			Fn:   a.migration060InstallsViewV2,
		},
		{
			Name: "061-installs-view-v3",
			Fn:   a.migration061InstallsViewV3,
		},
		{
			Name: "062-installers-unique-org",
			Fn:   a.migration062InstallerUniqueOrg,
		},
		{
			Name: "064-drop-installers-unique-org",
			Fn:   a.migration064RemoveInstallerConstraint,
		},
		{
			Name: "065-drop-app-config-fields",
			Fn:   a.migration065ConfigFields,
		},
		{
			Name: "063-drop-org-id-requirements",
			Fn:   a.migration063RoleDropOrgIDRequirements,
		},
		{
			Name:      "066-install-inputs-view-v1",
			Fn:        a.migration066InstallInputsViewV1,
			AlwaysRun: true,
		},
		{
			Name: "067-drop-runner-job-owner-index",
			Fn:   a.migration067DropRunnerJobOwnerIndex,
		},
		{
			Name: "068-drop-custom-cert",
			Fn:   a.migration068DropCustomCert,
		},
		{
			Name: "069-v2-to-default-orgs",
			Fn:   a.migration069V2ToDefaultOrgs,
		},
		{
			Name:      "070-table-sizes-view",
			Fn:        a.migration070TableSizesView,
			AlwaysRun: true,
		},
		{
			Name: "071-drop-settings-refresh-timeout",
			Fn:   a.migration071DropSettingsRefreshTimeout,
		},
		{
			Name:      "072-runner-settings-group-view",
			Fn:        a.migration072RunnerSettings,
			AlwaysRun: true,
		},
		{
			Name: "073-runner-jobs-drop-len-check-on-owner-type",
			Fn:   a.migration073DropLengthCheckOnOwnerType,
		},
		{
			Name:      "074-runner-wide-view",
			Fn:        a.migration074RunnerWideView,
			AlwaysRun: true,
		},
		{
			Name: "075-internal-accounts",
			Fn:   a.migration075InternalAccounts,
		},
		{
			Name:      "076-action-workflow-configs-view",
			Fn:        a.migration076ActionsWorkflowsView,
			AlwaysRun: true,
		},
		{
			Name:      "077-runner-jobs-view-v1",
			Fn:        a.migration077RunnerJobsView,
			AlwaysRun: true,
		},
		{
			Name:      "078-app-configs-view-v2",
			Fn:        a.migration078AppConfigsViewV2,
			AlwaysRun: true,
		},
		{
			Name:      "079-action-runs-active-to-finished",
			Fn:        a.migration079ActionRunsActiveToFinished,
			AlwaysRun: true,
		},
		{
			Name:      "080-runner-health-checks-view-v1",
			Fn:        a.migration080RunnerHealthChecks,
			AlwaysRun: true,
		},
		{
			Name: "081-drop-runner-health-checks-idx",
			Fn:   a.migration081DropRunnerHealtCheckIndex,
		},
		{
			Name: "082-recreate-install-action-workflows-idx",
			Fn:   a.migration082InstallActionWorkflowsIdx,
		},
		{
			Name: "083-clickhouse-table-sizes",
			Fn:   a.migration083ClickhouseTableSizes,
		},
		{
			Name: "084-psql-table-sizes",
			Fn:   a.migration070TableSizesView,
		},
		{
			Name: "085-actions-latest-runs",
			Fn:   a.migration085ActionsLatestRuns,
		},
	}
}
