package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *syncer) getAppBreakGlassRequest() *models.ServiceCreateAppBreakGlassConfigRequest {
	req := &models.ServiceCreateAppBreakGlassConfigRequest{
		AppConfigID: generics.ToPtr(s.appConfigID),
	}

	roles := make([]*models.ServiceAppAWSIAMRoleConfig, 0)
	for _, role := range s.cfg.BreakGlass.Roles {
		roles = append(roles, s.iamRoleToRequest(role))
	}

	req.Roles = roles
	return req
}

func (s *syncer) syncAppBreakGlass(ctx context.Context, resource string) error {
	if s.cfg.BreakGlass == nil {
		return nil
	}

	req := s.getAppBreakGlassRequest()
	_, err := s.apiClient.CreateAppBreakGlassConfig(ctx, s.appID, req)
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
