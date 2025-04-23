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

type AppSecret struct {
	Name        string `mapstructure:"name" jsonschema:"required"`
	DisplayName string `mapstructure:"display_name,omitempty"`
	Description string `mapstructure:"description" jsonschema:"required"`

	Required bool `mapstructure:"required,omitempty"`

	// optional fields

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
