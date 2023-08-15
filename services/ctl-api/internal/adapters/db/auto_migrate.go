package db

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AutoMigrate struct{}

func NewAutoMigrate(db *gorm.DB, l *zap.Logger, lc fx.Lifecycle) *AutoMigrate {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			l.Info("running auto migrate")
			db.AutoMigrate(&app.Org{})
			db.AutoMigrate(&app.UserOrg{})
			db.AutoMigrate(&app.VCSConnection{})
			db.AutoMigrate(&app.App{})
			db.AutoMigrate(&app.App{})
			db.AutoMigrate(&app.Build{})
			db.AutoMigrate(&app.AWSAccount{})
			db.AutoMigrate(&app.Install{})
			db.AutoMigrate(&app.Instance{})
			db.AutoMigrate(&app.Sandbox{})
			db.AutoMigrate(&app.SandboxRelease{})
			db.AutoMigrate(&app.Component{})
			db.AutoMigrate(&app.ComponentConfigConnection{})
			db.AutoMigrate(&app.HelmComponentConfig{})
			db.AutoMigrate(&app.TerraformModuleComponentConfig{})
			db.AutoMigrate(&app.DockerBuildComponentConfig{})
			db.AutoMigrate(&app.ExternalImageComponentConfig{})
			db.AutoMigrate(&app.ConnectedGithubVCSConfig{})
			db.AutoMigrate(&app.PublicGitVCSConfig{})
			db.AutoMigrate(&app.BasicDeployConfig{})
			db.AutoMigrate(&app.AWSECRImageConfig{})
			return nil
		},
		OnStop: func(_ context.Context) error {
			return nil
		},
	})

	return &AutoMigrate{}
}
