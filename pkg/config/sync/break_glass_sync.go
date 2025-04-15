package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s sync) getAppBreakGlassRequest() *models.ServiceCreateAppBreakGlassConfigRequest {
	req := &models.ServiceCreateAppBreakGlassConfigRequest{
		AppConfigID: generics.ToPtr(s.appConfigID),
	}

	roles := make([]*models.ServiceAppAWSIAMRoleConfig, 0)
	for _, role := range s.cfg.BreakGlass.Roles {
		roles = append(roles, s.awsIAMRoleToRequest(role))
	}

	req.Roles = roles
	return req
}

func (s sync) syncAppBreakGlass(ctx context.Context, resource string) error {
	if s.cfg.BreakGlass == nil {
		return nil
	}

	req := s.getAppBreakGlassRequest()
	_, err := s.apiClient.CreateAppBreakGlassConfig(ctx, s.appID, req)
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
