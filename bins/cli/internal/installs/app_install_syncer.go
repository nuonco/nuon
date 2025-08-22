package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

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

func (s *appInstallSyncer) syncInstall(ctx context.Context, installCfg *config.Install, installID string) (*models.AppInstall, error) {
	var err error
	view := ui.NewSpinnerView(false)
	view.Start(fmt.Sprintf("syncing install %s", installCfg.Name))

	if installCfg == nil {
		return nil, fmt.Errorf("install config cannot be nil")
	}

	view.Update(fmt.Sprintf("fetching install %s", installCfg.Name))

	if installID == "" {
		appInstall, err := s.syncNewInstall(ctx, installCfg, view)
		if err != nil {
			view.Fail(err)
		}
		return appInstall, err
	}

	appInstall, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		view.Fail(err)
		return nil, fmt.Errorf("error getting install %s: %w", installCfg.Name, err)
	}

	appInstall, err = s.syncExistingInstall(ctx, installCfg, appInstall, view)
	if err != nil {
		view.Fail(err)
	}
	return appInstall, err
}

func (s *appInstallSyncer) syncNewInstall(ctx context.Context, installCfg *config.Install, view *ui.SpinnerView) (*models.AppInstall, error) {
	// Use defaults for any missing inputs.
	{
		view.Update(fmt.Sprintf("fetching latest input config for app %s", s.appID))
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

	view.Update(fmt.Sprintf("creating install %s", installCfg.Name))
	appInstall, err := s.api.CreateInstall(ctx, s.appID, &req)
	if err != nil {
		return nil, fmt.Errorf("error creating install %s: %w", installCfg.Name, err)
	}

	view.Success(fmt.Sprintf("install %s created successfully", appInstall.Name))
	return appInstall, nil
}

func (s *appInstallSyncer) syncExistingInstall(ctx context.Context, installCfg *config.Install, appInstall *models.AppInstall, view *ui.SpinnerView) (*models.AppInstall, error) {
	var err error

	if installCfg.ApprovalOption != config.InstallApprovalOptionUnknown {
		if appInstall.InstallConfig == nil {
			view.Update(fmt.Sprintf("creating install config for %s", appInstall.Name))
			appInstall.InstallConfig, err = s.api.CreateInstallConfig(ctx, appInstall.ID, &models.ServiceCreateInstallConfigRequest{
				ApprovalOption: installCfg.ApprovalOption.APIType(),
			})
			if err != nil {
				return nil, err
			}
		} else {
			view.Update(fmt.Sprintf("updating install config for %s", appInstall.Name))
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
		view.Update(fmt.Sprintf("updating install metadata for %s", appInstall.Name))
		_, err = s.api.UpdateInstall(ctx, appInstall.ID, &models.ServiceUpdateInstallRequest{
			Name: appInstall.Name,
			Metadata: &models.HelpersInstallMetadata{
				ManagedBy: ManagedByNuonCLIConfig,
			},
		})
	}

	view.Update(fmt.Sprintf("fetching current input values for %s", appInstall.Name))
	currInputs, err := s.api.GetInstallCurrentInputs(ctx, appInstall.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting current inputs for install %s: %w", appInstall.Name, err)
	}
	// Use the current inputs as defaults, for missing values in the current inputs.
	for k, v := range currInputs.Values {
		if _, ok := installCfg.Inputs[k]; !ok {
			installCfg.Inputs[k] = v
		}
	}

	hasChanged := false
	if len(installCfg.Inputs) != len(currInputs.Values) {
		hasChanged = true
	} else {
		// length is same, go through each input to see if any have changed.
		for k, v := range installCfg.Inputs {
			if currInputs.Values[k] != v {
				hasChanged = true
				break
			}
		}
	}

	// If inputs have divereged, update the install inputs.
	if hasChanged {
		view.Update(fmt.Sprintf("updating inputs for install %s", appInstall.Name))
		_, err = s.api.UpdateInstallInputs(ctx, appInstall.ID, &models.ServiceUpdateInstallInputsRequest{
			Inputs: installCfg.Inputs,
		})
		if err != nil {
			return nil, fmt.Errorf("error updating inputs for install %s: %w", appInstall.Name, err)
		}
	}

	view.Success(fmt.Sprintf("install %s updated successfully", appInstall.Name))
	return appInstall, nil
}
