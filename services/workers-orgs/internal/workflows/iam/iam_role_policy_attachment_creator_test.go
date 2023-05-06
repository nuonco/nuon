package iam

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testAwsClientIAMRolePolicyAttacher struct {
	mock.Mock
}

var _ awsClientIAMRolePolicyAttacher = (*testAwsClientIAMRolePolicyAttacher)(nil)

func (t *testAwsClientIAMRolePolicyAttacher) AttachRolePolicy(ctx context.Context, req *iam.AttachRolePolicyInput, opts ...func(*iam.Options)) (*iam.AttachRolePolicyOutput, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*iam.AttachRolePolicyOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

func Test_iamRoleCreatorImpl_createIAMRolePolicyAttachment(t *testing.T) {
	testIAMRolePolicyAttachErr := fmt.Errorf("test-iam-role-policy-attach-err")
	req := generics.GetFakeObj[CreateIAMRolePolicyAttachmentRequest]()

	tests := map[string]struct {
		clientFn    func(*testing.T) awsClientIAMRolePolicyAttacher
		assertFn    func(*testing.T, awsClientIAMRolePolicyAttacher)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) awsClientIAMRolePolicyAttacher {
				client := &testAwsClientIAMRolePolicyAttacher{}
				resp := &iam.AttachRolePolicyOutput{}
				client.On("AttachRolePolicy", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)
				return client
			},
			assertFn: func(t *testing.T, client awsClientIAMRolePolicyAttacher) {
				obj := client.(*testAwsClientIAMRolePolicyAttacher)
				obj.AssertNumberOfCalls(t, "AttachRolePolicy", 1)

				inp := obj.Calls[0].Arguments[1].(*iam.AttachRolePolicyInput)

				// TODO(jm): add actual assertions here
				assert.NotNil(t, inp)
			},
		},
		"error": {
			clientFn: func(t *testing.T) awsClientIAMRolePolicyAttacher {
				client := &testAwsClientIAMRolePolicyAttacher{}
				client.On("AttachRolePolicy", mock.Anything, mock.Anything, mock.Anything).Return(nil, testIAMRolePolicyAttachErr)
				return client
			},
			errExpected: testIAMRolePolicyAttachErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			roleCreator := iamRolePolicyAttachmentCreatorImpl{}
			client := test.clientFn(t)
			err := roleCreator.createIAMRolePolicyAttachment(context.Background(), client, req.RoleName, req.PolicyArn)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, client)
		})
	}
}
