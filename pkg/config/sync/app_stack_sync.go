package sync

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/pkg/generics"
)

func (s sync) getAppStackRequest() *models.ServiceCreateAppStackConfigRequest {
	var customNestedStacks []*models.ConfigCustomNestedStack
	for _, stack := range s.cfg.Stack.CustomNestedStacks {
		customNestedStacks = append(customNestedStacks, &models.ConfigCustomNestedStack{
			Name:        stack.Name,
			TemplateURL: stack.TemplateURL,
			Index:       int64(stack.Index),
			Parameters:  stack.Parameters,
			Contents:    stack.Contents,
		})
	}

	return &models.ServiceCreateAppStackConfigRequest{
		AppConfigID:             generics.ToPtr(s.appConfigID),
		Name:                    generics.ToPtr(s.cfg.Stack.Name),
		Description:             generics.ToPtr(s.cfg.Stack.Description),
		RunnerNestedTemplateURL: s.cfg.Stack.RunnerNestedTemplateURL,
		VpcNestedTemplateURL:    s.cfg.Stack.VPCNestedTemplateURL,
		Type:                    generics.ToPtr(models.AppStackType(s.cfg.Stack.Type)),
		CustomNestedStacks:      customNestedStacks,
	}
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
