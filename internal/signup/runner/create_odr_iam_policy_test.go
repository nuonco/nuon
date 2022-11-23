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

func getFakeCreateOdrIAMPolicyRequest() CreateOdrIAMPolicyRequest {
	fkr := faker.New()
	var req CreateOdrIAMPolicyRequest
	fkr.Struct().Fill(&req)
	return req
}

type testAwsClientIAMPolicy struct {
	mock.Mock
}

var _ awsClientIAMPolicy = (*testAwsClientIAMPolicy)(nil)

func (t *testAwsClientIAMPolicy) CreatePolicy(ctx context.Context, req *iam.CreatePolicyInput, opts ...func(*iam.Options)) (*iam.CreatePolicyOutput, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*iam.CreatePolicyOutput), args.Error(1)
	}

	return nil, args.Error(1)
}

func Test_odrIAMPolicyCreatorImpl_createOdrIAMPolicy(t *testing.T) {
	testIAMPolicyErr := fmt.Errorf("test-iam-policy-err")
	req := getFakeCreateOdrIAMPolicyRequest()

	tests := map[string]struct {
		clientFn    func(*testing.T) awsClientIAMPolicy
		assertFn    func(*testing.T, awsClientIAMPolicy, string)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) awsClientIAMPolicy {
				client := &testAwsClientIAMPolicy{}
				resp := &iam.CreatePolicyOutput{
					Policy: &iam_types.Policy{
						Arn: toPtr("policy-arn-test"),
					},
				}
				client.On("CreatePolicy", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil)
				return client
			},
			assertFn: func(t *testing.T, client awsClientIAMPolicy, roleArn string) {
				obj := client.(*testAwsClientIAMPolicy)
				obj.AssertNumberOfCalls(t, "CreatePolicy", 1)
				assert.Equal(t, roleArn, "policy-arn-test")

				req := obj.Calls[0].Arguments[1].(*iam.CreatePolicyInput)

				// TODO(jm): add actual assertions here
				assert.NotNil(t, req)
			},
		},
		"error": {
			clientFn: func(t *testing.T) awsClientIAMPolicy {
				client := &testAwsClientIAMPolicy{}
				client.On("CreatePolicy", mock.Anything, mock.Anything, mock.Anything).Return(nil, testIAMPolicyErr)
				return client
			},
			errExpected: testIAMPolicyErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			policyCreator := odrIAMPolicyCreatorImpl{}
			client := test.clientFn(t)
			roleArn, err := policyCreator.createOdrIAMPolicy(context.Background(), client, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, client, roleArn)
		})
	}
}

func Test_odrIAMPolicyName(t *testing.T) {
	orgID := uuid.NewString()
	roleName := odrIAMPolicyName(orgID)
	assert.Contains(t, roleName, orgID)
	assert.Contains(t, roleName, "org-odr-policy")
}
