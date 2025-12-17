package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) ToggleSync(ctx context.Context, installID string, enable, disable bool) error {

	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	appInstall, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return ui.PrintError(fmt.Errorf("error fetching install %s: %w", installID, err))
	}

	managedBy := ManagedByNuonCLIConfig
	if appInstall.Metadata["managed_by"] == ManagedByNuonCLIConfig {
		managedBy = ManagedByNuonDashboard
	}
	// Explicitly set managed_by based if overriding flags are set.
	if enable {
		managedBy = ManagedByNuonCLIConfig
	} else if disable {
		managedBy = ManagedByNuonDashboard
	}

	appInstall, err = s.api.UpdateInstall(ctx, appInstall.ID, &models.ServiceUpdateInstallRequest{
		Name: appInstall.Name,
		Metadata: &models.HelpersInstallMetadata{
			ManagedBy: managedBy,
		},
	})
	if err != nil {
		return ui.PrintError(fmt.Errorf("error toggling install's config syncing: %w", err))
	}

	configuredState := "disabled"
	if appInstall.Metadata["managed_by"] == ManagedByNuonCLIConfig {
		configuredState = "enabled"
	}

	ui.PrintSuccess(fmt.Sprintf("config file syncing is now %s for install %s", configuredState, appInstall.Name))
	return nil
}
