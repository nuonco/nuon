package psql

import "github.com/powertoolsdev/mono/services/ctl-api/internal/app"

type joinTable struct {
	Model     interface{}
	Field     string
	JoinTable interface{}
}

func JoinTables() []joinTable {
	// NOTE: we have to register all join tables manually, since we use soft deletes + custom ID functions
	return []joinTable{
		{
			&app.Component{},
			"Dependencies",
			&app.ComponentDependency{},
		},
		{
			&app.Installer{},
			"Apps",
			&app.InstallerApp{},
		},
		{
			&app.Account{},
			"Roles",
			&app.AccountRole{},
		},
	}
}

// declare all models in the correct order they should be migrated.
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
		&app.OrgHealthCheck{},

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
		&app.Migration{},

		// waitlist
		&app.Waitlist{},
		// NOTE(jm): this is a special table used in both ch and postgres
		&app.TableSize{},
	}
}
