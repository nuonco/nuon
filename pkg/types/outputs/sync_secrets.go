package outputs

import "time"

type SyncSecretsOutput map[string]SecretSyncOutput

type SecretSyncOutput struct {
	Name string `mapstructure:"name" json:"name,omitempty"`
	ARN  string `mapstructure:"arn" json:"arn,omitempty"`

	KubernetesNamespace string `mapstructure:"kubernetes_namespace" json:"kubernetes_namespace,omitempty"`
	KubernetesName      string `mapstructure:"kubernetes_name" json:"kubernetes_name,omitempty"`
	KubernetesKey       string `mapstructure:"kubernetes_key" json:"kubernetes_key,omitempty"`

	Timestamp *time.Time `mapstructure:"timestamp" json:"timestamp,omitempty"`
	Length    int        `mapstructure:"length" json:"length,omitzero"`
	Exists    bool       `mapstructure:"exists" json:"exists"`
}
