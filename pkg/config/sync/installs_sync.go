package sync

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
)

func (s *sync) syncInstall(ctx context.Context, resource string, install *config.Install) (string, error) {
	isNew := false
	appInstall, err := s.apiClient.GetInstall(ctx, install.Name)
	if err != nil {
		if !nuon.IsNotFound(err) {
			return "", SyncAPIErr{
				Resource: resource,
				Err:      fmt.Errorf("error getting install %s: %w", install.Name, err),
			}
		}
		isNew = true
	}

	if isNew {
		// Use defaults for any missing inputs.
		{
			appInputCfg, err := s.apiClient.GetAppInputLatestConfig(ctx, s.appID)
			if err != nil {
				return "", SyncAPIErr{
					Resource: resource,
					Err:      fmt.Errorf("error getting latest input config for app %s: %w", s.appID, err),
				}
			}

			for _, ic := range appInputCfg.Inputs {
				val, ok := install.Inputs[ic.Name]
				if ok && val != "" {
					continue
				}
				if ic.Default != "" {
					install.Inputs[ic.Name] = ic.Default
				}
			}
		}

		req := models.ServiceCreateInstallRequest{
			Name:   &install.Name,
			Inputs: install.Inputs,
			AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
				Region: install.AWSRegion,
			},
		}
		if install.ApprovalOption != config.InstallApprovalOptionUnknown {
			req.InstallConfig = &models.HelpersCreateInstallConfigParams{
				ApprovalOption: install.ApprovalOption.APIType(),
			}
		}
		appInstall, err = s.apiClient.CreateInstall(ctx, s.appID, &req)
		if err != nil {
			return "", SyncAPIErr{
				Resource: resource,
				Err:      fmt.Errorf("error creating install %s: %w", install.Name, err),
			}
		}
	} else {
		if install.ApprovalOption != config.InstallApprovalOptionUnknown {
			if appInstall.InstallConfig == nil {
				appInstall.InstallConfig, err = s.apiClient.CreateInstallConfig(ctx, appInstall.ID, &models.ServiceCreateInstallConfigRequest{
					ApprovalOption: install.ApprovalOption.APIType(),
				})
				if err != nil {
					return "", SyncAPIErr{
						Resource: resource,
						Err:      err,
					}
				}
			} else {
				if appInstall.InstallConfig.ApprovalOption != install.ApprovalOption.APIType() {
					// Update the install config if the approval option has changed.
					_, err := s.apiClient.UpdateInstallConfig(ctx, appInstall.ID, appInstall.InstallConfig.ID, &models.ServiceUpdateInstallConfigRequest{
						ApprovalOption: install.ApprovalOption.APIType(),
					})
					if err != nil {
						return "", SyncAPIErr{
							Resource: resource,
							Err:      err,
						}
					}
				}
			}
		}

		currInputs, err := s.apiClient.GetInstallCurrentInputs(ctx, appInstall.ID)
		if err != nil {
			return "", SyncAPIErr{
				Resource: resource,
				Err:      fmt.Errorf("error getting current inputs for install %s: %w", appInstall.Name, err),
			}
		}
		// Use the current inputs as defaults, for missing values in the current inputs.
		for k, v := range currInputs.Values {
			if _, ok := install.Inputs[k]; !ok {
				install.Inputs[k] = v
			}
		}

		hasChanged := false
		if len(install.Inputs) != len(currInputs.Values) {
			hasChanged = true
		} else {
			// length is same, go through each input to see if any have changed.
			for k, v := range install.Inputs {
				if currInputs.Values[k] != v {
					hasChanged = true
					break
				}
			}
		}

		// If inputs have divereged, update the install inputs.
		if hasChanged {
			_, err = s.apiClient.UpdateInstallInputs(ctx, appInstall.ID, &models.ServiceUpdateInstallInputsRequest{
				Inputs: install.Inputs,
			})
			if err != nil {
				return "", SyncAPIErr{
					Resource: resource,
					Err:      fmt.Errorf("error updating inputs for install %s: %w", appInstall.Name, err),
				}
			}
		}
	}

	return appInstall.ID, nil
}
