package apisyncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *syncer) secretToRequest(secret *config.AppSecret) *models.ServiceAppSecretConfig {
	return &models.ServiceAppSecretConfig{
		Description: generics.ToPtr(secret.Description),
		DisplayName: generics.ToPtr(secret.DisplayName),
		Name:        generics.ToPtr(secret.Name),

		KubernetesSync:            secret.KubernetesSync,
		KubernetesSecretName:      secret.KubernetesSecretName,
		KubernetesSecretNamespace: secret.KubernetesSecretNamespace,

		Default:      secret.Default,
		Required:     secret.Required,
		AutoGenerate: secret.AutoGenerate,
		Format:       secret.Format,
	}
}

func (s *syncer) getAppSecretsRequest() *models.ServiceCreateAppSecretsConfigRequest {
	req := &models.ServiceCreateAppSecretsConfigRequest{
		AppConfigID: generics.ToPtr(s.appConfigID),
	}

	secrets := make([]*models.ServiceAppSecretConfig, 0)
	for _, role := range s.cfg.Secrets.Secrets {
		secrets = append(secrets, s.secretToRequest(role))
	}
	req.Secrets = secrets

	return req
}

func (s *syncer) syncAppSecrets(ctx context.Context, resource string) error {
	if s.cfg.Secrets == nil {
		return nil
	}

	req := s.getAppSecretsRequest()
	_, err := s.apiClient.CreateAppSecretsConfig(ctx, s.appID, req)
	if err != nil {
		return sync.SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
