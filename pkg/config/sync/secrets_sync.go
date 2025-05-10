package sync

import (
	"context"

	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s sync) secretToRequest(secret *config.AppSecret) *models.ServiceAppSecretConfig {
	return &models.ServiceAppSecretConfig{
		Description: generics.ToPtr(secret.Description),
		DisplayName: generics.ToPtr(secret.DisplayName),
		Name:        generics.ToPtr(secret.Name),

		KubernetesSync:            secret.KubernetesSync,
		KubernetesSecretName:      secret.KubernetesSecretName,
		KubernetesSecretNamespace: secret.KubernetesSecretNamespace,

		Required:     secret.Required,
		AutoGenerate: secret.AutoGenerate,
	}
}

func (s sync) getAppSecretsRequest() *models.ServiceCreateAppSecretsConfigRequest {
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

func (s sync) syncAppSecrets(ctx context.Context, resource string) error {
	if s.cfg.Secrets == nil {
		return nil
	}

	req := s.getAppSecretsRequest()
	_, err := s.apiClient.CreateAppSecretsConfig(ctx, s.appID, req)
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return nil
}
