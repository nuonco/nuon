package db

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *AutoMigrate) migrateModels(ctx context.Context) error {
	a.l.Info("running auto migrate")

	models := []interface{}{
		// org basics
		&app.Org{},
		&app.UserOrg{},
		&app.UserToken{},

		// vcs basics
		&app.VCSConnection{},
		&app.VCSConnectionCommit{},

		// apps
		&app.App{},
		&app.AppSandboxConfig{},
		&app.AppInput{},
		&app.AppInputConfig{},
		&app.AppInstaller{},
		&app.AppInstallerMetadata{},

		// built in sandboxes
		&app.Sandbox{},
		&app.SandboxRelease{},

		// installs
		&app.AWSAccount{},
		&app.Install{},
		&app.InstallInputs{},
		&app.InstallSandboxRun{},

		// component configuration
		&app.Component{},
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
	for _, model := range models {
		if err := a.db.WithContext(ctx).AutoMigrate(model); err != nil {
			return err
		}
	}

	return nil
}
