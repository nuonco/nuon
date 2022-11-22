package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iam_types "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getFakeCreateOdrIAMRoleRequest() *CreateOdrIAMRoleRequest {
	fkr := faker.New()
	var req CreateOdrIAMRoleRequest
	fkr.Struct().Fill(&req)
	return &req
}

type testAwsClientIAM struct {
	mock.Mock
}

var _ awsClientIAM = (*testAwsClientIAM)(nil)

func (t *testAwsClientIAM) CreatePolicy(ctx context.Context, req *iam.CreatePolicyInput, opts ...func(iam.Options)) (*iam.CreatePolicyOutput, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*iam.CreatePolicyOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

func (t *testAwsClientIAM) CreateRole(ctx context.Context, req *iam.CreateRoleInput, opts ...func(iam.Options)) (*iam.CreateRoleOutput, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*iam.CreateRoleOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

func (t *testAwsClientIAM) AttachRolePolicy(ctx context.Context, req *iam.AttachRolePolicyInput, opts ...func(iam.Options)) (*iam.AttachRolePolicyOutput, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*iam.AttachRolePolicyOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

func Test_odrIAMRoleName(t *testing.T) {
	orgID := uuid.NewString()
	roleName := odrIAMRoleName(orgID)
	assert.Contains(t, roleName, orgID)
	assert.Contains(t, roleName, "org-odr-role")
}

func Test_odrIAMRoleCreatorImpl_createOdrIAMPolicy(t *testing.T) {
	testIAMPolicyErr := fmt.Errorf("test-iam-policy-err")
	req := getFakeCreateOdrIAMRoleRequest()

	tests := map[string]struct {
		clientFn    func(*testing.T) awsClientIAM
		assertFn    func(*testing.T, awsClientIAM, string)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) awsClientIAM {
				client := &testAwsClientIAM{}
				resp := &iam.CreatePolicyOutput{
					Policy: &iam_types.Policy{
						Arn: toPtr("policy-arn-test"),
					},
				}
				client.On("CreatePolicy", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)
				return client
			},
			assertFn: func(t *testing.T, client awsClientIAM, roleArn string) {
				obj := client.(*testAwsClientIAM)
				obj.AssertNumberOfCalls(t, "CreatePolicy", 1)
				assert.Equal(t, roleArn, "policy-arn-test")

				req := obj.Calls[0].Arguments[1].(*iam.CreatePolicyInput)

				// TODO(jm): add actual assertions here
				assert.NotNil(t, req)
			},
		},
		"error": {
			clientFn: func(t *testing.T) awsClientIAM {
				client := &testAwsClientIAM{}
				client.On("CreatePolicy", mock.Anything, mock.Anything, mock.Anything).Return(nil, testIAMPolicyErr)
				return client
			},
			errExpected: testIAMPolicyErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			roleCreator := odrIAMRoleCreatorImpl{}
			client := test.clientFn(t)
			roleArn, err := roleCreator.createOdrIAMPolicy(context.Background(), client, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, client, roleArn)
		})
	}
}

func Test_odrIAMRoleCreatorImpl_createOdrIAMRole(t *testing.T) {
	testIAMRoleErr := fmt.Errorf("test-iam-role-err")
	req := getFakeCreateOdrIAMRoleRequest()

	tests := map[string]struct {
		clientFn    func(*testing.T) awsClientIAM
		assertFn    func(*testing.T, awsClientIAM)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) awsClientIAM {
				client := &testAwsClientIAM{}
				resp := &iam.CreateRoleOutput{}
				client.On("CreateRole", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)
				return client
			},
			assertFn: func(t *testing.T, client awsClientIAM) {
				obj := client.(*testAwsClientIAM)
				obj.AssertNumberOfCalls(t, "CreateRole", 1)

				req := obj.Calls[0].Arguments[1].(*iam.CreateRoleInput)

				// TODO(jm): add actual assertions here
				assert.NotNil(t, req)
			},
		},
		"error": {
			clientFn: func(t *testing.T) awsClientIAM {
				client := &testAwsClientIAM{}
				client.On("CreateRole", mock.Anything, mock.Anything, mock.Anything).Return(nil, testIAMRoleErr)
				return client
			},
			errExpected: testIAMRoleErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			roleCreator := odrIAMRoleCreatorImpl{}
			client := test.clientFn(t)
			err := roleCreator.createOdrIAMRole(context.Background(), client, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, client)
		})
	}
}

func Test_odrIAMRoleCreatorImpl_createOdrIAMRolePolicyAttachment(t *testing.T) {
	testIAMRolePolicyAttachErr := fmt.Errorf("test-iam-role-policy-attach-err")
	req := getFakeCreateOdrIAMRoleRequest()
	roleArn := uuid.NewString()

	tests := map[string]struct {
		clientFn    func(*testing.T) awsClientIAM
		assertFn    func(*testing.T, awsClientIAM)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) awsClientIAM {
				client := &testAwsClientIAM{}
				resp := &iam.AttachRolePolicyOutput{}
				client.On("AttachRolePolicy", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)
				return client
			},
			assertFn: func(t *testing.T, client awsClientIAM) {
				obj := client.(*testAwsClientIAM)
				obj.AssertNumberOfCalls(t, "AttachRolePolicy", 1)

				req := obj.Calls[0].Arguments[1].(*iam.AttachRolePolicyInput)

				// TODO(jm): add actual assertions here
				assert.NotNil(t, req)
			},
		},
		"error": {
			clientFn: func(t *testing.T) awsClientIAM {
				client := &testAwsClientIAM{}
				client.On("AttachRolePolicy", mock.Anything, mock.Anything, mock.Anything).Return(nil, testIAMRolePolicyAttachErr)
				return client
			},
			errExpected: testIAMRolePolicyAttachErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			roleCreator := odrIAMRoleCreatorImpl{}
			client := test.clientFn(t)
			err := roleCreator.createOdrIAMRolePolicyAttachment(context.Background(), client, roleArn, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, client)
		})
	}
}
