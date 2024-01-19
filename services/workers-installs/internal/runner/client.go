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

func (a *Activities) getEFSClient(ctx context.Context, iamRoleARN, region string) (*efs.Client, error) {
	cfg, err := credentials.Fetch(ctx, &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:        iamRoleARN,
			SessionName:    defaultSessionName,
			TwoStepRoleARN: a.cfg.NuonAccessRoleArn,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create credential config: %w", err)
	}

	cfg.Region = region

	svc := efs.NewFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create credentials: %w", err)
	}

	return svc, nil
}

func (a *Activities) getECSClient(ctx context.Context, iamRoleARN, region string) (*ecs.Client, error) {
	cfg, err := credentials.Fetch(ctx, &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:        iamRoleARN,
			SessionName:    defaultSessionName,
			TwoStepRoleARN: a.cfg.NuonAccessRoleArn,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create credential config: %w", err)
	}

	cfg.Region = region

	svc := ecs.NewFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create credentials: %w", err)
	}

	return svc, nil
}

func (a *Activities) getCloudwatchClient(ctx context.Context, iamRoleARN, region string) (*cloudwatchlogs.Client, error) {
	cfg, err := credentials.Fetch(ctx, &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:        iamRoleARN,
			SessionName:    defaultSessionName,
			TwoStepRoleARN: a.cfg.NuonAccessRoleArn,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create credential config: %w", err)
	}
	cfg.Region = region

	svc := cloudwatchlogs.NewFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create credentials: %w", err)
	}

	return svc, nil
}
