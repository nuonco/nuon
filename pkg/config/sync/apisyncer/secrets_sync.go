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

		// Sync is enabled by the legacy flag or by the presence of any v2 targets.
		KubernetesSync:            secret.KubernetesSyncEnabled(),
		KubernetesSecretName:      secret.KubernetesSecretName,
		KubernetesSecretNamespace: secret.KubernetesSecretNamespace,
		KubernetesSyncTargets:     kubernetesSyncTargetsToRequest(secret.KubernetesSyncTargets),

		Default:      secret.Default,
		Required:     secret.Required,
		AutoGenerate: secret.AutoGenerate,
		Format:       secret.Format,
	}
}

func kubernetesSyncTargetsToRequest(targets []*config.KubernetesSyncTarget) []*models.ServiceKubernetesSyncTarget {
	if len(targets) == 0 {
		return nil
	}

	reqs := make([]*models.ServiceKubernetesSyncTarget, 0, len(targets))
	for _, t := range targets {
		if t == nil {
			continue
		}
		reqs = append(reqs, &models.ServiceKubernetesSyncTarget{
			Namespaces: t.Namespaces,
			Name:       generics.ToPtr(t.Name),
			Key:        generics.ToPtr(t.Key),
		})
	}
	return reqs
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
