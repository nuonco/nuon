package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s sync) getAppStackRequest() *models.ServiceCreateAppStackConfigRequest {
	req := &models.ServiceCreateAppStackConfigRequest{
		AppConfigID:             generics.ToPtr(s.appConfigID),
		Name:                    generics.ToPtr(s.cfg.Stack.Name),
		Description:             generics.ToPtr(s.cfg.Stack.Description),
		RunnerNestedTemplateURL: s.cfg.Stack.RunnerNestedTemplateURL,
		VpcNestedTemplateURL:    s.cfg.Stack.VPCNestedTemplateURL,
		Type:                    generics.ToPtr(models.AppStackType(s.cfg.Stack.Type)),
	}

	return req
}

func (s sync) syncAppCloudFormationStack(ctx context.Context, resource string) error {
	if s.cfg.Stack == nil {
		return nil
	}

	req := s.getAppStackRequest()
	_, err := s.apiClient.CreateAppStackConfig(ctx, s.appID, req)
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
