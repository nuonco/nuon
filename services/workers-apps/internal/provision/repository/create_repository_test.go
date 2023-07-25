package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecr_types "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testAwsEcrRepoCreator struct {
	mock.Mock
}

var _ awsClientECR = (*testAwsEcrRepoCreator)(nil)

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

func (t *testAwsEcrRepoCreator) DescribeRepositories(
	ctx context.Context,
	params *ecr.DescribeRepositoriesInput,
	optFns ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	args := t.Called(ctx, params, optFns)
	if args.Get(0) != nil {
		return args.Get(0).(*ecr.DescribeRepositoriesOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestCreateRepository_createRepository(t *testing.T) {
	req := generics.GetFakeObj[CreateRepositoryRequest]()
	createRepoErr := fmt.Errorf("test-create-repo-err")
	ecrRepo := generics.GetFakeObj[*ecr_types.Repository]()

	tests := map[string]struct {
		clientFn    func(*testing.T) awsClientECR
		assertFn    func(*testing.T, awsClientECR, *ecr_types.Repository)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) awsClientECR {
				client := &testAwsEcrRepoCreator{}
				client.On("CreateRepository", mock.Anything, mock.Anything, mock.Anything).Return(&ecr.CreateRepositoryOutput{
					Repository: ecrRepo,
				}, nil)
				return client
			},
			assertFn: func(t *testing.T, client awsClientECR, repo *ecr_types.Repository) {
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
			clientFn: func(t *testing.T) awsClientECR {
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
