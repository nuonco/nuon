package iam

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	sts_types "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/google/uuid"
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

func TestIamRoleAssumer_assumeIamRole(t *testing.T) {
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
						AccessKeyId:     toPtr("aws_access_key_id"),
						SecretAccessKey: toPtr("aws_secret_access_key"),
						SessionToken:    toPtr("aws_session_token"),
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
			assumer := iamRoleAssumerImpl{}
			client := test.clientFn(t)
			creds, err := assumer.assumeIamRole(context.Background(), iamRoleArn, client)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			test.assertFn(t, client, creds)
			assert.NoError(t, err)
		})
	}
}
