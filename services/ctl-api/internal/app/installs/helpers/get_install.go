package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @temporal-gen as-activity
func (h *Helpers) GetInstallByID(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	q := h.prepareGetInstall(ctx)
	q.First(&install, "id = ?", installID)

	if q.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", q.Error)
	}

	return &install, nil
}

// @temporal-gen as-activity
func (h *Helpers) GetInstallByName(ctx context.Context, installName, orgID string) (*app.Install, error) {
	install := app.Install{}
	q := h.prepareGetInstall(ctx)
	q.Where("name = ? AND org_id = ?", installName, orgID).
		First(&install)

	if q.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", q.Error)
	}

	return &install, nil
}

// FindInstall loads an Install either by ID, or a combination of name and org ID.
//
// Use this over [GetInstallByID] or [GetInstallByName] when you don't know if the input is an ID or name.
//
// @temporal-gen as-activity
func (h *Helpers) FindInstall(ctx context.Context, installID, orgID string) (*app.Install, error) {
	install := app.Install{}
	q := h.prepareGetInstall(ctx).
		Where("name = ? AND org_id = ?", installID, orgID).
		Or("id = ?", installID).
		First(&install)

	if q.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", q.Error)
	}

	return &install, nil
}

// prepareGetInstall prepares a gorm query object with common preloads, etc. for loading a single install
func (h *Helpers) prepareGetInstall(ctx context.Context) *gorm.DB {
	return h.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Org").
		Preload("Org.RunnerGroup").
		Preload("Org.RunnerGroup.Runners").
		Preload("App").
		Preload("App.Org").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("AppSandboxConfig").
		Preload("AppSandboxConfig.AWSDelegationConfig").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_inputs_view_v1.created_at DESC")
		}).
		Preload("InstallComponents").
		Preload("InstallComponents.InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_deploys.created_at DESC")
		}).
		Preload("InstallComponents.Component").
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC")
		}).
		Preload("InstallSandboxRuns.AppSandboxConfig").

		// load app secrets for deploys
		Preload("App.AppSecrets").
		Preload("AppRunnerConfig").

		// load connected github
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig.VCSConnection").

		// load public git
		Preload("AppSandboxConfig.PublicGitVCSConfig").

		// load runners
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("RunnerGroup.Runners.RunnerGroup")
}
