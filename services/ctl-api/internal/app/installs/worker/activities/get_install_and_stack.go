package activities

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

type InstallAndStack struct {
	Install *app.Install      `json:"install"`
	Stack   *app.InstallStack `json:"stack"`
}

// @temporal-gen-v2 activity
// @as-wrapper
// @by-field stackID
func (a *Activities) getInstallAndStackForStack(ctx context.Context, stackID string) (*InstallAndStack, error) {
	var stack app.InstallStack
	if res := a.db.WithContext(ctx).
		Where(app.InstallStack{ID: stackID}).
		First(&stack); res.Error != nil {
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get install stack")
	}

	install := app.Install{}
	res := a.db.WithContext(ctx).
		// Install.AfterQuery derives RunnerID, CurrentInstallInputs, SandboxMode
		// and RunnerType from these — dropping one changes a derived field
		// rather than failing.
		Preload("Org").
		Preload("AppConfig").
		Preload("AppRunnerConfig").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Preload("InstallInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(db, &app.InstallInputs{}, ".created_at DESC"))
		}).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("GCPAccount").
		First(&install, "id = ?", stack.InstallID)
	if res.Error != nil {
		return nil, dbgenerics.TemporalGormError(res.Error, "unable to get install for stack: %w")
	}

	return &InstallAndStack{Install: &install, Stack: &stack}, nil
}
