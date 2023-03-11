package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecr_types "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testAwsClientIamRoleAssumer struct {
	mock.Mock
}

var _ awsClientIamRoleAssumer = (*testAwsClientIamRoleAssumer)(nil)

func (t *testAwsClientIamRoleAssumer) AssumeRole(
	ctx context.Context,
	params *sts.AssumeRoleInput,
	optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	args := t.Called(ctx, params, optFns)
	if args.Get(0) != nil {
		return args.Get(0).(*sts.AssumeRoleOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestCreateRepository_assumeIamRole(t *testing.T) {
	iamRoleArn := uuid.NewString()
	assumeIamRoleErr := fmt.Errorf("test-assume-iam-role-err")

	tests := map[string]struct {
		clientFn    func(*testing.T) awsClientIamRoleAssumer
		assertFn    func(*testing.T, awsClientIamRoleAssumer, *sts_types.Credentials)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) awsClientIamRoleAssumer {
				client := &testAwsClientIamRoleAssumer{}
				client.On("AssumeRole", mock.Anything, mock.Anything, mock.Anything).Return(&sts.AssumeRoleOutput{
					Credentials: &sts_types.Credentials{
						AccessKeyId:     generics.ToPtr("aws_access_key_id"),
						SecretAccessKey: generics.ToPtr("aws_secret_access_key"),
						SessionToken:    generics.ToPtr("aws_session_token"),
					},
				}, nil)
				return client
			},
			assertFn: func(t *testing.T, client awsClientIamRoleAssumer, creds *sts_types.Credentials) {
				obj := client.(*testAwsClientIamRoleAssumer)
				obj.AssertNumberOfCalls(t, "AssumeRole", 1)
				aReq := obj.Calls[0].Arguments[1].(*sts.AssumeRoleInput)
				assert.Equal(t, iamRoleArn, *aReq.RoleArn)
				assert.Equal(t, "aws_access_key_id", *creds.AccessKeyId)
				assert.Equal(t, "aws_secret_access_key", *creds.SecretAccessKey)
				assert.Equal(t, "aws_session_token", *creds.SessionToken)
			},
			errExpected: nil,
		},
		"error": {
			clientFn: func(t *testing.T) awsClientIamRoleAssumer {
				client := &testAwsClientIamRoleAssumer{}
				client.On("AssumeRole", mock.Anything, mock.Anything, mock.Anything).Return(nil, assumeIamRoleErr)
				return client
			},
			errExpected: assumeIamRoleErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			repoCreator := repositoryCreatorImpl{}
			client := test.clientFn(t)
			creds, err := repoCreator.assumeIamRole(context.Background(), iamRoleArn, client)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			test.assertFn(t, client, creds)
			assert.NoError(t, err)
		})
	}
}

type testAwsEcrRepoCreator struct {
	mock.Mock
}

var _ awsClientEcrRepoCreator = (*testAwsEcrRepoCreator)(nil)

func (t *testAwsEcrRepoCreator) CreateRepository(
	ctx context.Context,
	params *ecr.CreateRepositoryInput,
	optFns ...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error) {
	args := t.Called(ctx, params, optFns)
	if args.Get(0) != nil {
		return args.Get(0).(*ecr.CreateRepositoryOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestCreateRepository_createRepository(t *testing.T) {
	req := generics.GetFakeObj[CreateRepositoryRequest]()
	createRepoErr := fmt.Errorf("test-create-repo-err")
	ecrRepo := generics.GetFakeObj[*ecr_types.Repository]()

	tests := map[string]struct {
		clientFn    func(*testing.T) awsClientEcrRepoCreator
		assertFn    func(*testing.T, awsClientEcrRepoCreator, *ecr_types.Repository)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) awsClientEcrRepoCreator {
				client := &testAwsEcrRepoCreator{}
				client.On("CreateRepository", mock.Anything, mock.Anything, mock.Anything).Return(&ecr.CreateRepositoryOutput{
					Repository: ecrRepo,
				}, nil)
				return client
			},
			assertFn: func(t *testing.T, client awsClientEcrRepoCreator, repo *ecr_types.Repository) {
				obj := client.(*testAwsEcrRepoCreator)
				obj.AssertNumberOfCalls(t, "CreateRepository", 1)
				rReq := obj.Calls[0].Arguments[1].(*ecr.CreateRepositoryInput)

				assert.Equal(t, req.OrgID+"/"+req.AppID, *rReq.RepositoryName)

				assert.Equal(t, req.AppID, *rReq.Tags[0].Value)
				assert.Equal(t, "app-id", *rReq.Tags[0].Key)

				assert.Equal(t, req.OrgID, *rReq.Tags[1].Value)
				assert.Equal(t, "org-id", *rReq.Tags[1].Key)

				assert.Equal(t, repo, ecrRepo)
			},
			errExpected: nil,
		},
		"error": {
			clientFn: func(t *testing.T) awsClientEcrRepoCreator {
				client := &testAwsEcrRepoCreator{}
				client.On("CreateRepository", mock.Anything, mock.Anything, mock.Anything).Return(nil, createRepoErr)
				return client
			},
			errExpected: createRepoErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			repoCreator := repositoryCreatorImpl{}
			client := test.clientFn(t)
			repo, err := repoCreator.createECRRepo(context.Background(), req, client)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			test.assertFn(t, client, repo)
			assert.NoError(t, err)
		})
	}
}
