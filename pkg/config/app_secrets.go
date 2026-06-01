package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/generics"
)

// secretNameValidator validates Kubernetes names with the same hostname_rfc1123 rule the API enforces, so the CLI
// surfaces the same errors before syncing.
var secretNameValidator = validator.New()

// validKubernetesName reports whether val is a valid RFC 1123 hostname (the rule used for Kubernetes secret names and
// namespaces). Templated values are skipped — they are validated server-side after rendering, since the unrendered
// template (e.g. "{{.nuon.install.id}}-ns") is not itself a valid hostname.
func validKubernetesName(val string) bool {
	if strings.Contains(val, "{{") {
		return true
	}
	return secretNameValidator.Var(val, "hostname_rfc1123") == nil
}

type SecretsConfig struct {
	Secrets []*AppSecret `mapstructure:"secret,omitempty" toml:"secret,omitempty"`
}

func (a SecretsConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("secret").Short("list of secrets").
		Long("Array of secret definitions that customers can provide during installation or that are auto-generated")
}

func (a *SecretsConfig) parse() error {
	return nil
}

func (a *SecretsConfig) Validate() error {
	// seen maps a Kubernetes destination (namespace/name/key) to the Nuon secret that targets it, so we can reject two
	// secrets writing the same key of the same Kubernetes secret (which would clobber one another). Comparison uses the
	// raw, unrendered values; collisions that only differ after templating are not detected here.
	seen := make(map[string]string)
	var warnings []string

	for _, secret := range a.Secrets {
		if err := secret.Validate(); err != nil {
			return err
		}

		for _, dst := range secret.kubernetesDestinations() {
			if owner, ok := seen[dst]; ok {
				return ErrConfig{
					Description: "secrets '" + owner + "' and '" + secret.Name + "' both target the same kubernetes secret key (" + dst + "); each kubernetes secret key may only be targeted by one secret",
				}
			}
			seen[dst] = secret.Name
		}

		if secret.kubernetesSyncExplicitlyDisabled() {
			warnings = append(warnings, fmt.Sprintf("[secrets:%s] kubernetes_sync is set to false but kubernetes_sync_targets are defined; sync will be enabled because targets take precedence", secret.Name))
		}
	}

	// Warnings are only surfaced once all hard validation has passed, so a real error always takes precedence. Each
	// warning is on its own line so multiple warnings render legibly.
	if len(warnings) > 0 {
		return ErrConfig{
			Description: strings.Join(warnings, "\n"),
			Warning:     true,
		}
	}

	return nil
}

type AppSecret struct {
	Name        string `mapstructure:"name" toml:"name" jsonschema:"required"`
	DisplayName string `mapstructure:"display_name,omitempty" toml:"display_name,omitempty"`
	Description string `mapstructure:"description" toml:"description" jsonschema:"required"`

	Required     bool   `mapstructure:"required,omitempty" toml:"required,omitempty"`
	AutoGenerate bool   `mapstructure:"auto_generate,omitempty" toml:"auto_generate,omitempty"`
	Format       string `mapstructure:"format,omitempty" toml:"format,omitempty"`
	Default      string `mapstructure:"default,omitempty" toml:"default,omitempty"`

	// optional fields. KubernetesSync is a pointer so we can distinguish "omitted" (nil) from an explicit
	// "kubernetes_sync = false", which lets us warn when sync is explicitly disabled but v2 targets are present.
	KubernetesSync            *bool  `mapstructure:"kubernetes_sync,omitempty" toml:"kubernetes_sync,omitempty"`
	KubernetesSecretNamespace string `mapstructure:"kubernetes_secret_namespace,omitempty" toml:"kubernetes_secret_namespace,omitempty"`
	KubernetesSecretName      string `mapstructure:"kubernetes_secret_name,omitempty" toml:"kubernetes_secret_name,omitempty"`

	// kubernetes secrets v2: a secret may target multiple Kubernetes destinations, each with its own namespace(s),
	// secret name, and key. When present, sync is implied. The single-valued kubernetes_secret_* fields above remain
	// supported for backwards compatibility.
	KubernetesSyncTargets []*KubernetesSyncTarget `mapstructure:"kubernetes_sync_targets,omitempty" toml:"kubernetes_sync_targets,omitempty"`
}

// KubernetesSyncTarget describes a single Kubernetes destination for a secret: the secret name and key written into
// each of the listed namespaces.
type KubernetesSyncTarget struct {
	Namespaces []string `mapstructure:"namespaces" toml:"namespaces" jsonschema:"required"`
	Name       string   `mapstructure:"name" toml:"name" jsonschema:"required"`
	Key        string   `mapstructure:"key" toml:"key" jsonschema:"required"`
}

func (t KubernetesSyncTarget) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("namespaces").Short("target namespaces").Required().
		Long("List of Kubernetes namespaces the secret will be created in. Supports templating").
		Example("datadog").
		Field("name").Short("Kubernetes secret name").Required().
		Long("Name of the Kubernetes Secret resource. Supports templating").
		Example("datadog-api-key").
		Field("key").Short("Kubernetes secret key").Required().
		Long("Key within the Kubernetes Secret to write the value to. Supports templating").
		Example("api-key")
}

