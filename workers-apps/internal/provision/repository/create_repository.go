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
	"github.com/powertoolsdev/go-generics"
)

type CreateRepositoryRequest struct {
	OrgID                string `validate:"required" json:"org_id"`
	AppID                string `validate:"required" json:"app_id"`
	OrgsEcrAccessRoleArn string `validate:"required" json:"orgs_ecr_access_role_arn"`
}

func (r CreateRepositoryRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type CreateRepositoryResponse struct {
	RegistryID     string
	RepositoryName string
	RepositoryArn  string
	RepositoryURI  string
}

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
	creds, err := a.assumeIamRole(ctx, req.OrgsEcrAccessRoleArn, stsClient)
	if err != nil {
		return resp, fmt.Errorf("failed to assume role: %w", err)
	}

	credsProvider := credentials.NewStaticCredentialsProvider(*creds.AccessKeyId, *creds.SecretAccessKey, *creds.SessionToken)
	cfg, err = config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(credsProvider))
	if err != nil {
		return resp, fmt.Errorf("failed to get config with STS creds: %w", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	repo, err := a.createECRRepo(ctx, req, ecrClient)
	if err != nil {
		return resp, fmt.Errorf("failed to create ecr repo: %w", err)
	}

	resp.RegistryID = *repo.RegistryId
	resp.RepositoryName = *repo.RepositoryName
	resp.RepositoryArn = *repo.RepositoryArn
	resp.RepositoryURI = *repo.RepositoryUri
	return resp, nil
}

type repositoryCreator interface {
	assumeIamRole(context.Context, string, awsClientIamRoleAssumer) (*sts_types.Credentials, error)
	createECRRepo(context.Context, CreateRepositoryRequest, awsClientEcrRepoCreator) (*ecr_types.Repository, error)
}

type repositoryCreatorImpl struct{}

var _ repositoryCreator = (*repositoryCreatorImpl)(nil)

type awsClientIamRoleAssumer interface {
	AssumeRole(ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

type awsClientEcrRepoCreator interface {
	CreateRepository(context.Context,
		*ecr.CreateRepositoryInput,
		...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error)
}

func (r *repositoryCreatorImpl) assumeIamRole(ctx context.Context, roleArn string, client awsClientIamRoleAssumer) (*sts_types.Credentials, error) {
	params := &sts.AssumeRoleInput{
		RoleArn:         &roleArn,
		RoleSessionName: generics.ToPtr("workers-apps-create-repo"),
	}
	resp, err := client.AssumeRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Credentials, nil
}

func (r *repositoryCreatorImpl) createECRRepo(ctx context.Context, req CreateRepositoryRequest, client awsClientEcrRepoCreator) (*ecr_types.Repository, error) {
	params := &ecr.CreateRepositoryInput{
		RepositoryName:     generics.ToPtr(req.OrgID + "/" + req.AppID),
		ImageTagMutability: ecr_types.ImageTagMutabilityImmutable,
		Tags: []ecr_types.Tag{
			{
				Key:   generics.ToPtr("app-id"),
				Value: &req.AppID,
			},
			{
				Key:   generics.ToPtr("org-id"),
				Value: &req.OrgID,
			},
			{
				Key:   generics.ToPtr("managed-by"),
				Value: generics.ToPtr("workers-apps"),
			},
		},
	}

	resp, err := client.CreateRepository(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to create repository: %w", err)
	}

	return resp.Repository, nil
}
