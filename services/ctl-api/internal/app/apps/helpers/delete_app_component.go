package helpers

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Helpers) DeleteAppComponent(ctx context.Context, compID string) error {
	comp := app.Component{
		ID: compID,
	}

	res := s.db.WithContext(ctx).Model(&comp).First(&comp).Where("id = ?", compID)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get component %s: %w", compID, res.Error)
	}

	appID := comp.AppID
	appCfg, err := s.GetAppLatestConfig(ctx, appID)
	if err != nil {
		return fmt.Errorf("unable to get app latest config: %w", err)
	}

	// Check if the component is part of the current app config, if so, do not allow deletion.
	if slices.Contains(appCfg.ComponentIDs, compID) {
		return fmt.Errorf("unable to delete component %s as it's a part of current app config", compID)
	}

	// Check if any active installs are using this component, if so, do not allow deletion.
	{
		installs, err := s.GetAppInstalls(ctx, appID)
		if err != nil {
			return fmt.Errorf("unable to get app installs: %w", err)
		}

		activeInstalls := make([]string, 0)
		for _, inst := range installs {
			// if an install was never attempted, it does not need to be polled
			if len(inst.InstallSandboxRuns) < 1 {
				continue
			}

			if inst.InstallSandboxRuns[0].Status == app.SandboxRunStatusAccessError ||
				inst.InstallSandboxRuns[0].Status == app.SandboxRunStatusDeprovisioned {
				continue
			}

			for _, instComp := range inst.InstallComponents {
				if instComp.ComponentID == comp.ID {
					if instComp.Status == app.InstallComponentStatusInactive || instComp.Status == "" {
						continue
					}
					activeInstalls = append(activeInstalls, inst.ID)
					break
				}
			}
		}

		if len(activeInstalls) > 0 {
			return fmt.Errorf("unable to delete component %s, active installs using this component exists: %s", compID, activeInstalls)
		}
	}

	// Mark component to signal it is queued for deletion.
	res = s.db.WithContext(ctx).Model(&comp).Updates(app.Component{
		Status:            app.ComponentStatusDeleteQueued,
		StatusDescription: "delete has been queued and waiting",
	})

	if res.Error != nil {
		return fmt.Errorf("unable to update component: %w", res.Error)
	}

	if res.RowsAffected < 1 {
		return fmt.Errorf("component not found %s: %w", compID, gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *Helpers) GetAppInstalls(ctx context.Context, appID string) ([]app.Install, error) {
	cmpApp := app.App{
		ID: appID,
	}

	res := s.db.WithContext(ctx).Model(&cmpApp).
		Preload("Installs").
		Preload("Installs.InstallComponents").
		Preload("Installs.InstallSandboxRuns").
		First(&cmpApp).Where("id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installs for app %s: %w", appID, res.Error)
	}

	return cmpApp.Installs, nil
}
