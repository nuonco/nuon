package secrets

import (
	"context"
	"fmt"
	"regexp"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

var (
	entityNameRegex = regexp.MustCompile(`^[a-z0-9_-]*$`)
	dnsRFC1123Regex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)
)

// Sync creates the app secrets configuration.
// Duplicates logic from services/ctl-api/internal/app/apps/service/create_app_secrets_config.go
func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, appID, appConfigID string) error {
	if cfg.Secrets == nil {
		return nil
	}

	// Validate secrets
	if err := validateSecrets(cfg); err != nil {
		return sync.SyncErr{
			Resource:    "app-secrets",
			Description: fmt.Sprintf("validation failed: %v", err),
		}
	}

	secrets := make([]app.AppSecretConfig, 0, len(cfg.Secrets.Secrets))
	for _, secret := range cfg.Secrets.Secrets {
		secrets = append(secrets, app.AppSecretConfig{
			AppID:                     appID,
			AppConfigID:               appConfigID,
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
		AppID:       appID,
		AppConfigID: appConfigID,
		Secrets:     secrets,
	}

	res := db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return sync.SyncInternalErr{
			Description: "unable to create app secrets config",
			Err:         res.Error,
		}
	}

	return nil
}

// validateSecrets validates secret names, required fields, and Kubernetes secret names.
// Duplicates validation logic from services/ctl-api/internal/app/apps/service/create_app_secrets_config.go
func validateSecrets(cfg *config.AppConfig) error {
	for _, secret := range cfg.Secrets.Secrets {
		// Validate secret name using entity_name pattern (lowercase, numbers, underscores, hyphens)
		if secret.Name == "" {
			return stderr.ErrUser{
				Err:         fmt.Errorf("secret name is required"),
				Description: "Secret name cannot be empty",
			}
		}

		if !entityNameRegex.MatchString(secret.Name) {
			return stderr.ErrUser{
				Err:         fmt.Errorf("invalid secret name: %s", secret.Name),
				Description: fmt.Sprintf("Secret name '%s' must contain only lowercase letters, numbers, underscores, and hyphens", secret.Name),
			}
		}

		// Validate required fields
		if secret.DisplayName == "" {
			return stderr.ErrUser{
				Err:         fmt.Errorf("secret display_name is required"),
				Description: fmt.Sprintf("Secret '%s' is missing required field 'display_name'", secret.Name),
			}
		}

		if secret.Description == "" {
			return stderr.ErrUser{
				Err:         fmt.Errorf("secret description is required"),
				Description: fmt.Sprintf("Secret '%s' is missing required field 'description'", secret.Name),
			}
		}

		// Validate Kubernetes secret name if provided (DNS RFC 1123 format)
		if secret.KubernetesSecretName != "" {
			if len(secret.KubernetesSecretName) > 253 {
				return stderr.ErrUser{
					Err:         fmt.Errorf("invalid kubernetes_secret_name: %s", secret.KubernetesSecretName),
					Description: fmt.Sprintf("Secret '%s' has kubernetes_secret_name that exceeds 253 characters", secret.Name),
				}
			}

			if !dnsRFC1123Regex.MatchString(secret.KubernetesSecretName) {
				return stderr.ErrUser{
					Err:         fmt.Errorf("invalid kubernetes_secret_name: %s", secret.KubernetesSecretName),
					Description: fmt.Sprintf("Secret '%s' has invalid kubernetes_secret_name '%s'. Must be a valid DNS RFC 1123 subdomain (lowercase alphanumeric, hyphens, dots)", secret.Name, secret.KubernetesSecretName),
				}
			}
		}
	}

	return nil
}
