package syncer

import (
	"context"

	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// syncAppSecrets creates the app secrets configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_secrets_config.go
func (s *syncer) syncAppSecrets(ctx context.Context) error {
	if s.cfg.Secrets == nil {
		return nil
	}

	secrets := make([]app.AppSecretConfig, 0, len(s.cfg.Secrets.Secrets))
	for _, secret := range s.cfg.Secrets.Secrets {
		secrets = append(secrets, app.AppSecretConfig{
			AppID:                     s.appID,
			AppConfigID:               s.appConfigID,
			Name:                      secret.Name,
			DisplayName:               secret.DisplayName,
			Description:               secret.Description,
			Required:                  secret.Required,
			AutoGenerate:              secret.AutoGenerate,
			Default:                   secret.Default,
			Format:                    app.AppSecretConfigFmt(secret.Format),
			KubernetesSync:            secret.KubernetesSync,
			KubernetesSecretNamespace: secret.KubernetesSecretNamespace,
			KubernetesSecretName:      secret.KubernetesSecretName,
		})
	}

	obj := app.AppSecretsConfig{
		AppID:       s.appID,
		AppConfigID: s.appConfigID,
		Secrets:     secrets,
	}

	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app secrets config",
			Err:         res.Error,
		}
	}

	return nil
}
