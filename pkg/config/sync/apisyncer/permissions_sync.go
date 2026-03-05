package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *syncer) getAppPermissionsRequest() *models.ServiceCreateAppPermissionsConfigRequest {
	req := &models.ServiceCreateAppPermissionsConfigRequest{
		AppConfigID: generics.ToPtr(s.appConfigID),
	}

	if s.cfg.Permissions.DeprovisionRole != nil {
		req.DeprovisionRole = s.iamRoleToRequest(s.cfg.Permissions.DeprovisionRole)
	}

	if s.cfg.Permissions.ProvisionRole != nil {
		req.ProvisionRole = s.iamRoleToRequest(s.cfg.Permissions.ProvisionRole)
	}

	if s.cfg.Permissions.MaintenanceRole != nil {
		req.MaintenanceRole = s.iamRoleToRequest(s.cfg.Permissions.MaintenanceRole)
	}

	if s.cfg.BreakGlass != nil && len(s.cfg.BreakGlass.Roles) > 0 {
		breakGlassRoles := make([]*models.ServiceAppAWSIAMRoleConfig, 0, len(s.cfg.BreakGlass.Roles))
		for _, role := range s.cfg.BreakGlass.Roles {
			breakGlassRoles = append(breakGlassRoles, s.iamRoleToRequest(role))
		}
		req.BreakGlassRoles = breakGlassRoles
	}

	if len(s.cfg.Permissions.CustomRoles) > 0 {
		customRoles := make([]*models.ServiceAppAWSIAMRoleConfig, 0, len(s.cfg.Permissions.CustomRoles))
		for _, role := range s.cfg.Permissions.CustomRoles {
			customRoles = append(customRoles, s.iamRoleToRequest(role))
		}
		req.CustomRoles = customRoles
	}

	return req
}

func (s *syncer) syncAppPermissions(ctx context.Context, resource string) error {
	if s.cfg.Permissions == nil {
		return nil
	}

	req := s.getAppPermissionsRequest()
	_, err := s.apiClient.CreateAppPermissionsConfig(ctx, s.appID, req)
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
