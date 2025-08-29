package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/pterm/pterm"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
)

const ManagedByNuonCLIConfig = "nuon/cli/install-config"

type appInstallSyncer struct {
	api   nuon.Client
	appID string
}

func newAppInstallSyncer(api nuon.Client, appID string) *appInstallSyncer {
	return &appInstallSyncer{
		api:   api,
		appID: appID,
	}
}

func (s *appInstallSyncer) syncInstall(
	ctx context.Context, installCfg *config.Install, installID string, autoApprove bool,
) (*models.AppInstall, error) {
	var err error
	ui.PrintLn(fmt.Sprintf("syncing install %s", installCfg.Name))

	if installCfg == nil {
		return nil, fmt.Errorf("install config cannot be nil")
	}

	if installID == "" {
		appInstall, err := s.syncNewInstall(ctx, installCfg, autoApprove)
		return appInstall, err
	}

	appInstall, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("error getting install %s: %w", installCfg.Name, err)
	}

	appInstall, err = s.syncExistingInstall(ctx, installCfg, appInstall, autoApprove)
	return appInstall, err
}

func (s *appInstallSyncer) syncNewInstall(ctx context.Context, installCfg *config.Install, autoApprove bool) (*models.AppInstall, error) {
	// Use defaults for any missing inputs.
	{
		appInputCfg, err := s.api.GetAppInputLatestConfig(ctx, s.appID)
		if err != nil {
			return nil, fmt.Errorf("error getting latest input config for app %s: %w", s.appID, err)
		}

		for _, ic := range appInputCfg.Inputs {
			val, ok := installCfg.Inputs[ic.Name]
			if ok && val != "" {
				continue
			}
			if ic.Default != "" {
				installCfg.Inputs[ic.Name] = ic.Default
			}
		}
	}

	diff, _, err := installCfg.Diff(nil)
	pterm.DefaultBasicText.Println(diff)

	if !autoApprove {
		ok, err := pterm.DefaultInteractiveConfirm.Show("Do you want to proceed with creating this install?")
		if err != nil {
			return nil, fmt.Errorf("error getting confirmation: %w", err)
		}
		if !ok {
			ui.PrintSuccess(fmt.Sprintf("skipping install %s, sync aborted by user", installCfg.Name))
			return nil, nil
		}
	}

	req := models.ServiceCreateInstallRequest{
		Name:   &installCfg.Name,
		Inputs: installCfg.Inputs,
		Metadata: &models.HelpersInstallMetadata{
			ManagedBy: ManagedByNuonCLIConfig,
		},
	}
	if installCfg.AWSAccount != nil {
		req.AwsAccount = &models.ServiceCreateInstallRequestAwsAccount{
			Region: installCfg.AWSAccount.Region,
		}
	}
	if installCfg.ApprovalOption != config.InstallApprovalOptionUnknown {
		req.InstallConfig = &models.HelpersCreateInstallConfigParams{
			ApprovalOption: installCfg.ApprovalOption.APIType(),
		}
	}

	appInstall, err := s.api.CreateInstall(ctx, s.appID, &req)
	if err != nil {
		return nil, fmt.Errorf("error creating install %s: %w", installCfg.Name, err)
	}

	ui.PrintSuccess(fmt.Sprintf("install %s created successfully", appInstall.Name))
	return appInstall, nil
}

func (s *appInstallSyncer) syncExistingInstall(
	ctx context.Context, installCfg *config.Install, appInstall *models.AppInstall, autoApprove bool,
) (*models.AppInstall, error) {
	var err error

	currInputs, err := s.api.GetInstallCurrentInputs(ctx, appInstall.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting current inputs for install %s: %w", appInstall.Name, err)
	}

	upstreamConfig := &config.Install{}
	upstreamConfig.ParseIntoInstall(appInstall, currInputs)

	diff, diffRes, err := installCfg.Diff(upstreamConfig)
	if err != nil {
		return nil, fmt.Errorf("error generating diff for install %s: %w", installCfg.Name, err)
	}
	if !diffRes.HasChanged {
		ui.PrintSuccess(fmt.Sprintf("install %s is up to date, no changes needed", installCfg.Name))
		return appInstall, nil
	}

	pterm.DefaultBasicText.Printf(`[install diff]
%s
(added %d, removed %d, changed %d)
`, diff, diffRes.Added, diffRes.Removed, diffRes.Changed)

	if !autoApprove {
		ok, err := pterm.DefaultInteractiveConfirm.Show("Do you want to proceed with creating this install?")
		if err != nil {
			return nil, fmt.Errorf("error getting confirmation: %w", err)
		}
		if !ok {
			ui.PrintSuccess(fmt.Sprintf("skipping install %s, sync aborted by user", installCfg.Name))
			return nil, nil
		}
	}

	if installCfg.ApprovalOption != config.InstallApprovalOptionUnknown {
		if appInstall.InstallConfig == nil {
			appInstall.InstallConfig, err = s.api.CreateInstallConfig(ctx, appInstall.ID, &models.ServiceCreateInstallConfigRequest{
				ApprovalOption: installCfg.ApprovalOption.APIType(),
			})
			if err != nil {
				return nil, err
			}
		} else {
			if appInstall.InstallConfig.ApprovalOption != installCfg.ApprovalOption.APIType() {
				// Update the install config if the approval option has changed.
				_, err := s.api.UpdateInstallConfig(ctx, appInstall.ID, appInstall.InstallConfig.ID, &models.ServiceUpdateInstallConfigRequest{
					ApprovalOption: installCfg.ApprovalOption.APIType(),
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if appInstall.Metadata["managed_by"] != ManagedByNuonCLIConfig {
		_, err = s.api.UpdateInstall(ctx, appInstall.ID, &models.ServiceUpdateInstallRequest{
			Name: appInstall.Name,
			Metadata: &models.HelpersInstallMetadata{
				ManagedBy: ManagedByNuonCLIConfig,
			},
		})
	}

	// Use the current inputs as defaults, for missing values in the current inputs.
	for k, v := range currInputs.Values {
		if _, ok := installCfg.Inputs[k]; !ok {
			installCfg.Inputs[k] = v
		}
	}

	hasInputChanged := false
	if len(installCfg.Inputs) != len(currInputs.Values) {
		hasInputChanged = true
	} else {
		// length is same, go through each input to see if any have changed.
		for k, v := range installCfg.Inputs {
			if currInputs.Values[k] != v {
				hasInputChanged = true
				break
			}
		}
	}

	// If inputs have divereged, update the install inputs.
	if hasInputChanged {
		_, err = s.api.UpdateInstallInputs(ctx, appInstall.ID, &models.ServiceUpdateInstallInputsRequest{
			Inputs: installCfg.Inputs,
		})
		if err != nil {
			return nil, fmt.Errorf("error updating inputs for install %s: %w", appInstall.Name, err)
		}
	}

	ui.PrintSuccess(fmt.Sprintf("install %s updated successfully", appInstall.Name))
	return appInstall, nil
}