func (a AppSecret) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("secret name").Required().
		Long("Identifier for the secret used to reference it via variable templating (e.g., {{.nuon.secrets.name}}). Supports templating").
		Example("database_password").
		Example("api_key").
		Field("display_name").Short("display name").
		Long("Human-readable name shown in the installer UI. Supports templating").
		Example("Database Password").
		Example("API Key").
		Field("description").Short("secret description").Required().
		Long("Detailed explanation of what this secret is for, displayed to users during installation").
		Example("Master password for the database").
		Example("API key for external service authentication").
		Field("required").Short("whether secret is required").
		Long("If true, customer must provide a value during installation. If false, can be skipped").
		Field("auto_generate").Short("whether to auto-generate secret").
		Long("If true, a random secret will be generated if customer does not provide one. Cannot be used with required or default").
		Field("format").Short("secret format").
		Long("Format of the secret value. Supported values: 'base64' for base64-encoded secrets, or empty for plain text").
		Example("base64").
		Field("default").Short("default value").
		Long("Default value used if customer does not provide one. Cannot be used with required or auto_generate").
		Field("kubernetes_sync").Short("sync to Kubernetes").
		Long("If true, the secret will be synced to a Kubernetes Secret resource").
		Field("kubernetes_secret_namespace").Short("Kubernetes namespace").
		Long("Kubernetes namespace where the secret will be created. Required if kubernetes_sync is true. Supports templating").
		Example("default").
		Example("{{.nuon.install.id}}-namespace").
		Field("kubernetes_secret_name").Short("Kubernetes secret name").
		Long("Name of the Kubernetes Secret resource. Required if kubernetes_sync is true. Supports templating").
		Example("app-secret").
		Example("{{.nuon.install.id}}-secret").
		Field("kubernetes_sync_targets").Short("Kubernetes sync targets").
		Long("List of Kubernetes destinations to sync this secret to. When present, Kubernetes sync is enabled. Each target writes the secret value into a given key of a named secret across one or more namespaces, and supersedes the single-valued kubernetes_secret_* fields")
}

func (a *AppSecret) Validate() error {
	if a.AutoGenerate && a.Required {
		return ErrConfig{
			Description: "both auto_generate and required can not be set.",
		}
	}

	if a.Default != "" && a.Required {
		return ErrConfig{
			Description: "can not have both required and default set.",
		}
	}
	if a.Default != "" && a.AutoGenerate {
		return ErrConfig{
			Description: "can not have both auto_generate and default set.",
		}
	}

	// Legacy single-target requirement only applies when no v2 targets are configured. When targets are present they
	// supersede the single-valued kubernetes_secret_* fields.
	if a.KubernetesSync != nil && *a.KubernetesSync && len(a.KubernetesSyncTargets) == 0 {
		if a.KubernetesSecretName == "" {
			return ErrConfig{
				Description: "kubernetes_secret_name must be set when kubernetes_sync is true",
			}
		}

		if a.KubernetesSecretNamespace == "" {
			return ErrConfig{
				Description: "kubernetes_secret_namespace must be set when kubernetes_sync is true",
			}
		}
	}

	for _, target := range a.KubernetesSyncTargets {
		if target == nil {
			continue
		}
		if len(target.Namespaces) == 0 {
			return ErrConfig{
				Description: "each kubernetes_sync_targets entry must set at least one namespace",
			}
		}
		if target.Name == "" {
			return ErrConfig{
				Description: "each kubernetes_sync_targets entry must set name",
			}
		}
		if target.Key == "" {
			return ErrConfig{
				Description: "each kubernetes_sync_targets entry must set key",
			}
		}

		if !validKubernetesName(target.Name) {
			return ErrConfig{
				Description: "kubernetes_sync_targets name '" + target.Name + "' must be a valid RFC 1123 hostname (lowercase alphanumeric, '-' and '.')",
			}
		}
		for _, ns := range target.Namespaces {
			if !validKubernetesName(ns) {
				return ErrConfig{
					Description: "kubernetes_sync_targets namespace '" + ns + "' must be a valid RFC 1123 hostname (lowercase alphanumeric, '-' and '.')",
				}
			}
		}
	}

	if !generics.SliceContains(a.Format, []string{
		"base64",
		"",
	}) {
		return ErrConfig{
			Description: "Invalid format " + a.Format,
		}
	}

	return nil
}

// KubernetesSyncEnabled reports whether the secret should be synced to Kubernetes. Sync is enabled either by the legacy
// kubernetes_sync flag or by the presence of one or more kubernetes_sync_targets.
func (a *AppSecret) KubernetesSyncEnabled() bool {
	return (a.KubernetesSync != nil && *a.KubernetesSync) || len(a.KubernetesSyncTargets) > 0
}

// kubernetesSyncExplicitlyDisabled reports whether the secret sets kubernetes_sync = false while also defining v2
// targets. This is a contradiction we warn (but do not error) on: targets take precedence and sync stays enabled.
func (a *AppSecret) kubernetesSyncExplicitlyDisabled() bool {
	return a.KubernetesSync != nil && !*a.KubernetesSync && len(a.KubernetesSyncTargets) > 0
}

// kubernetesDestinations returns the set of Kubernetes destinations this secret writes to, each as a
// "namespace/name/key" string. v2 targets fan out across their namespaces; a legacy single-target sync contributes its
// namespace/name with the implicit "value" key.
func (a *AppSecret) kubernetesDestinations() []string {
	dsts := make([]string, 0)

	for _, target := range a.KubernetesSyncTargets {
		if target == nil {
			continue
		}
		for _, ns := range target.Namespaces {
			dsts = append(dsts, ns+"/"+target.Name+"/"+target.Key)
		}
	}

	if a.KubernetesSync != nil && *a.KubernetesSync && len(a.KubernetesSyncTargets) == 0 {
		dsts = append(dsts, a.KubernetesSecretNamespace+"/"+a.KubernetesSecretName+"/value")
	}

	return dsts
}
