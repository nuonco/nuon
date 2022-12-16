package iam

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iam_types "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testAwsClientIAMRoleCreator struct {
	mock.Mock
}

var _ awsClientIAMRoleCreator = (*testAwsClientIAMRoleCreator)(nil)

func (t *testAwsClientIAMRoleCreator) CreateRole(ctx context.Context, req *iam.CreateRoleInput, opts ...func(*iam.Options)) (*iam.CreateRoleOutput, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*iam.CreateRoleOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

func Test_iamRoleCreatorImpl_createIAMRole(t *testing.T) {
	testIAMRoleErr := fmt.Errorf("test-iam-role-err")
	req := getFakeObj[CreateIAMRoleRequest]()

	tests := map[string]struct {
		clientFn    func(*testing.T) awsClientIAMRoleCreator
		assertFn    func(*testing.T, awsClientIAMRoleCreator, string)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) awsClientIAMRoleCreator {
				client := &testAwsClientIAMRoleCreator{}
				resp := &iam.CreateRoleOutput{
					Role: &iam_types.Role{
						Arn: toPtr("output-role-arn"),
					},
				}
				client.On("CreateRole", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)
				return client
			},
			assertFn: func(t *testing.T, client awsClientIAMRoleCreator, roleArn string) {
				obj := client.(*testAwsClientIAMRoleCreator)
				obj.AssertNumberOfCalls(t, "CreateRole", 1)
				assert.Equal(t, "output-role-arn", roleArn)

				inp := obj.Calls[0].Arguments[1].(*iam.CreateRoleInput)
				assert.NotNil(t, inp)
			},
		},
		"error": {
			clientFn: func(t *testing.T) awsClientIAMRoleCreator {
				client := &testAwsClientIAMRoleCreator{}
				client.On("CreateRole", mock.Anything, mock.Anything, mock.Anything).Return(nil, testIAMRoleErr)
				return client
			},
			errExpected: testIAMRoleErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			roleCreator := iamRoleCreatorImpl{}
			client := test.clientFn(t)
			roleArn, err := roleCreator.createIAMRole(context.Background(), client, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, client, roleArn)
		})
	}
}
