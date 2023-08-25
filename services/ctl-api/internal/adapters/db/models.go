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
		&app.Sandbox{},
		&app.SandboxRelease{},

		// installs
		&app.AWSAccount{},
		&app.Install{},

		// component configuration
		&app.Component{},
		&app.ComponentConfigConnection{},
		&app.HelmComponentConfig{},
		&app.TerraformModuleComponentConfig{},
		&app.DockerBuildComponentConfig{},
		&app.ExternalImageComponentConfig{},
		&app.ConnectedGithubVCSConfig{},
		&app.PublicGitVCSConfig{},
		&app.BasicDeployConfig{},
		&app.AWSECRImageConfig{},

		// component management
		&app.ComponentBuild{},

		// install management
		&app.InstallDeploy{},
		&app.InstallComponent{},
	}
	for _, model := range models {
		if err := a.db.WithContext(ctx).AutoMigrate(model); err != nil {
			return err
		}
	}

	return nil
}
