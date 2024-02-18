package migrations

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (a *Migrations) migration026EnsureOrgIDs(ctx context.Context) error {
	a.l.Error("running 026")

	methods := []func(context.Context) error{
		a.migration026AppSandboxConfigs,
		a.migration026PublicGitVCSConfigs,
		a.migration026ComponentBuilds,
		a.migration026AppInstallerMetadata,
		a.migration026AwsAccounts,
	}
	for idx, method := range methods {
		if err := method(ctx); err != nil {
			return fmt.Errorf("method %d failed: %w", idx, err)
		}
	}

	return nil
}

func (a *Migrations) migration026AppInstallerMetadata(ctx context.Context) error {
	var objs []app.AppInstallerMetadata
	res := a.db.Unscoped().WithContext(ctx).
		Find(&objs)
	if res.Error != nil {
		return res.Error
	}

	for _, obj := range objs {
		var installer app.AppInstaller
		res := a.db.Unscoped().WithContext(ctx).
			Find(&installer, "id = ?", obj.AppInstallerID)
		if res.Error != nil {
			return res.Error
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.AppInstallerMetadata{
				ID: obj.ID,
			}).
			Updates(app.AppInstallerMetadata{
				OrgID: installer.OrgID,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update app installer metadata: %w", res.Error)
		}
	}

	return nil
}

func (a *Migrations) migration026AppSandboxConfigs(ctx context.Context) error {
	var orgs []*app.Org
	res := a.db.Unscoped().WithContext(ctx).
		Find(&orgs)
	if res.Error != nil {
		return res.Error
	}
	a.l.Error("total orgs", zap.Int("count", len(orgs)))
	orgsByID := make(map[string]struct{})
	for _, org := range orgs {
		orgsByID[org.ID] = struct{}{}

	}
	// migrate app sandbox configs
	var appSandboxConfigs []app.AppSandboxConfig
	res = a.db.Unscoped().WithContext(ctx).
		Preload("App").
		Find(&appSandboxConfigs)
	if res.Error != nil {
		return res.Error
	}
	for _, obj := range appSandboxConfigs {
		if len(obj.OrgID) < 1 {
			obj.OrgID = obj.App.OrgID
		}

		if _, found := orgsByID[obj.OrgID]; !found {
			res := a.db.Unscoped().WithContext(ctx).
				Delete(&app.AppSandboxConfig{
					ID: obj.ID,
				})
			if res.Error != nil {
				return fmt.Errorf("unable to delete app sandbox config: %w", res.Error)
			}
			continue
		}
		res := a.db.Unscoped().WithContext(ctx).
			Model(&app.AppSandboxConfig{
				ID: obj.ID,
			}).
			Updates(app.AppSandboxConfig{
				OrgID: obj.OrgID,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update app sandbox config: %w", res.Error)
		}
	}
	return nil
}

func (a *Migrations) migration026ComponentBuilds(ctx context.Context) error {
	var objs []app.ComponentBuild
	res := a.db.Unscoped().WithContext(ctx).
		Preload("ComponentConfigConnection").
		Find(&objs)
	if res.Error != nil {
		return res.Error
	}

	for _, obj := range objs {
		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.ComponentBuild{
				ID: obj.ID,
			}).
			Updates(app.ComponentBuild{
				OrgID: obj.ComponentConfigConnection.OrgID,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update component config connection: %w", res.Error)
		}
	}

	return nil
}

func (a *Migrations) migration026AwsAccounts(ctx context.Context) error {
	var objs []app.AWSAccount
	res := a.db.Unscoped().WithContext(ctx).
		Find(&objs)
	if res.Error != nil {
		return res.Error
	}

	for _, obj := range objs {
		var install app.Install
		res := a.db.Unscoped().WithContext(ctx).
			Find(&install, "id = ?", obj.InstallID)
		if res.Error != nil {
			return res.Error
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.AWSAccount{
				ID: obj.ID,
			}).
			Updates(app.AWSAccount{
				OrgID: install.OrgID,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update app install: %w", res.Error)
		}
	}

	return nil
}

func (a *Migrations) migration026PublicGitVCSConfigs(ctx context.Context) error {
	var objs []app.PublicGitVCSConfig
	res := a.db.Unscoped().WithContext(ctx).
		Find(&objs)
	if res.Error != nil {
		return res.Error
	}

	for _, obj := range objs {
		var org app.Org
		res := a.db.Unscoped().WithContext(ctx).
			First(&org, "id = ?", obj.OrgID)
		if res.Error != nil {
			if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return res.Error
			}

			res := a.db.Unscoped().WithContext(ctx).
				Delete(&obj, "id = ?", obj.ID)
			if res.Error != nil {
				return res.Error
			}
		}

		var cfg app.ComponentConfigConnection
		res = a.db.Unscoped().WithContext(ctx).
			Find(&cfg, "id = ?", obj.ComponentConfigID)
		if res.Error != nil {
			return res.Error
		}

		res = a.db.Unscoped().WithContext(ctx).
			Model(&app.PublicGitVCSConfig{
				ID: obj.ID,
			}).
			Updates(app.PublicGitVCSConfig{
				OrgID: cfg.OrgID,
			})
		if res.Error != nil {
			return fmt.Errorf("unable to update public git vcs config: %w", res.Error)
		}
	}

	return nil
}
