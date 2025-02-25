package psql

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// declare all
// models in the correct order they should be migrated.
func AllModels() []interface{} {
	return []interface{}{
		// management, auth and user management
		&app.Role{},
		&app.Account{},
		&app.AccountRole{},
		&app.Token{},
		&app.Policy{},

		&app.NotificationsConfig{},

		// org basics
		&app.Org{},
		&app.OrgInvite{},

		// installers
		&app.Installer{},
		&app.InstallerApp{},
		&app.InstallerMetadata{},

		// vcs basics
		&app.VCSConnection{},
		&app.VCSConnectionCommit{},

		// apps
		&app.App{},
		&app.AppConfig{},
		&app.AppSandboxConfig{},
		&app.AppAWSDelegationConfig{},
		&app.AppRunnerConfig{},
		&app.AppInput{},
		&app.AppInputGroup{},
		&app.AppInputConfig{},
		&app.AppSecret{},

		// installs
		&app.AWSAccount{},
		&app.AzureAccount{},
		&app.Install{},
		&app.InstallEvent{},
		&app.InstallInputs{},
		&app.InstallSandboxRun{},
		&app.InstallIntermediateData{},

		// component configuration
		&app.Component{},
		&app.ComponentDependency{},
		&app.ComponentConfigConnection{},
		&app.HelmComponentConfig{},
		&app.TerraformModuleComponentConfig{},
		&app.DockerBuildComponentConfig{},
		&app.JobComponentConfig{},
		&app.ExternalImageComponentConfig{},
		&app.ConnectedGithubVCSConfig{},
		&app.PublicGitVCSConfig{},
		&app.AWSECRImageConfig{},

		// component management
		&app.ComponentBuild{},
		&app.ComponentRelease{},
		&app.ComponentReleaseStep{},

		// install management
		&app.InstallDeploy{},
		&app.InstallComponent{},

		// log streams
		&app.LogStream{},

		// runner jobs and groups
		&app.RunnerGroup{},
		&app.RunnerOperation{},
		&app.RunnerGroupSettings{},
		&app.Runner{},
		&app.RunnerJob{},
		&app.RunnerJobPlan{},
		&app.RunnerJobExecution{},
		&app.RunnerJobExecutionOutputs{},
		&app.RunnerJobExecutionResult{},

		// diagnostics
		&app.ActionWorkflow{},
		&app.ActionWorkflowConfig{},
		&app.ActionWorkflowStepConfig{},
		&app.ActionWorkflowTriggerConfig{},
		&app.InstallActionWorkflow{},
		&app.InstallActionWorkflowRun{},
		&app.InstallActionWorkflowManualTrigger{},
		&app.InstallActionWorkflowRunStep{},

		// internal
		&migrations.MigrationModel{},

		// waitlist
		&app.Waitlist{},
		// NOTE(jm): this is a special table used in both ch and postgres
		&app.PSQLTableSize{},
	}
}
