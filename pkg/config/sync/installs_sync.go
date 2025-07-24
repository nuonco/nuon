package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
)

// NOTE(onprem): This is very sub-optimal, especially for new installs we need to iterate through all installs.
// TODO(onprem): Replace with an API call to get an install by name directly.
func (s *sync) getInstallByName(ctx context.Context, name string) (*models.AppInstall, error) {
	hasMore := true
	offset := 0
	paginationLimit := 100

	// Get all app installs with pagination, and loop through them to find the specific install by name.
	for hasMore {
		var appInstalls []*models.AppInstall
		var err error

		appInstalls, hasMore, err = s.apiClient.GetAppInstalls(ctx, s.appID, &models.GetPaginatedQuery{
			PaginationEnabled: true,
			Offset:            offset,
			Limit:             paginationLimit,
		})
		if err != nil {
			return nil, err
		}

		for _, install := range appInstalls {
			if install.Name == name {
				return install, nil
			}
		}

		offset += paginationLimit
	}

	return nil, nil
}

func (s *sync) syncInstall(ctx context.Context, resource string, install *config.Install) (string, error) {
	isNew := false
	appInstall, err := s.getInstallByName(ctx, install.Name)
	if err != nil {
		return "", err
	}

	if appInstall == nil {
		isNew = true

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
	}

	if !isNew {
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
