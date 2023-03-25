package iam

import (
	"context"
	"fmt"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/go-playground/validator/v10"
)

type assumer struct {
	RoleARN         string `validate:"required"`
	RoleSessionName string `validate:"required"`

	// internal state
	validator *validator.Validate
}

type assumerOptions func(*assumer) error

// New creates a new, validated assumer with the given options
func New(v *validator.Validate, opts ...assumerOptions) (*assumer, error) {
	a := &assumer{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating assumer: validator is nil")
	}
	a.validator = v

	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, err
		}
	}
	if err := a.validator.Struct(a); err != nil {
		return nil, err
	}
	return a, nil
}

// WithRoleARN sets the ARN of the role to assume
func WithRoleARN(s string) assumerOptions {
	return func(a *assumer) error {
		a.RoleARN = s
		return nil
	}
}

// WithRoleSessionName specifies the session name to use when assuming the role
func WithRoleSessionName(s string) assumerOptions {
	return func(a *assumer) error {
		a.RoleSessionName = s
		return nil
	}
}

// LoadConfigWithAssumedRole loads an AWS config using the default credential provider chain
// to assume the provided role with the provided session name
func (a *assumer) LoadConfigWithAssumedRole(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to get default config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	creds, err := a.assumeIamRole(ctx, stsClient)
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

func (a *assumer) assumeIamRole(ctx context.Context, client awsClientIamRoleAssumer) (*sts_types.Credentials, error) {
	// TODO(jdt): expose and/or set some of the additional fields - external ID, source identity, tags, duration, etc...
	params := &sts.AssumeRoleInput{
		RoleArn:         &a.RoleARN,
		RoleSessionName: &a.RoleSessionName,
	}
	resp, err := client.AssumeRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Credentials, nil
}
