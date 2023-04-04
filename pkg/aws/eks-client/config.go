package eks_client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
)

func (e *eksClient) getConfig(ctx context.Context) (aws.Config, error) {
	if e.RoleARN == "" {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return aws.Config{}, fmt.Errorf("unable to get config: %w", err)
		}

		return cfg, nil
	}

	assumer, err := assumerole.New(e.v,
		assumerole.WithRoleARN(e.RoleARN),
		assumerole.WithRoleSessionName(e.RoleSessionName))
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to get role assumer: %w", err)
	}

	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to assume role: %w", err)
	}

	return cfg, nil
}
