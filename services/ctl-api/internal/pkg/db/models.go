package db

import "github.com/powertoolsdev/mono/services/ctl-api/internal/app"

// declare all models in the correct order they should be migrated.
func allModels() []interface{} {
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

		// internal
		&app.Migration{},
	}
}
