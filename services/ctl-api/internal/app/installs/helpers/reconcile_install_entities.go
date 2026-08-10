package helpers

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Derived from action_workflow_configs, not AppConfig.ActionIDs: configs synced
// before server-side sync never populated that column, and nil there is
// indistinguishable from "this app has no actions".
func (h *Helpers) desiredActionWorkflowIDs(ctx context.Context, appConfigID string) ([]string, error) {
	var ids []string
	if err := h.db.WithContext(ctx).
		Model(&app.ActionWorkflowConfig{}).
		Where(app.ActionWorkflowConfig{AppConfigID: appConfigID}).
		Distinct().
		Pluck("action_workflow_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("unable to get action workflow configs: %w", err)
	}
	return ids, nil
}

// Same rationale as desiredActionWorkflowIDs.
func (h *Helpers) desiredRunbookIDs(ctx context.Context, appConfigID string) ([]string, error) {
	var ids []string
	if err := h.db.WithContext(ctx).
		Model(&app.RunbookConfig{}).
		Where(app.RunbookConfig{AppConfigID: appConfigID}).
		Distinct().
		Pluck("runbook_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("unable to get runbook configs: %w", err)
	}
	return ids, nil
}

func (h *Helpers) installAppConfig(ctx context.Context, installID string, appConfigCols ...string) (*app.AppConfig, error) {
	var install app.Install
	if err := h.db.WithContext(ctx).
		Select("id", "app_config_id").
		First(&install, "id = ?", installID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	var appCfg app.AppConfig
	if err := h.db.WithContext(ctx).
		Select(append([]string{"id"}, appConfigCols...)).
		First(&appCfg, "id = ?", install.AppConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to get app config: %w", err)
	}

	return &appCfg, nil
}

func (h *Helpers) ReconcileInstallActions(ctx context.Context, installID string) error {
	appCfg, err := h.installAppConfig(ctx, installID)
	if err != nil {
		return err
	}

	desired, err := h.desiredActionWorkflowIDs(ctx, appCfg.ID)
	if err != nil {
		return err
	}

	desiredSet := make(map[string]bool, len(desired))
	for _, id := range desired {
		desiredSet[id] = true
	}

	var existing []app.InstallActionWorkflow
	if err := h.db.WithContext(ctx).
		Where(app.InstallActionWorkflow{InstallID: installID}).
		Find(&existing).Error; err != nil {
		return fmt.Errorf("unable to get existing install action workflows: %w", err)
	}

	existingSet := make(map[string]*app.InstallActionWorkflow, len(existing))
	for i := range existing {
		existingSet[existing[i].ActionWorkflowID] = &existing[i]
	}

	var toCreate []app.InstallActionWorkflow
	for _, actionID := range desired {
		iaw, exists := existingSet[actionID]
		if !exists {
			toCreate = append(toCreate, app.InstallActionWorkflow{
				ActionWorkflowID: actionID,
				InstallID:        installID,
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
				if err := h.db.WithContext(ctx).
					Model(iaw).
					Update("status_v2", restoredStatus).Error; err != nil {
					return fmt.Errorf("unable to restore install action workflow %s: %w", iaw.ID, err)
				}
			}
		}
	}

	if len(toCreate) > 0 {
		if err := h.db.WithContext(ctx).
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
		if err := h.db.WithContext(ctx).
			Model(iaw).
			Update("status_v2", removedStatus).Error; err != nil {
			return fmt.Errorf("unable to mark install action workflow %s as removed: %w", iaw.ID, err)
		}
	}

	return nil
}

func (h *Helpers) ReconcileInstallRunbooks(ctx context.Context, installID string) error {
	appCfg, err := h.installAppConfig(ctx, installID)
	if err != nil {
		return err
	}

	desired, err := h.desiredRunbookIDs(ctx, appCfg.ID)
	if err != nil {
		return err
	}

	desiredSet := make(map[string]bool, len(desired))
	for _, id := range desired {
		desiredSet[id] = true
	}

	var existing []app.InstallRunbook
	if err := h.db.WithContext(ctx).
		Where(app.InstallRunbook{InstallID: installID}).
		Find(&existing).Error; err != nil {
		return fmt.Errorf("unable to get existing install runbooks: %w", err)
	}

	existingSet := make(map[string]*app.InstallRunbook, len(existing))
	for i := range existing {
		existingSet[existing[i].RunbookID] = &existing[i]
	}

	var toCreate []app.InstallRunbook
	for _, runbookID := range desired {
		irb, exists := existingSet[runbookID]
		if !exists {
			toCreate = append(toCreate, app.InstallRunbook{
				RunbookID: runbookID,
				InstallID: installID,
			})
			continue
		}

		if irb.StatusV2.Metadata != nil {
			if _, removed := irb.StatusV2.Metadata["removed_at_app_config_id"]; removed {
				restoredStatus := app.CompositeStatus{
					Status:                 app.StatusPending,
					StatusHumanDescription: fmt.Sprintf("restored by app config %s", appCfg.ID),
					CreatedAtTS:            time.Now().Unix(),
					Metadata:               map[string]any{},
				}
				if err := h.db.WithContext(ctx).
					Model(irb).
					Update("status_v2", restoredStatus).Error; err != nil {
					return fmt.Errorf("unable to restore install runbook %s: %w", irb.ID, err)
				}
			}
		}
	}

	if len(toCreate) > 0 {
		if err := h.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&toCreate).Error; err != nil {
			return fmt.Errorf("unable to create install runbooks: %w", err)
		}
	}

	for i := range existing {
		irb := &existing[i]
		if desiredSet[irb.RunbookID] {
			continue
		}
		if irb.StatusV2.Metadata != nil {
			if _, already := irb.StatusV2.Metadata["removed_at_app_config_id"]; already {
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
		if err := h.db.WithContext(ctx).
			Model(irb).
			Update("status_v2", removedStatus).Error; err != nil {
			return fmt.Errorf("unable to mark install runbook %s as removed: %w", irb.ID, err)
		}
	}

	return nil
}

// Components stay on AppConfig.ComponentIDs: component_config_connections rows are
// deltas, so an unchanged component has no row on the new config and deriving from
// them would drop it.
func (h *Helpers) ReconcileInstallComponents(ctx context.Context, installID string) error {
	appCfg, err := h.installAppConfig(ctx, installID, "component_ids")
	if err != nil {
		return err
	}

	if appCfg.ComponentIDs == nil {
		return nil
	}

	desiredSet := make(map[string]bool, len(appCfg.ComponentIDs))
	for _, id := range appCfg.ComponentIDs {
		desiredSet[id] = true
	}

	var existing []app.InstallComponent
	if err := h.db.WithContext(ctx).
		Where(app.InstallComponent{InstallID: installID}).
		Find(&existing).Error; err != nil {
		return fmt.Errorf("unable to get existing install components: %w", err)
	}

	existingSet := make(map[string]*app.InstallComponent, len(existing))
	for i := range existing {
		existingSet[existing[i].ComponentID] = &existing[i]
	}

	var components []app.Component
	if len(appCfg.ComponentIDs) > 0 {
		if err := h.db.WithContext(ctx).
			Where("id IN ?", []string(appCfg.ComponentIDs)).
			Find(&components).Error; err != nil {
			return fmt.Errorf("unable to get components: %w", err)
		}
	}

	var toCreate []app.InstallComponent
	for _, componentID := range appCfg.ComponentIDs {
		ic, exists := existingSet[componentID]
		if !exists {
			toCreate = append(toCreate, app.InstallComponent{
				ComponentID: componentID,
				InstallID:   installID,
			})
			continue
		}

		if ic.StatusV2.Metadata != nil {
			if _, removed := ic.StatusV2.Metadata["removed_at_app_config_id"]; removed {
				restoredStatus := app.CompositeStatus{
					Status:                 app.Status(app.InstallComponentStatusPending),
					StatusHumanDescription: fmt.Sprintf("restored by app config %s", appCfg.ID),
					CreatedAtTS:            time.Now().Unix(),
					Metadata:               map[string]any{},
				}
				if err := h.db.WithContext(ctx).
					Model(ic).
					Updates(map[string]any{
						"status_v2": restoredStatus,
						"status":    app.InstallComponentStatusPending,
					}).Error; err != nil {
					return fmt.Errorf("unable to restore install component %s: %w", ic.ID, err)
				}
			}
		}
	}

	if len(toCreate) > 0 {
		if err := h.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&toCreate).Error; err != nil {
			return fmt.Errorf("unable to create install components: %w", err)
		}

		if err := h.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(h.componentHelpers.TFWorkspacesFromICs(toCreate)).Error; err != nil {
			return fmt.Errorf("unable to create terraform workspaces: %w", err)
		}

		helmCmps := make(map[string]bool)
		for _, comp := range components {
			if comp.Type == app.ComponentTypeHelmChart {
				helmCmps[comp.ID] = true
			}
		}
		helmCharts := h.componentHelpers.HelmChartFromICs(toCreate, helmCmps)
		if len(helmCharts) > 0 {
			if err := h.db.WithContext(ctx).
				Clauses(clause.OnConflict{DoNothing: true}).
				Create(helmCharts).Error; err != nil {
				return fmt.Errorf("unable to create helm charts: %w", err)
			}
		}
	}

	for i := range existing {
		ic := &existing[i]
		if desiredSet[ic.ComponentID] {
			continue
		}
		if ic.StatusV2.Metadata != nil {
			if _, already := ic.StatusV2.Metadata["removed_at_app_config_id"]; already {
				continue
			}
		}

		removedStatus := app.CompositeStatus{
			Status:                 app.Status(app.InstallComponentStatusInactive),
			StatusHumanDescription: fmt.Sprintf("removed from app config %s", appCfg.ID),
			CreatedAtTS:            time.Now().Unix(),
			Metadata: map[string]any{
				"removed_at_app_config_id": appCfg.ID,
			},
		}
		if err := h.db.WithContext(ctx).
			Model(ic).
			Updates(map[string]any{
				"status_v2": removedStatus,
				"status":    app.InstallComponentStatusInactive,
			}).Error; err != nil {
			return fmt.Errorf("unable to mark install component %s as removed: %w", ic.ID, err)
		}
	}

	return nil
}
