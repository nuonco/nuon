package kube

import (
	"context"
	"fmt"
	"net/url"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
)

type ClusterInfo struct {
	// ID is the ID of the EKS cluster
	ID string `json:"id"`
	// Endpoint is the URL of the k8s api server
	Endpoint string `json:"endpoint"`
	// CAData is the base64 encoded public certificate
	CAData string `json:"ca_data"`

	EnvVars map[string]string `json:"env_vars"`

	// KubeConfig will override the kube config, and be parsed instead of generating a new one
	KubeConfig string `json:"kube_config" faker:"-"`

	// If either an AWS auth or Azure auth is passed in, we will automatically use it to resolve credentials and set
	// them in the environment.
	AWSAuth   *awscredentials.Config   `json:"aws_auth"`
	AzureAuth *azurecredentials.Config `json:"azure_auth"`

	// TrustedRoleARN is the arn of the role that should be assumed to interact with the cluster
	// NOTE(JM): we are deprecating this
	TrustedRoleARN string `json:"trusted_role_arn"`
}

func ConfigForCluster(ctx context.Context, cInfo *ClusterInfo) (*rest.Config, error) {
	if cInfo.KubeConfig != "" {
		config, err := clientcmd.RESTConfigFromKubeConfig([]byte(cInfo.KubeConfig))
		if err != nil {
			return nil, fmt.Errorf("unable to parse kube config: %w", err)
		}

		return config, nil
	}

	u, err := url.Parse(cInfo.Endpoint)
	if err != nil {
		return nil, err
	}

	envVars, err := cInfo.fetchEnv(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch environment: %w", err)
	}

	cfg := &rest.Config{
		Host: cInfo.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			ServerName: u.Hostname(),
			CAData:     []byte(cInfo.CAData),
		},
		ExecProvider: &clientcmdapi.ExecConfig{
			APIVersion:      "client.authentication.k8s.io/v1beta1",
			Command:         "aws-iam-authenticator",
			Env:             envVars,
			Args:            []string{"token", "-i", cInfo.ID},
			InteractiveMode: clientcmdapi.NeverExecInteractiveMode,
		},
	}
	// TODO(jm): this is deprecated and only used in legacy users of this
	if cInfo.TrustedRoleARN != "" {
		cfg.ExecProvider.Args = []string{"token", "-i", cInfo.ID, "-r", cInfo.TrustedRoleARN}
	}

	return cfg, nil
}
