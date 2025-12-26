package config

import (
	"github.com/invopop/jsonschema"
)

type KubernetesManifestComponentConfig struct {
	Manifest      string  `mapstructure:"manifest,omitempty" jsonschema:"required" features:"get,template"`
	Namespace     string  `mapstructure:"namespace,omitempty" jsonschema:"required"`
	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`
}

func (k KubernetesManifestComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("manifest").Short("Kubernetes manifest").Required().
		Long("YAML manifest content for Kubernetes resources. Supports templating with variables like {{.nuon.install.id}}").
		Example("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: app-config\ndata:\n  env: production").
		Field("namespace").Short("Kubernetes namespace").Required().
		Long("Kubernetes namespace where the manifest will be deployed").
		Example("default").
		Example("production").
		Example("{{.nuon.install.id}}").
		Field("drift_schedule").Short("drift detection schedule").
		Long("Cron expression for periodic drift detection. If not set, drift detection is disabled.").Example("0 2 * * *")
}

func (t *KubernetesManifestComponentConfig) Validate() error {
	// validate if kuebrnetes manifest is a valid config or not, return if not
	return nil
}

func (t *KubernetesManifestComponentConfig) Parse() error {
	return nil
}
