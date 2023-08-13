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
			db.AutoMigrate(&app.App{})
			db.AutoMigrate(&app.App{})
			db.AutoMigrate(&app.Build{})
			db.AutoMigrate(&app.AWSSettings{})
			db.AutoMigrate(&app.AWSSettings{})
			db.AutoMigrate(&app.Install{})
			db.AutoMigrate(&app.Instance{})
			db.AutoMigrate(&app.Sandbox{})
			db.AutoMigrate(&app.SandboxRelease{})
			db.AutoMigrate(&app.VCSConnection{})
			return nil
		},
		OnStop: func(_ context.Context) error {
			return nil
		},
	})

	return &AutoMigrate{}
}
