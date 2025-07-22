package auth

import (
	"context"
	"fmt"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	azurecredentials "github.com/powertoolsdev/mono/pkg/azure/credentials"
	"github.com/powertoolsdev/mono/pkg/terraform/variables"
)

func (v *auth) Init(context.Context) error {
	return nil
}

func (v *auth) GetEnv(ctx context.Context) (map[string]string, error) {
	switch {
	case v.AzureAuth != nil:
		envVars, err := azurecredentials.FetchEnv(ctx, v.AzureAuth)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch environment vars: %w", err)
		}

		return envVars, nil
	case v.AWSAuth != nil:
		if v.AWSAuth.UseDefault {
			return map[string]string{}, nil
		}

		envVars, err := awscredentials.FetchEnv(ctx, v.AWSAuth)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch environment vars: %w", err)
		}
		return envVars, nil
	default:
		return map[string]string{}, nil
	}
}

func (v *auth) GetFiles(context.Context) ([]variables.VarFile, error) {
	return nil, nil
}
