package iam

import (
	"context"
	"fmt"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/powertoolsdev/go-generics"
)

type iamRoleAssumer interface {
	loadConfigWithAssumeRole(context.Context, string) (aws.Config, error)
	assumeIamRole(context.Context, string, awsClientIamRoleAssumer) (*sts_types.Credentials, error)
}

type iamRoleAssumerImpl struct{}

var _ iamRoleAssumer = (*iamRoleAssumerImpl)(nil)

type awsClientIamRoleAssumer interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

func (r *iamRoleAssumerImpl) assumeIamRole(ctx context.Context, roleArn string, client awsClientIamRoleAssumer) (*sts_types.Credentials, error) {
	params := &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: generics.ToPtr("workers-orgs"),
	}
	resp, err := client.AssumeRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Credentials, nil
}

// TODO(jm): break this down and test. This code was mainly imported from various other activities that did
// this manually and probably needs to be redesigned into a more testable fashion.
func (r *iamRoleAssumerImpl) loadConfigWithAssumeRole(ctx context.Context, roleArn string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to get default config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	creds, err := r.assumeIamRole(ctx, roleArn, stsClient)
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
