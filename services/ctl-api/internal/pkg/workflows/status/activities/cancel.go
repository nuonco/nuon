package statusactivities

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func CancelIfUnfinished(ctx context.Context, db *gorm.DB, owner any, description string, skipStatusSync bool) (bool, error) {
	var terminal func(string) bool
	columns := []string{"id", "status", "status_v2"}
	switch owner.(type) {
	case *app.InstallDeploy:
		terminal = func(s string) bool { return app.InstallDeployStatus(s).IsTerminal() }
		columns = append(columns, "install_component_id")
	case *app.InstallSandboxRun:
		terminal = func(s string) bool { return app.SandboxRunStatus(s).IsTerminal() }
		columns = append(columns, "install_id", "install_sandbox_id")
	case *app.InstallActionWorkflowRun:
		terminal = func(s string) bool { return app.InstallActionWorkflowRunStatus(s).IsTerminal() }
	default:
		return false, fmt.Errorf("unsupported cancellation owner %T", owner)
	}

	changed := false
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current struct {
			ID                 string
			Status             string
			StatusV2           app.CompositeStatus
			InstallComponentID string
			InstallID          string
			InstallSandboxID   *string
		}
		if err := tx.Scopes(scopes.WithDisableViews).Model(owner).Select(columns).
			Clauses(clause.Locking{Strength: "UPDATE"}).Where(owner, "id").Take(&current).Error; err != nil {
			return err
		}
		if terminal(current.Status) || terminal(string(current.StatusV2.Status)) {
			return nil
		}
		status := app.NewCompositeStatus(ctx, app.StatusCancelled)
		status.StatusHumanDescription = description
		status = prepareCompositeStatus(current.StatusV2, status)
		if err := tx.Scopes(scopes.WithDisableViews).Model(owner).Updates(map[string]any{
			"status": app.StatusCancelled, "status_description": description, "status_v2": status,
		}).Error; err != nil {
			return err
		}
		changed = true
		if skipStatusSync {
			return nil
		}

		var projection any
		switch owner.(type) {
		case *app.InstallDeploy:
			projection = &app.InstallComponent{ID: current.InstallComponentID}
		case *app.InstallSandboxRun:
			var sandbox app.InstallSandbox
			query := tx.Where(app.InstallSandbox{InstallID: current.InstallID})
			if current.InstallSandboxID != nil {
				query = tx.Where(app.InstallSandbox{ID: *current.InstallSandboxID})
			}
			if err := query.First(&sandbox).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil
				}
				return err
			}
			projection = &app.InstallSandbox{ID: sandbox.ID}
			if err := tx.Scopes(scopes.WithDisableViews).Model(owner).Update("install_sandbox_id", sandbox.ID).Error; err != nil {
				return err
			}
		default:
			return nil
		}
		var existing struct{ StatusV2 app.CompositeStatus }
		if err := tx.Model(projection).Select("status_v2").Clauses(clause.Locking{Strength: "UPDATE"}).Where(projection, "id").Take(&existing).Error; err != nil {
			return err
		}
		status = app.NewCompositeStatus(ctx, app.StatusCancelled)
		status.StatusHumanDescription = description
		return tx.Model(projection).Updates(map[string]any{
			"status": app.StatusCancelled, "status_description": description,
			"status_v2": prepareCompositeStatus(existing.StatusV2, status),
		}).Error
	})
	return changed && err == nil, err
}
