package config

import (
	"github.com/invopop/jsonschema"
)

func ToTerraformVarsMap(inp []TerraformVariable) map[string]string {
	v := make(map[string]string, 0)
	for _, kv := range inp {
		v[kv.Name] = kv.Value
	}

	return v
}

func TerraformVariables(input []TerraformVariable) []string {
	vals := make([]string, 0)
	for _, inp := range input {
		vals = append(vals, inp.Value)
	}

	return vals
}

type TerraformVariable struct {
	Name  string `mapstructure:"name,omitempty" toml:"name"`
	Value string `mapstructure:"value,omitempty" toml:"value"`
}

func (t TerraformVariable) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("terraform variable name").Examples("cluster_name", "install_id", "region").
		Field("value").Short("terraform variable value").Examples("{{.nuon.install.id}}", "true", "us-west-2")
}

type EnvironmentVariable struct {
	Name  string `mapstructure:"name,omitempty" toml:"name"`
	Value string `mapstructure:"value,omitempty" toml:"value"`
}

func (e EnvironmentVariable) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("environment variable name").Examples("NAMESPACE", "INGRESS_NAME", "LOG_LEVEL").
		Field("value").Short("environment variable value").Examples("kube-system", "{{.nuon.install.id}}-public", "info")
}
