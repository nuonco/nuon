package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func (a *assumer) fetchSTSClient(ctx context.Context) (*sts.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(a.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to get default config: %w", err)
	}
	stsClient := sts.NewFromConfig(cfg)

	// if now two step config is set, we use the default config
	if a.TwoStepConfig == nil || *(a.TwoStepConfig) == (TwoStepConfig{}) {
		return stsClient, nil
	}

	if a.TwoStepConfig.IAMRoleARN == "" {
		return nil, fmt.Errorf("iam role arn must be set to use the two step config")
	}

	// if the static creds are set, we will create an STS client using them
	if a.TwoStepConfig.SrcStaticCredentials.AccessKeyID != "" {
		credsProvider := credentials.NewStaticCredentialsProvider(
			a.TwoStepConfig.SrcStaticCredentials.AccessKeyID,
			a.TwoStepConfig.SrcStaticCredentials.SecretAccessKey,
			a.TwoStepConfig.SrcStaticCredentials.SessionToken)

		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithCredentialsProvider(credsProvider),
			config.WithRegion(a.Region))
		if err != nil {
			return nil, fmt.Errorf("failed to get config with STS creds: %w", err)
		}

		stsClient = sts.NewFromConfig(cfg)
	}

	if a.TwoStepConfig.SrcIAMRoleARN != "" {
		creds, err := a.assumeIamRole(ctx, stsClient, a.TwoStepConfig.SrcIAMRoleARN)
		if err != nil {
			return nil, fmt.Errorf("failed to assume two step src role: %w", err)
		}

		credsProvider := credentials.NewStaticCredentialsProvider(*creds.AccessKeyId,
			*creds.SecretAccessKey,
			*creds.SessionToken)
		cfg, err = config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credsProvider))
		if err != nil {
			return nil, fmt.Errorf("unable to create provider: %w", err)
		}

		stsClient = sts.NewFromConfig(cfg)
	}

	// finally, if an IAM role is set, we create a set of credentials and then return an STS client using them
	creds, err := a.assumeIamRole(ctx, stsClient, a.TwoStepConfig.IAMRoleARN)
	if err != nil {
		return nil, fmt.Errorf("failed to assume two step role: %w", err)
	}

	credsProvider := credentials.NewStaticCredentialsProvider(*creds.AccessKeyId,
		*creds.SecretAccessKey,
		*creds.SessionToken)
	cfg, err = config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credsProvider),
		config.WithRegion(a.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to create provider: %w", err)
	}

	stsClient = sts.NewFromConfig(cfg)
	return stsClient, nil
}
