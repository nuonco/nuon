package config

import (
	"github.com/invopop/jsonschema"

	"github.com/powertoolsdev/mono/pkg/generics"
)

type SecretsConfig struct {
	Secrets []*AppSecret `mapstructure:"secret,omitempty"`
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
	for _, secret := range a.Secrets {
		if err := secret.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type AppSecret struct {
	Name        string `mapstructure:"name" jsonschema:"required"`
	DisplayName string `mapstructure:"display_name,omitempty"`
	Description string `mapstructure:"description" jsonschema:"required"`

	Required     bool   `mapstructure:"required,omitempty"`
	AutoGenerate bool   `mapstructure:"auto_generate,omitempty"`
	Format       string `mapstructure:"format,omitempty"`
	Default      string `mapstructure:"default,omitempty"`

	// optional fields
	KubernetesSync            bool   `mapstructure:"kubernetes_sync,omitempty"`
	KubernetesSecretNamespace string `mapstructure:"kubernetes_secret_namespace,omitempty"`
	KubernetesSecretName      string `mapstructure:"kubernetes_secret_name,omitempty"`
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
		Example("{{.nuon.install.id}}-secret")
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

	if a.KubernetesSync {
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
