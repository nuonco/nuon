package secrets

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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
			AppID:        appID,
			AppConfigID:  appConfigID,
			Name:         secret.Name,
			DisplayName:  secret.DisplayName,
			Description:  secret.Description,
			Required:     secret.Required,
			AutoGenerate: secret.AutoGenerate,
			Default:      secret.Default,
			Format:       app.AppSecretConfigFmt(secret.Format),
			// Sync is enabled by the legacy flag or by the presence of any v2 targets.
			KubernetesSync:            secret.KubernetesSyncEnabled(),
			KubernetesSecretNamespace: secret.KubernetesSecretNamespace,
			KubernetesSecretName:      secret.KubernetesSecretName,
			KubernetesSyncTargets:     kubernetesSyncTargets(secret.KubernetesSyncTargets),
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

// kubernetesSyncTargets maps config sync targets onto the app model child rows.
func kubernetesSyncTargets(targets []*config.KubernetesSyncTarget) []app.AppSecretKubernetesSyncTarget {
	if len(targets) == 0 {
		return nil
	}

	rows := make([]app.AppSecretKubernetesSyncTarget, 0, len(targets))
	for _, t := range targets {
		if t == nil {
			continue
		}
		rows = append(rows, app.AppSecretKubernetesSyncTarget{
			Namespaces: t.Namespaces,
			Name:       t.Name,
			Key:        t.Key,
		})
	}
	return rows
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

		// Validate v2 sync targets.
		for _, target := range secret.KubernetesSyncTargets {
			if target == nil {
				continue
			}

			if len(target.Namespaces) == 0 {
				return stderr.ErrUser{
					Err:         fmt.Errorf("kubernetes_sync_targets entry missing namespace"),
					Description: fmt.Sprintf("Secret '%s' has a kubernetes_sync_targets entry with no namespace", secret.Name),
				}
			}

			if target.Name == "" {
				return stderr.ErrUser{
					Err:         fmt.Errorf("kubernetes_sync_targets entry missing name"),
					Description: fmt.Sprintf("Secret '%s' has a kubernetes_sync_targets entry with no name", secret.Name),
				}
			}

			if target.Key == "" {
				return stderr.ErrUser{
					Err:         fmt.Errorf("kubernetes_sync_targets entry missing key"),
					Description: fmt.Sprintf("Secret '%s' has a kubernetes_sync_targets entry with no key", secret.Name),
				}
			}

			// Templated values (e.g. "{{.nuon.install.id}}-ns") are validated after rendering, so skip them here.
			if !strings.Contains(target.Name, "{{") && !dnsRFC1123Regex.MatchString(target.Name) {
				return stderr.ErrUser{
					Err:         fmt.Errorf("invalid kubernetes_sync_targets name: %s", target.Name),
					Description: fmt.Sprintf("Secret '%s' has invalid kubernetes_sync_targets name '%s'. Must be a valid DNS RFC 1123 subdomain (lowercase alphanumeric, hyphens, dots)", secret.Name, target.Name),
				}
			}

			for _, ns := range target.Namespaces {
				if strings.Contains(ns, "{{") {
					continue
				}
				if !dnsRFC1123Regex.MatchString(ns) {
					return stderr.ErrUser{
						Err:         fmt.Errorf("invalid kubernetes_sync_targets namespace: %s", ns),
						Description: fmt.Sprintf("Secret '%s' has invalid kubernetes_sync_targets namespace '%s'. Must be a valid DNS RFC 1123 subdomain (lowercase alphanumeric, hyphens, dots)", secret.Name, ns),
					}
				}
			}
		}
	}

	return nil
}
