package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s sync) getCloudFormationStackRequest() *models.ServiceCreateAppCloudFormationStackConfigRequest {
	req := &models.ServiceCreateAppCloudFormationStackConfigRequest{
		AppConfigID:             s.appConfigID,
		Name:                    generics.ToPtr(s.cfg.CloudFormationStack.Name),
		Description:             generics.ToPtr(s.cfg.CloudFormationStack.Description),
		RunnerNestedTemplateURL: s.cfg.CloudFormationStack.RunnerNestedTemplateURL,
		VpcNestedTemplateURL:    s.cfg.CloudFormationStack.VPCNestedTemplateURL,
	}

	return req
}

func (s sync) syncAppCloudFormationStack(ctx context.Context, resource string) error {
	if s.cfg.CloudFormationStack == nil {
		return nil
	}

	req := s.getCloudFormationStackRequest()
	_, err := s.apiClient.CreateAppCloudFormationStackConfig(ctx, s.appID, req)
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
