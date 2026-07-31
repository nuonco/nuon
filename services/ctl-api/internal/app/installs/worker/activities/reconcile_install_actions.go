package activities

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ReconcileInstallActionsInput struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ReconcileInstallActions(ctx context.Context, input *ReconcileInstallActionsInput) error {
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
		Select("id", "action_ids").
		First(&appCfg, "id = ?", install.AppConfigID).Error; err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	if appCfg.ActionIDs == nil {
		return nil
	}

	desiredSet := make(map[string]bool, len(appCfg.ActionIDs))
	for _, id := range appCfg.ActionIDs {
		desiredSet[id] = true
	}

	var existing []app.InstallActionWorkflow
	if err := a.db.WithContext(ctx).
		Where(app.InstallActionWorkflow{InstallID: input.InstallID}).
		Find(&existing).Error; err != nil {
		return fmt.Errorf("unable to get existing install action workflows: %w", err)
	}

	existingSet := make(map[string]*app.InstallActionWorkflow, len(existing))
	for i := range existing {
		existingSet[existing[i].ActionWorkflowID] = &existing[i]
	}

	var toCreate []app.InstallActionWorkflow
	for _, actionID := range appCfg.ActionIDs {
		iaw, exists := existingSet[actionID]
		if !exists {
			toCreate = append(toCreate, app.InstallActionWorkflow{
				ActionWorkflowID: actionID,
				InstallID:        input.InstallID,
			})
			continue
		}

		if iaw.StatusV2.Metadata != nil {
			if _, removed := iaw.StatusV2.Metadata["removed_at_app_config_id"]; removed {
				restoredStatus := app.CompositeStatus{
					Status:                 app.StatusPending,
					StatusHumanDescription: fmt.Sprintf("restored by app config %s", appCfg.ID),
					CreatedAtTS:            time.Now().Unix(),
					Metadata:               map[string]any{},
				}
				if err := a.db.WithContext(ctx).
					Model(iaw).
					Update("status_v2", restoredStatus).Error; err != nil {
					return fmt.Errorf("unable to restore install action workflow %s: %w", iaw.ID, err)
				}
			}
		}
	}

	if len(toCreate) > 0 {
		if err := a.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&toCreate).Error; err != nil {
			return fmt.Errorf("unable to create install action workflows: %w", err)
		}
	}

	for i := range existing {
		iaw := &existing[i]
		if desiredSet[iaw.ActionWorkflowID] {
			continue
		}
		if iaw.StatusV2.Metadata != nil {
			if _, already := iaw.StatusV2.Metadata["removed_at_app_config_id"]; already {
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
			Model(iaw).
			Update("status_v2", removedStatus).Error; err != nil {
			return fmt.Errorf("unable to mark install action workflow %s as removed: %w", iaw.ID, err)
		}
	}

	return nil
}
