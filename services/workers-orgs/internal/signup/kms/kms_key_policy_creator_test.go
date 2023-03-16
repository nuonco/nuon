package kms

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_createKMSKeyPolicy(t *testing.T) {
	errCreateKMSKeyPolicy := fmt.Errorf("error creating kms key policy")
	req := generics.GetFakeObj[CreateKMSKeyPolicyRequest]()

	tests := map[string]struct {
		awsClient   func(*gomock.Controller) awsClientKMSKeyPolicyCreator
		errExpected error
	}{
		"happy path": {
			awsClient: func(mockCtl *gomock.Controller) awsClientKMSKeyPolicyCreator {
				mock := NewMockawsClientKMSKeyPolicyCreator(mockCtl)

				mockReq := &kms.PutKeyPolicyInput{
					KeyId:      generics.ToPtr(req.KeyID),
					Policy:     generics.ToPtr(req.Policy),
					PolicyName: generics.ToPtr(req.PolicyName),
				}
				mockResp := &kms.PutKeyPolicyOutput{}
				mock.EXPECT().PutKeyPolicy(gomock.Any(), mockReq, gomock.Any()).
					Return(mockResp, nil)

				return mock
			},
		},
		"error": {
			awsClient: func(mockCtl *gomock.Controller) awsClientKMSKeyPolicyCreator {
				mock := NewMockawsClientKMSKeyPolicyCreator(mockCtl)
				mock.EXPECT().PutKeyPolicy(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errCreateKMSKeyPolicy)
				return mock
			},
			errExpected: errCreateKMSKeyPolicy,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)
			awsClient := test.awsClient(mockCtl)

			impl := &kmsKeyPolicyCreatorImpl{}
			err := impl.createKMSKeyPolicy(ctx, awsClient, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
