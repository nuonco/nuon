package kube

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

func GetKubeConfig() (*rest.Config, error) {
	if cfg, err := rest.InClusterConfig(); err == nil && cfg != nil {
		return cfg, nil
	}
	home := filepath.Join(homedir.HomeDir(), ".kube", "config")
	if cfg, err := clientcmd.BuildConfigFromFlags("", home); err == nil && cfg != nil {
		return cfg, nil
	}

	return nil, fmt.Errorf("failed to create k8s config")
}

type ClusterInfo struct {
	// ID is the ID of the EKS cluster
	ID string `json:"id"`
	// Endpoint is the URL of the k8s api server
	Endpoint string `json:"endpoint"`
	// CAData is the base64 encoded public certificate
	CAData string `json:"ca_data"`

	// TrustedRoleARN is the arn of the role that should be assumed to interact with the cluster
	TrustedRoleARN string            `json:"trusted_role_arn"`
	EnvVars        map[string]string `json:"env_vars"`

	// KubeConfig will override the kube config, and be parsed instead of generating a new one
	KubeConfig string `json:"kube_config" faker:"-"`
}

func ConfigForCluster(cInfo *ClusterInfo) (*rest.Config, error) {
	if cInfo.KubeConfig != "" {
		config, err := clientcmd.RESTConfigFromKubeConfig([]byte(cInfo.KubeConfig))
		if err != nil {
			return nil, fmt.Errorf("unable to parse kube config: %w", err)
		}

		return config, nil
	}

	if err := validateClusterInfo(cInfo); err != nil {
		return nil, err
	}

	u, err := url.Parse(cInfo.Endpoint)
	if err != nil {
		return nil, err
	}

	envVars := make([]clientcmdapi.ExecEnvVar, 0)
	for k, v := range cInfo.EnvVars {
		envVars = append(envVars, clientcmdapi.ExecEnvVar{
			Name:  k,
			Value: v,
		})
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
	if cInfo.TrustedRoleARN != "" {
		cfg.ExecProvider.Args = []string{"token", "-i", cInfo.ID, "-r", cInfo.TrustedRoleARN}
	}

	return cfg, nil
}

var (
	ErrInvalidCert        error = fmt.Errorf("invalid certificate data")
	ErrInvalidCluster     error = fmt.Errorf("invalid cluster")
	ErrInvalidCredentials error = fmt.Errorf("invalid credentials")
)

func validateClusterInfo(c *ClusterInfo) error {
	if c.Endpoint == "" {
		return fmt.Errorf("%w: empty endpoint", ErrInvalidCluster)
	}

	if c.ID == "" {
		return fmt.Errorf("%w: empty ID", ErrInvalidCluster)
	}

	if c.TrustedRoleARN == "" {
		return fmt.Errorf("%w: missing role ARN", ErrInvalidCredentials)
	}

	if c.CAData == "" {
		return fmt.Errorf("%w: empty certificate data", ErrInvalidCert)
	}

	caDec, err := base64.StdEncoding.DecodeString(c.CAData)
	if err != nil {
		return fmt.Errorf("%w: error decoding certificate data: %v ", ErrInvalidCert, err)
	}
	c.CAData = string(caDec)

	return nil
}
