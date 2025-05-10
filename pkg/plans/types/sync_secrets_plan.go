package plantypes

import (
	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
	"github.com/powertoolsdev/mono/pkg/kube"
)

type KubernetesSecretSync struct {
	SecretARN string `json:"secret_arn"`

	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	KeyName   string `json:"key_name"`
}

type SyncSecretsPlan struct {
	KubernetesSecrets []KubernetesSecretSync `json:"kubernetes_secrets"`

	ClusterInfo *kube.ClusterInfo        `json:"cluster_info,block"`
	AzureAuth   *azurecredentials.Config `json:"azure_auth"`
	AWSAuth     *awscredentials.Config   `json:"aws_auth"`
}
