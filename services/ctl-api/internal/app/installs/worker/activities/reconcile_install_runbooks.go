package activities

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ReconcileInstallRunbooksInput struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ReconcileInstallRunbooks(ctx context.Context, input *ReconcileInstallRunbooksInput) error {
	if err := a.v.Struct(input); err != nil {
		return fmt.Errorf("invalid input: %w", err)
	}

	var install app.Install
	if err := a.db.WithContext(ctx).
		Select("id", "app_config_id").
		First(&install, "id = ?", input.InstallID).Error; err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	var appCfg app.AppConfig
	if err := a.db.WithContext(ctx).
		Select("id", "runbook_ids").
		First(&appCfg, "id = ?", install.AppConfigID).Error; err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	if appCfg.RunbookIDs == nil {
		return nil
	}

	desiredSet := make(map[string]bool, len(appCfg.RunbookIDs))
	for _, id := range appCfg.RunbookIDs {
		desiredSet[id] = true
	}

	var existing []app.InstallRunbook
	if err := a.db.WithContext(ctx).
		Where(app.InstallRunbook{InstallID: input.InstallID}).
		Find(&existing).Error; err != nil {
		return fmt.Errorf("unable to get existing install runbooks: %w", err)
	}

	existingSet := make(map[string]*app.InstallRunbook, len(existing))
	for i := range existing {
		existingSet[existing[i].RunbookID] = &existing[i]
	}

	var toCreate []app.InstallRunbook
	for _, runbookID := range appCfg.RunbookIDs {
		ir, exists := existingSet[runbookID]
		if !exists {
			toCreate = append(toCreate, app.InstallRunbook{
				RunbookID: runbookID,
				InstallID: input.InstallID,
			})
			continue
		}

		if ir.StatusV2.Metadata != nil {
			if _, removed := ir.StatusV2.Metadata["removed_at_app_config_id"]; removed {
				restoredStatus := app.CompositeStatus{
					Status:                 app.StatusPending,
					StatusHumanDescription: fmt.Sprintf("restored by app config %s", appCfg.ID),
					CreatedAtTS:            time.Now().Unix(),
					Metadata:               map[string]any{},
				}
				if err := a.db.WithContext(ctx).
					Model(ir).
					Update("status_v2", restoredStatus).Error; err != nil {
					return fmt.Errorf("unable to restore install runbook %s: %w", ir.ID, err)
				}
			}
		}
	}

	if len(toCreate) > 0 {
		if err := a.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&toCreate).Error; err != nil {
			return fmt.Errorf("unable to create install runbooks: %w", err)
		}
	}

	for i := range existing {
		ir := &existing[i]
		if desiredSet[ir.RunbookID] {
			continue
		}
		if ir.StatusV2.Metadata != nil {
			if _, already := ir.StatusV2.Metadata["removed_at_app_config_id"]; already {
				continue
			}
		}

		removedStatus := app.CompositeStatus{
			Status:                 app.Status("inactive"),
			StatusHumanDescription: fmt.Sprintf("removed from app config %s", appCfg.ID),
			CreatedAtTS:            time.Now().Unix(),
			Metadata: map[string]any{
				"removed_at_app_config_id": appCfg.ID,
			},
		}
		if err := a.db.WithContext(ctx).
			Model(ir).
			Update("status_v2", removedStatus).Error; err != nil {
			return fmt.Errorf("unable to mark install runbook %s as removed: %w", ir.ID, err)
		}
	}

	return nil
}
