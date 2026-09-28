package iam

import (
	"context"
	"fmt"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"

	"github.com/nuonco/nuon/pkg/generics"
)

type stsRoleAssumer interface {
	AssumeRole(context.Context, *sts.AssumeRoleInput, ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
	AssumeRoleWithWebIdentity(context.Context, *sts.AssumeRoleWithWebIdentityInput, ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error)
}

// LoadConfigWithAssumedRole loads an AWS config using the default credential provider chain
// to assume the provided role with the provided session name
func (a *assumer) LoadConfigWithAssumedRole(ctx context.Context) (aws.Config, error) {
	stsClient, err := a.fetchSTSClient(ctx)
	if err != nil {
		return aws.Config{}, err
	}

	creds, err := a.assumeIamRole(ctx, stsClient, a.RoleARN, a.ExternalID)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to assume role: %w", err)
	}

	credsProvider := credentials.NewStaticCredentialsProvider(*creds.AccessKeyId,
		*creds.SecretAccessKey,
		*creds.SessionToken)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credsProvider),
		config.WithRegion(a.Region))
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to get config with STS creds: %w", err)
	}

	return cfg, nil
}

func (a *assumer) assumeIamRole(ctx context.Context, client stsRoleAssumer, role, externalID string) (*sts_types.Credentials, error) {
	if a.UseGithubOIDC || a.UseGCPOIDC {
		var token string
		var err error
		if a.UseGithubOIDC {
			token, err = a.getGithubOIDCToken(ctx)
		} else {
			token, err = a.getGCPOIDCToken(ctx)
		}
		if err != nil {
			return nil, fmt.Errorf("unable to get OIDC token: %w", err)
		}

		resp, err := client.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
			RoleArn:          generics.ToPtr(role),
			RoleSessionName:  generics.ToPtr(a.RoleSessionName),
			WebIdentityToken: generics.ToPtr(token),
			DurationSeconds:  generics.ToPtr(int32(a.RoleSessionDuration.Seconds())),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to assume role with web identity: %w", err)
		}

		return resp.Credentials, nil
	}

	params := &sts.AssumeRoleInput{
		RoleArn:         &role,
		RoleSessionName: &a.RoleSessionName,
		DurationSeconds: generics.ToPtr(int32(a.RoleSessionDuration.Seconds())),
	}
	if externalID != "" {
		params.ExternalId = &externalID
	}

	resp, err := client.AssumeRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Credentials, nil
}
