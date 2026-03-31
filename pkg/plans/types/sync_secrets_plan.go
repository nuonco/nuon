package plantypes

import (
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/kube"
)

type KubernetesSecretSync struct {
	SecretARN     string `json:"secret_arn"`
	GCPSecretName string `json:"gcp_secret_name"` // projects/{project}/secrets/{id}/versions/latest
	SecretName    string `json:"secret_name"`     // the name of the secret from the config

	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	KeyName   string `json:"key_name"`

	// NOTE(jm): this should probably come from the app config, but for now we just use string parsing to avoid
	// updating the runner job and save time.
	Format string `json:"format"`
}

type SyncSecretsPlan struct {
	KubernetesSecrets []KubernetesSecretSync `json:"kubernetes_secrets"`

	ClusterInfo *kube.ClusterInfo        `json:"cluster_info,block"`
	AWSAuth     *awscredentials.Config   `json:"aws_auth"`
	AzureAuth   *azurecredentials.Config `json:"azure_auth"`
	GCPAuth     *gcpcredentials.Config   `json:"gcp_auth"`

	MinSandboxMode
}
