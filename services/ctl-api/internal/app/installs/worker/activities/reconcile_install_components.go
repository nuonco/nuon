package activities

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ReconcileInstallComponentsInput struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) ReconcileInstallComponents(ctx context.Context, input *ReconcileInstallComponentsInput) error {
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
		Select("id", "component_ids").
		First(&appCfg, "id = ?", install.AppConfigID).Error; err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	if appCfg.ComponentIDs == nil {
		return nil
	}

	desiredSet := make(map[string]bool, len(appCfg.ComponentIDs))
	for _, id := range appCfg.ComponentIDs {
		desiredSet[id] = true
	}

	var existing []app.InstallComponent
	if err := a.db.WithContext(ctx).
		Where(app.InstallComponent{InstallID: input.InstallID}).
		Find(&existing).Error; err != nil {
		return fmt.Errorf("unable to get existing install components: %w", err)
	}

	existingSet := make(map[string]*app.InstallComponent, len(existing))
	for i := range existing {
		existingSet[existing[i].ComponentID] = &existing[i]
	}

	var components []app.Component
	if len(appCfg.ComponentIDs) > 0 {
		if err := a.db.WithContext(ctx).
			Where("id IN ?", []string(appCfg.ComponentIDs)).
			Find(&components).Error; err != nil {
			return fmt.Errorf("unable to get components: %w", err)
		}
	}
	componentsByID := make(map[string]*app.Component, len(components))
	for i := range components {
		componentsByID[components[i].ID] = &components[i]
	}

	var toCreate []app.InstallComponent
	for _, componentID := range appCfg.ComponentIDs {
		ic, exists := existingSet[componentID]
		if !exists {
			toCreate = append(toCreate, app.InstallComponent{
				ComponentID: componentID,
				InstallID:   input.InstallID,
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
				if err := a.db.WithContext(ctx).
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
		if err := a.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&toCreate).Error; err != nil {
			return fmt.Errorf("unable to create install components: %w", err)
		}

		if err := a.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(a.componentsHelpers.TFWorkspacesFromICs(toCreate)).Error; err != nil {
			return fmt.Errorf("unable to create terraform workspaces: %w", err)
		}

		helmCmps := make(map[string]bool)
		for _, comp := range components {
			if comp.Type == app.ComponentTypeHelmChart {
				helmCmps[comp.ID] = true
			}
		}
		helmCharts := a.componentsHelpers.HelmChartFromICs(toCreate, helmCmps)
		if len(helmCharts) > 0 {
			if err := a.db.WithContext(ctx).
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
		if err := a.db.WithContext(ctx).
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
