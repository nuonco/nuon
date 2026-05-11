package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
)

func (s *syncer) kubernetesContextToRequest(ctx *config.KubernetesContext) *models.ServiceAppKubernetesContext {
	return &models.ServiceAppKubernetesContext{
		Name:      generics.ToPtr(ctx.Name),
		Component: generics.ToPtr(ctx.Component),
	}
}

func (s *syncer) getAppKubernetesContextsRequest() *models.ServiceCreateAppKubernetesContextsConfigRequest {
	req := &models.ServiceCreateAppKubernetesContextsConfigRequest{
		AppConfigID: generics.ToPtr(s.appConfigID),
	}

	contexts := make([]*models.ServiceAppKubernetesContext, 0, len(s.cfg.KubernetesContexts.Contexts))
	for _, c := range s.cfg.KubernetesContexts.Contexts {
		contexts = append(contexts, s.kubernetesContextToRequest(c))
	}
	req.Contexts = contexts

	return req
}

func (s *syncer) syncAppKubernetesContexts(ctx context.Context, resource string) error {
	if s.cfg.KubernetesContexts == nil || len(s.cfg.KubernetesContexts.Contexts) == 0 {
		return nil
	}

	req := s.getAppKubernetesContextsRequest()
	_, err := s.apiClient.CreateAppKubernetesContextsConfig(ctx, s.appID, req)
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
