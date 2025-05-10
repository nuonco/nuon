package config

import (
	"context"

	"github.com/invopop/jsonschema"
)

type SecretsConfig struct {
	Secrets []*AppSecret `mapstructure:"secret,omitempty"`
}

func (a SecretsConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "secret", "Secrets")
}

func (a *SecretsConfig) parse(context.Context) error {
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

	Required     bool `mapstructure:"required,omitempty"`
	AutoGenerate bool `mapstructure:"auto_generate,omitempty"`

	// optional fields
	KubernetesSync            bool   `mapstructure:"kubernetes_sync,omitempty"`
	KubernetesSecretNamespace string `mapstructure:"kubernetes_secret_namespace,omitempty"`
	KubernetesSecretName      string `mapstructure:"kubernetes_secret_name,omitempty"`
}

func (a AppSecret) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "Input name, which is used to reference the secret via variable templating. Supports templating.")
	addDescription(schema, "display_name", "Human readable name of the secret. Supports templating.")
	addDescription(schema, "description", "Description")
	addDescription(schema, "required", "Whether or not the secret is required.")

	addDescription(schema, "kubernetes_secret_name", "The secret name to sync the secret into. Supports templating.")
	addDescription(schema, "kubernetes_secret_namespace", "The namespace to sync the secret into. Supports templating.")
}

func (a *AppSecret) Validate() error {
	if a.AutoGenerate && a.Required {
		return ErrConfig{
			Description: "both auto_generate and required can not be set.",
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

	return nil
}
