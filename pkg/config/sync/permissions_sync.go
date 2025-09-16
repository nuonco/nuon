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

	if s.cfg.BreakGlass != nil && len(s.cfg.BreakGlass.Roles) > 0 {
		breakGlassRoles := make([]*models.ServiceAppAWSIAMRoleConfig, 0, len(s.cfg.BreakGlass.Roles))
		for _, role := range s.cfg.BreakGlass.Roles {
			breakGlassRoles = append(breakGlassRoles, s.awsIAMRoleToRequest(role))
		}
		req.BreakGlassRoles = breakGlassRoles
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
