package service

import (
	"context"
	"maps"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

func (s *service) findCustomerInstall(ctx context.Context, orgID, ref string) (*app.Install, error) {
	var result *app.Install
	err := authz.New(authz.Params{DB: s.db}).ReadCustomerResources(ctx, orgID, func(tx, _, installs *gorm.DB) error {
		id, err := require.CustomerResourceID(ctx, tx, permissions.KindInstall, orgID, ref)
		if err != nil {
			return err
		}
		var row struct{ ID string }
		if err := installs.Select(views.TableOrViewName(tx, &app.Install{}, ".id")).Where(app.Install{ID: id}).Take(&row).Error; err != nil {
			return stderr.ErrNotFound{Err: err, Description: "install not found"}
		}
		reader := *s
		reader.db = tx
		result, err = reader.loadInstallDetails(tx.Statement.Context, orgID, tx.Where(app.Install{ID: id}))
		if err != nil {
			return err
		}
		result.App = app.App{ID: result.AppID, Name: result.App.Name}
		for i := range result.InstallInputs {
			result.InstallInputs[i].Values = maps.Clone(result.InstallInputs[i].ValuesRedacted)
		}
		if result.CurrentInstallInputs != nil {
			result.CurrentInstallInputs.Values = maps.Clone(result.CurrentInstallInputs.ValuesRedacted)
		}
		return nil
	})
	return result, err
}

type InstallAppRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CustomerInstall struct {
	*app.Install
	App InstallAppRef `json:"app"`
}

func customerInstalls(installs []app.Install) []CustomerInstall {
	result := make([]CustomerInstall, 0, len(installs))
	for i := range installs {
		result = append(result, CustomerInstall{Install: &installs[i], App: InstallAppRef{ID: installs[i].AppID, Name: installs[i].App.Name}})
	}
	return result
}

func liveInstallParents(orgID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(views.TableOrViewName(db, &app.Install{}, ".app_id")+" IN (?)",
			db.Session(&gorm.Session{NewDB: true}).Model(&app.App{}).Select("id").Where(app.App{OrgID: orgID}))
	}
}
