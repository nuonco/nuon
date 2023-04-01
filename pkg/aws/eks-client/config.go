package eks_client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
)

func (e *eksClient) getConfig(ctx context.Context) (aws.Config, error) {
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
