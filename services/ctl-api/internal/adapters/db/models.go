package db

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type joinTable struct {
	model     interface{}
	field     string
	joinTable interface{}
}

func (a *AutoMigrate) migrateModels(ctx context.Context) error {
	a.l.Info("running auto migrate")

	// NOTE: we have to register all join tables manually, since we use soft deletes + custom ID functions
	joinTables := []joinTable{
		{
			&app.Component{},
			"Dependencies",
			&app.ComponentDependency{},
		},
	}
	for _, joinTable := range joinTables {
		if err := a.db.WithContext(ctx).SetupJoinTable(joinTable.model, joinTable.field, joinTable.joinTable); err != nil {
			return fmt.Errorf("unable to create join table: %w", err)
		}
	}

	models := []interface{}{
		// org basics
		&app.Org{},
		&app.OrgHealthCheck{},
		&app.UserOrg{},
		&app.UserToken{},

		// vcs basics
		&app.VCSConnection{},
		&app.VCSConnectionCommit{},

		// apps
		&app.App{},
		&app.AppConfig{},
		&app.AppSandboxConfig{},
		&app.AppRunnerConfig{},
		&app.AppInput{},
		&app.AppInputConfig{},
		&app.AppInstaller{},
		&app.AppInstallerMetadata{},

		// built in sandboxes
		&app.Sandbox{},
		&app.SandboxRelease{},

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
	for _, model := range models {
		if err := a.db.WithContext(ctx).AutoMigrate(model); err != nil {
			return err
		}
	}

	return nil
}
