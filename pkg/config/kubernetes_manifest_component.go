package config

type KubernetesManifestComponentConfig struct {
	Manifest      string  `mapstructure:"manifest,omitempty" jsonschema:"required"`
	Namespace     string  `mapstructure:"namespace,omitempty" jsonschema:"required"`
	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`
}

func (t *KubernetesManifestComponentConfig) Validate() error {
	// validate if kuebrnetes manifest is a valid config or not, return if not
	return nil
}

func (t *KubernetesManifestComponentConfig) Parse() error {
	return nil
}
