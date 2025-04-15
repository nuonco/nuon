package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s sync) getAppPermissionsRequest() *models.ServiceCreateAppPermissionsConfigRequest {
	req := &models.ServiceCreateAppPermissionsConfigRequest{
		AppConfigID: generics.ToPtr(s.appConfigID),
	}

	if s.cfg.Permissions.DeprovisionRole != nil {
		req.DeprovisionRole = s.awsIAMRoleToRequest(s.cfg.Permissions.DeprovisionRole)
	}

	if s.cfg.Permissions.ProvisionRole != nil {
		req.ProvisionRole = s.awsIAMRoleToRequest(s.cfg.Permissions.ProvisionRole)
	}

	if s.cfg.Permissions.MaintenanceRole != nil {
		req.MaintenanceRole = s.awsIAMRoleToRequest(s.cfg.Permissions.MaintenanceRole)
	}

	return req
}

func (s sync) syncAppPermissions(ctx context.Context, resource string) error {
	if s.cfg.Permissions == nil {
		return nil
	}

	req := s.getAppPermissionsRequest()
	_, err := s.apiClient.CreateAppPermissionsConfig(ctx, s.appID, req)
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
