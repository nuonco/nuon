package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecr_types "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/go-playground/validator/v10"
)

type CreateRepositoryRequest struct {
	OrgID          string `validate:"required" json:"org_id"`
	AppID          string `validate:"required" json:"app_id"`
	OrgsIamRoleArn string `validate:"required" json:"orgs_iam_role_arn"`
}

func (r CreateRepositoryRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type CreateRepositoryResponse struct{}

func (a *Activities) CreateRepository(ctx context.Context, req CreateRepositoryRequest) (CreateRepositoryResponse, error) {
	var resp CreateRepositoryResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("failed to validate request: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return resp, fmt.Errorf("failed to get default config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	creds, err := a.assumeIamRole(ctx, req.OrgsIamRoleArn, stsClient)
	if err != nil {
		return resp, fmt.Errorf("failed to assume role: %w", err)
	}

	credsProvider := credentials.NewStaticCredentialsProvider(*creds.AccessKeyId, *creds.SecretAccessKey, *creds.SessionToken)
	cfg, err = config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credsProvider))
	if err != nil {
		return resp, fmt.Errorf("failed to get config with STS creds: %w", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	if err := a.createECRRepo(ctx, req, ecrClient); err != nil {
		return resp, fmt.Errorf("failed to create ECR repo: %w", err)
	}

	return resp, nil
}

type repositoryCreator interface {
	assumeIamRole(context.Context, string, awsClientIamRoleAssumer) (*sts_types.Credentials, error)
	createECRRepo(context.Context, CreateRepositoryRequest, awsClientEcrRepoCreator) error
}

type repositoryCreatorImpl struct{}

var _ repositoryCreator = (*repositoryCreatorImpl)(nil)

type awsClientIamRoleAssumer interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

func toPtr[T any](t T) *T {
	return &t
}

type awsClientEcrRepoCreator interface {
	CreateRepository(context.Context,
		*ecr.CreateRepositoryInput,
		...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error)
}

func (r *repositoryCreatorImpl) assumeIamRole(ctx context.Context, roleArn string, client awsClientIamRoleAssumer) (*sts_types.Credentials, error) {
	params := &sts.AssumeRoleInput{
		RoleArn:         toPtr(roleArn),
		RoleSessionName: toPtr("workers-apps-create-repo"),
	}
	resp, err := client.AssumeRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Credentials, nil
}

func (r *repositoryCreatorImpl) createECRRepo(ctx context.Context, req CreateRepositoryRequest, client awsClientEcrRepoCreator) error {
	params := &ecr.CreateRepositoryInput{
		RepositoryName:     toPtr(req.AppID),
		ImageTagMutability: ecr_types.ImageTagMutabilityImmutable,
		Tags: []ecr_types.Tag{
			{
				Key:   toPtr("app-id"),
				Value: toPtr(req.AppID),
			},
			{
				Key:   toPtr("org-id"),
				Value: toPtr(req.OrgID),
			},
			{
				Key:   toPtr("managed-by"),
				Value: toPtr("workers-apps"),
			},
		},
	}

	_, err := client.CreateRepository(ctx, params)
	return err
}
