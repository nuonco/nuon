// Package ensure holds the dependency-free logic that fans an app's action
// workflows and components out across installs. It operates directly on a
// *gorm.DB so both the domain helpers (single-install path) and the apps worker
// activities (app-wide fan-out) can call it without importing each other —
// which would otherwise create an import cycle between the helpers and the
// worker activities package.
package ensure

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	// batch sizes bound how much we hold in memory / send per statement so a large
	// app (many installs and/or many action workflows / components) can't OOM the api.
	actionWorkflowFetchBatchSize = 20
	componentFetchBatchSize      = 20
	installFetchBatchSize        = 20
	insertBatchSize              = 20
)

// ActionWorkflows ensures every action workflow for the app has an
// InstallActionWorkflow row for the given installs. If installIDs is empty,
// every install belonging to the app is ensured.
func ActionWorkflows(ctx context.Context, db *gorm.DB, appID string, installIDs []string) error {
	var actionWorkflows []app.ActionWorkflow
	return db.WithContext(ctx).
		Where(app.ActionWorkflow{AppID: appID}).
		FindInBatches(&actionWorkflows, actionWorkflowFetchBatchSize, func(_ *gorm.DB, _ int) error {
			if len(installIDs) > 0 {
				return createInstallActionWorkflows(ctx, db, installIDs, actionWorkflows)
			}

			var installs []app.Install
			return db.WithContext(ctx).
				Where(app.Install{AppID: appID}).
				FindInBatches(&installs, installFetchBatchSize, func(_ *gorm.DB, _ int) error {
					batchIDs := make([]string, 0, len(installs))
					for _, install := range installs {
						batchIDs = append(batchIDs, install.ID)
					}
					return createInstallActionWorkflows(ctx, db, batchIDs, actionWorkflows)
				}).Error
		}).Error
}

func createInstallActionWorkflows(ctx context.Context, db *gorm.DB, installIDs []string, actionWorkflows []app.ActionWorkflow) error {
	batch := make([]app.InstallActionWorkflow, 0, insertBatchSize)
	flush := func() error {
		if len(batch) < 1 {
			return nil
		}
		res := db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&batch)
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to create install action workflows")
		}
		batch = batch[:0]
		return nil
	}

	for _, installID := range installIDs {
		for _, actionWorkflow := range actionWorkflows {
			batch = append(batch, app.InstallActionWorkflow{
				ActionWorkflowID: actionWorkflow.ID,
				InstallID:        installID,
			})
			if len(batch) < insertBatchSize {
				continue
			}
			if err := flush(); err != nil {
				return err
			}
		}
	}

	return flush()
}

// Components ensures every component for the app has an InstallComponent row
// (plus its terraform workspace and helm chart) for the given installs. If
// installIDs is empty, every install belonging to the app is ensured.
func Components(ctx context.Context, db *gorm.DB, appID string, installIDs []string) error {
	var components []app.Component
	return db.WithContext(ctx).
		Where(app.Component{AppID: appID}).
		FindInBatches(&components, componentFetchBatchSize, func(_ *gorm.DB, _ int) error {
			if len(installIDs) > 0 {
				return createInstallComponents(ctx, db, installIDs, components)
			}

			var installs []app.Install
			return db.WithContext(ctx).
				Where(app.Install{AppID: appID}).
				FindInBatches(&installs, installFetchBatchSize, func(_ *gorm.DB, _ int) error {
					batchIDs := make([]string, 0, len(installs))
					for _, install := range installs {
						batchIDs = append(batchIDs, install.ID)
					}
					return createInstallComponents(ctx, db, batchIDs, components)
				}).Error
		}).Error
}

func createInstallComponents(ctx context.Context, db *gorm.DB, installIDs []string, components []app.Component) error {
	helmCmps := make(map[string]bool, len(components))
	for _, component := range components {
		if component.Type == app.ComponentTypeHelmChart {
			helmCmps[component.ID] = true
		}
	}

	batch := make([]app.InstallComponent, 0, insertBatchSize)
	flush := func() error {
		if len(batch) < 1 {
			return nil
		}
		res := db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&batch)
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to create install components")
		}

		res = db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(tfWorkspacesFromICs(batch))
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to create terraform workspaces")
		}

		helmCharts := helmChartsFromICs(batch, helmCmps)
		if len(helmCharts) > 0 {
			res = db.WithContext(ctx).
				Clauses(clause.OnConflict{DoNothing: true}).
				Create(helmCharts)
			if res.Error != nil {
				return errors.Wrap(res.Error, "unable to create helm releases")
			}
		}

		batch = batch[:0]
		return nil
	}

	for _, installID := range installIDs {
		for _, component := range components {
			batch = append(batch, app.InstallComponent{
				ComponentID: component.ID,
				InstallID:   installID,
			})
			if len(batch) < insertBatchSize {
				continue
			}
			if err := flush(); err != nil {
				return err
			}
		}
	}

	return flush()
}

func tfWorkspacesFromICs(ics []app.InstallComponent) []app.TerraformWorkspace {
	workspaces := make([]app.TerraformWorkspace, 0, len(ics))
	for _, ic := range ics {
		workspaces = append(workspaces, app.TerraformWorkspace{
			OrgID:     ic.OrgID,
			OwnerID:   ic.ID,
			OwnerType: "install_components",
		})
	}
	return workspaces
}

func helmChartsFromICs(ics []app.InstallComponent, helmCmps map[string]bool) []app.HelmChart {
	releases := make([]app.HelmChart, 0, len(ics))
	for _, ic := range ics {
		if !helmCmps[ic.ComponentID] {
			continue
		}
		releases = append(releases, app.HelmChart{
			OrgID:     ic.OrgID,
			OwnerID:   ic.ID,
			OwnerType: "install_components",
		})
	}
	return releases
}
