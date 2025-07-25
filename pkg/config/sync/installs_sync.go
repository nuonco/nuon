package sync

import (
	"context"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
)

func (s *sync) syncInstall(ctx context.Context, resource string, install *config.Install) (string, error) {
	isNew := false
	appInstall, err := s.apiClient.GetInstall(ctx, install.Name)
	if err != nil {
		if !nuon.IsNotFound(err) {
			return "", err
		}
		isNew = true
	}

	if isNew {
		appInstall, err = s.apiClient.CreateInstall(ctx, s.appID, &models.ServiceCreateInstallRequest{
			Name:   &install.Name,
			Inputs: install.Inputs,
			AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
				Region: install.AWSRegion,
			},
		})
		if err != nil {
			return "", SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}
	} else {
		_, err := s.apiClient.UpdateInstallInputs(ctx, appInstall.ID, &models.ServiceUpdateInstallInputsRequest{
			Inputs: install.Inputs,
		})
		if err != nil {
			return "", SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}
	}

	return appInstall.ID, nil
}
