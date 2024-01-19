package iam

import (
	"context"
	"fmt"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/powertoolsdev/mono/pkg/generics"
)

// LoadConfigWithAssumedRole loads an AWS config using the default credential provider chain
// to assume the provided role with the provided session name
func (a *assumer) LoadConfigWithAssumedRole(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to get default config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)

	if a.TwoStepRoleARN != "" {
		creds, err := a.assumeIamRole(ctx, stsClient, a.TwoStepRoleARN)
		if err != nil {
			return aws.Config{}, fmt.Errorf("unable to assume two step role: %w", err)
		}

		credsProvider := credentials.NewStaticCredentialsProvider(*creds.AccessKeyId,
			*creds.SecretAccessKey,
			*creds.SessionToken)
		cfg, err = config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credsProvider))
		if err != nil {
			return aws.Config{}, fmt.Errorf("unable to create provider: %w", err)
		}

		stsClient = sts.NewFromConfig(cfg)
	}

	creds, err := a.assumeIamRole(ctx, stsClient, a.RoleARN)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to assume role: %w", err)
	}

	credsProvider := credentials.NewStaticCredentialsProvider(*creds.AccessKeyId, *creds.SecretAccessKey, *creds.SessionToken)
	cfg, err = config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credsProvider))
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to get config with STS creds: %w", err)
	}

	return cfg, nil
}

type awsClientIamRoleAssumer interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

func (a *assumer) assumeIamRole(ctx context.Context, client awsClientIamRoleAssumer, role string) (*sts_types.Credentials, error) {
	params := &sts.AssumeRoleInput{
		RoleArn:         &role,
		RoleSessionName: &a.RoleSessionName,
		DurationSeconds: generics.ToPtr(int32(a.RoleSessionDuration.Seconds())),
	}
	resp, err := client.AssumeRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Credentials, nil
}
