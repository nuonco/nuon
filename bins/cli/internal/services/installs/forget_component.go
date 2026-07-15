package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
)

func (s *Service) ForgetComponent(ctx context.Context, installID, componentID string, skipConfirm, asJSON bool) error {
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

	if !skipConfirm {
		ok, err := bubbles.ShowConfirmDialog(
			fmt.Sprintf("Forget %s? This removes it from Nuon's tracking without destroying its infrastructure and cannot be undone.", component.Name),
			s.cfg.Interactive,
		)
		if err != nil {
			return ui.PrintError(err)
		}
		if !ok {
			return nil
		}
	}

	if err := s.api.ForgetInstallComponent(ctx, installID, componentID); err != nil {
		return ui.PrintJSONError(err)
	}

	ui.PrintLn("successfully forgot component")
	return nil
}
