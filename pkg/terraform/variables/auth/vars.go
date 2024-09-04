package auth

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

func (v *auth) Init(context.Context) error {
	return nil
}

func (v *auth) GetEnv(ctx context.Context) (map[string]string, error) {
	if v.AWSAuth.UseDefault {
		return map[string]string{}, nil
	}

	envVars, err := credentials.FetchEnv(ctx, v.AWSAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch environment vars: %w", err)
	}

	return envVars, nil
}

func (v *auth) GetFile(context.Context) ([]byte, error) {
	return []byte{}, nil
}
