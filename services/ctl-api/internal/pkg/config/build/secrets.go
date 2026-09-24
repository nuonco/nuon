package build

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/validation"
)

type SecretInput struct {
	Name        string
	DisplayName string
	Description string

	Required     bool
	AutoGenerate bool
	Format       string
	Default      string

	KubernetesSync            bool
	KubernetesSecretNamespace string
	KubernetesSecretName      string
	KubernetesSyncTargets     []config.KubernetesSyncTarget
}

// SecretInputsFromConfig enables sync from the legacy flag or any v2 target.
func SecretInputsFromConfig(secrets *config.SecretsConfig) []SecretInput {
	if secrets == nil {
		return nil
	}

	out := make([]SecretInput, 0, len(secrets.Secrets))
	for _, secret := range secrets.Secrets {
		targets := make([]config.KubernetesSyncTarget, 0, len(secret.KubernetesSyncTargets))
		for _, t := range secret.KubernetesSyncTargets {
			if t == nil {
				continue
			}
			targets = append(targets, *t)
		}

		out = append(out, SecretInput{
			Name:                      secret.Name,
			DisplayName:               secret.DisplayName,
			Description:               secret.Description,
			Required:                  secret.Required,
			AutoGenerate:              secret.AutoGenerate,
			Format:                    secret.Format,
			Default:                   secret.Default,
			KubernetesSync:            secret.KubernetesSyncEnabled(),
			KubernetesSecretNamespace: secret.KubernetesSecretNamespace,
			KubernetesSecretName:      secret.KubernetesSecretName,
			KubernetesSyncTargets:     targets,
		})
	}
	return out
}

func SecretsConfig(secrets []SecretInput, appID, appConfigID string) (*app.AppSecretsConfig, error) {
	objs := make([]app.AppSecretConfig, 0, len(secrets))
	for _, secret := range secrets {
		if err := validateSecret(secret); err != nil {
			return nil, err
		}

		targets := make([]app.AppSecretKubernetesSyncTarget, 0, len(secret.KubernetesSyncTargets))
		for _, t := range secret.KubernetesSyncTargets {
			targets = append(targets, app.AppSecretKubernetesSyncTarget{
				Namespaces: t.Namespaces,
				Name:       t.Name,
				Key:        t.Key,
			})
		}

		objs = append(objs, app.AppSecretConfig{
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
			KubernetesSyncTargets:     targets,
		})
	}

	return &app.AppSecretsConfig{
		AppID:       appID,
		AppConfigID: appConfigID,
		Secrets:     objs,
	}, nil
}

func validateSecret(secret SecretInput) error {
	if err := validateSecretName(secret.Name); err != nil {
		return err
	}
	if secret.DisplayName == "" {
		return userErr("secret_display_name_required", fmt.Sprintf("Secret '%s' is missing required field 'display_name'", secret.Name))
	}
	if secret.Description == "" {
		return userErr("secret_description_required", fmt.Sprintf("Secret '%s' is missing required field 'description'", secret.Name))
	}

	if secret.KubernetesSecretName != "" {
		if err := validateDNSSubdomain(secret.Name, "kubernetes_secret_name", secret.KubernetesSecretName); err != nil {
			return err
		}
	}

	for _, target := range secret.KubernetesSyncTargets {
		if len(target.Namespaces) == 0 {
			return userErr("kubernetes_sync_target_namespace_required", fmt.Sprintf("Secret '%s' has a kubernetes_sync_targets entry with no namespace", secret.Name))
		}
		if target.Name == "" {
			return userErr("kubernetes_sync_target_name_required", fmt.Sprintf("Secret '%s' has a kubernetes_sync_targets entry with no name", secret.Name))
		}
		if target.Key == "" {
			return userErr("kubernetes_sync_target_key_required", fmt.Sprintf("Secret '%s' has a kubernetes_sync_targets entry with no key", secret.Name))
		}

		if err := validateDNSSubdomain(secret.Name, "kubernetes_sync_targets name", target.Name); err != nil {
			return err
		}
		for _, ns := range target.Namespaces {
			if err := validateDNSSubdomain(secret.Name, "kubernetes_sync_targets namespace", ns); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateSecretName(name string) error {
	if strings.Contains(name, "{{") {
		return validation.ValidateInterpolatedName(name)
	}
	return validation.ValidateEntityName(name)
}

// validateDNSSubdomain skips templated values, resolvable only after rendering.
func validateDNSSubdomain(secretName, field, value string) error {
	if strings.Contains(value, "{{") {
		return nil
	}
	if err := validation.ValidateDNSSubdomain(value); err != nil {
		return userErr("invalid_"+strings.ReplaceAll(field, " ", "_"), fmt.Sprintf("Secret '%s' has invalid %s '%s': must be a valid DNS RFC 1123 subdomain (lowercase alphanumeric, hyphens, dots) of at most 253 characters", secretName, field, value))
	}
	return nil
}

func userErr(code, description string) error {
	return stderr.ErrUser{
		Err:         fmt.Errorf("%s", code),
		Code:        code,
		Description: description,
	}
}
