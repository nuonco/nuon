package runner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/efs"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

const (
	defaultSessionName string = "workers-installs"
)

func (a *Activities) getEFSClient(ctx context.Context, region string, auth *credentials.Config) (*efs.Client, error) {
	cfg, err := credentials.Fetch(ctx, auth)
	if err != nil {
		return nil, fmt.Errorf("unable to get efs client credentials: %w", err)
	}
	cfg.Region = region

	svc := efs.NewFromConfig(cfg)
	return svc, nil
}

func (a *Activities) getECSClient(ctx context.Context, region string, auth *credentials.Config) (*ecs.Client, error) {
	cfg, err := credentials.Fetch(ctx, auth)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecs client credentials: %w", err)
	}

	cfg.Region = region

	svc := ecs.NewFromConfig(cfg)
	return svc, nil
}

func (a *Activities) getCloudwatchClient(ctx context.Context, region string, auth *credentials.Config) (*cloudwatchlogs.Client, error) {
	cfg, err := credentials.Fetch(ctx, auth)
	if err != nil {
		return nil, fmt.Errorf("unable to create cloudwatch credentials: %w", err)
	}

	cfg.Region = region

	svc := cloudwatchlogs.NewFromConfig(cfg)
	return svc, nil
}
