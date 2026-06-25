package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/workflow"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) ToggleComponent(ctx context.Context, installID, componentID string, enableFlag, disableFlag, planOnly, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	componentID, err = lookup.ComponentID(ctx, s.api, install.AppID, componentID)
	if err != nil {
		return ui.PrintError(err)
	}

	component, err := s.api.GetAppComponent(ctx, install.AppID, componentID)
	if err != nil {
		return ui.PrintError(err)
	}

	currentlyEnabled := true
	if install.InstallConfig != nil && install.InstallConfig.ComponentToggles != nil {
		if v, ok := install.InstallConfig.ComponentToggles[componentID]; ok {
			currentlyEnabled = v
		}
	}

	enabled := enableFlag
	if !enableFlag && !disableFlag {
		status := "enabled"
		if !currentlyEnabled {
			status = "disabled"
		}

		action := "disable"
		if !currentlyEnabled {
			action = "enable"
		}

		ok, err := bubbles.ShowConfirmDialog(
			fmt.Sprintf("%s is currently %s. Do you want to %s it?", component.Name, status, action),
			s.cfg.Interactive,
		)
		if err != nil {
			return nil
		}
		if !ok {
			return nil
		}
		enabled = !currentlyEnabled
	}

	resp, err := s.api.ToggleInstallComponent(ctx, installID, componentID, &models.ServiceToggleInstallComponentRequest{
		Enabled:  enabled,
		PlanOnly: planOnly,
	})
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		ui.PrintJSON(resp)
		return nil
	}

	workflow.WorkflowApp(ctx, s.cfg, s.api, installID, resp.WorkflowID, false)
	return nil
}
