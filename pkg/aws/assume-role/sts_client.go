package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func (a *assumer) fetchSTSClient(ctx context.Context) (*sts.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get default config: %w", err)
	}
	stsClient := sts.NewFromConfig(cfg)

	// if now two step config is set, we use the default config
	if a.TwoStepConfig == nil {
		return stsClient, nil
	}

	// if the static creds are set, we will create an STS client using them
	if a.TwoStepConfig.AccessKeyID != "" {
		credsProvider := credentials.NewStaticCredentialsProvider(a.TwoStepConfig.AccessKeyID,
			a.TwoStepConfig.SecretAccessKey, "")

		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithCredentialsProvider(credsProvider),
			config.WithRegion(a.Region))
		if err != nil {
			return nil, fmt.Errorf("failed to get config with STS creds: %w", err)
		}

		stsClient = sts.NewFromConfig(cfg)
	}

	// if no IAM role is set, we either return the default sts client, or the one that was provisioned using static
	// creds
	if a.TwoStepConfig.IAMRoleARN == "" {
		return stsClient, nil
	}

	// finally, if an IAM role is set, we create a set of credentials and then return an STS client using them
	creds, err := a.assumeIamRole(ctx, stsClient, a.TwoStepConfig.IAMRoleARN)
	if err != nil {
		return nil, fmt.Errorf("failed to assume two step role: %w", err)
	}

	credsProvider := credentials.NewStaticCredentialsProvider(*creds.AccessKeyId,
		*creds.SecretAccessKey,
		*creds.SessionToken)
	cfg, err = config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credsProvider))
	if err != nil {
		return nil, fmt.Errorf("unable to create provider: %w", err)
	}

	stsClient = sts.NewFromConfig(cfg)
	return stsClient, nil
}
