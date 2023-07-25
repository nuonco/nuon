package repository

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecr_types "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
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

	awsCfg, err := credentials.Fetch(ctx, &credentials.Config{
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:     req.OrgsEcrAccessRoleArn,
			SessionName: "workers-apps",
		},
	})
	if err != nil {
		return resp, fmt.Errorf("unable to fetch credentials: %w", err)
	}

	ecrClient := ecr.NewFromConfig(awsCfg)
	repo, err := a.createECRRepo(ctx, req, ecrClient)
	if err == nil {
		resp.RegistryID = *repo.RegistryId
		resp.RepositoryName = *repo.RepositoryName
		resp.RepositoryArn = *repo.RepositoryArn
		resp.RepositoryURI = *repo.RepositoryUri
		return resp, nil
	}
	if !isEntityExistsException(err) {
		return resp, fmt.Errorf("failed to create ecr repo: %w", err)
	}

	repo, err = a.getECRRepo(ctx, req, ecrClient)
	if err != nil {
		return resp, fmt.Errorf("unable to get ecr repo: %w", err)
	}

	resp.RegistryID = *repo.RegistryId
	resp.RepositoryName = *repo.RepositoryName
	resp.RepositoryArn = *repo.RepositoryArn
	resp.RepositoryURI = *repo.RepositoryUri
	return resp, nil
}

type repositoryCreator interface {
	createECRRepo(context.Context, CreateRepositoryRequest, awsClientECR) (*ecr_types.Repository, error)
	getECRRepo(context.Context, CreateRepositoryRequest, awsClientECR) (*ecr_types.Repository, error)
}

type repositoryCreatorImpl struct{}

var _ repositoryCreator = (*repositoryCreatorImpl)(nil)

type awsClientECR interface {
	CreateRepository(context.Context,
		*ecr.CreateRepositoryInput,
		...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error)

	DescribeRepositories(context.Context,
		*ecr.DescribeRepositoriesInput,
		...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error)
}

func (r *repositoryCreatorImpl) getECRRepo(ctx context.Context, req CreateRepositoryRequest, client awsClientECR) (*ecr_types.Repository, error) {
	params := &ecr.DescribeRepositoriesInput{
		RepositoryNames: []string{
			req.OrgID + "/" + req.AppID,
		},
	}
	resp, err := client.DescribeRepositories(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to describe repositories: %w", err)
	}
	if len(resp.Repositories) != 1 {
		return nil, fmt.Errorf("no repositories returned")
	}

	return &resp.Repositories[0], nil
}

func (r *repositoryCreatorImpl) createECRRepo(ctx context.Context, req CreateRepositoryRequest, client awsClientECR) (*ecr_types.Repository, error) {
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
