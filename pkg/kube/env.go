package kube

import (
	"context"
	"fmt"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (c *ClusterInfo) fetchEnv(ctx context.Context) ([]clientcmdapi.ExecEnvVar, error) {
	envVars := c.EnvVars
	if c.AWSAuth != nil {
		credEnvVars, err := awscredentials.FetchEnv(ctx, c.AWSAuth)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch credentials for aws auth: %w", err)
		}

		envVars = generics.MergeMap(envVars, credEnvVars)
	}
	if c.AzureAuth != nil {
		credEnvVars, err := azurecredentials.FetchEnv(ctx, c.AzureAuth)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch credentials for aws auth: %w", err)
		}

		envVars = generics.MergeMap(envVars, credEnvVars)
	}

	execEnvVars := make([]clientcmdapi.ExecEnvVar, 0)
	for k, v := range c.EnvVars {
		execEnvVars = append(execEnvVars, clientcmdapi.ExecEnvVar{
			Name:  k,
			Value: v,
		})
	}

	return execEnvVars, nil
}
