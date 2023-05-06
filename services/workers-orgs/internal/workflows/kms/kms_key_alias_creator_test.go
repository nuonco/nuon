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

func Test_createKMSKeyAlias(t *testing.T) {
	errCreateKMSKeyAlias := fmt.Errorf("error creating kms key policy")
	req := generics.GetFakeObj[CreateKMSKeyAliasRequest]()

	tests := map[string]struct {
		awsClient   func(*gomock.Controller) awsClientKMSKeyAliasCreator
		errExpected error
	}{
		"happy path": {
			awsClient: func(mockCtl *gomock.Controller) awsClientKMSKeyAliasCreator {
				mock := NewMockawsClientKMSKeyAliasCreator(mockCtl)

				mockReq := &kms.CreateAliasInput{
					TargetKeyId: generics.ToPtr(req.KeyID),
					AliasName:   generics.ToPtr(req.Alias),
				}
				mockResp := &kms.CreateAliasOutput{}
				mock.EXPECT().CreateAlias(gomock.Any(), mockReq, gomock.Any()).
					Return(mockResp, nil)

				return mock
			},
		},
		"error": {
			awsClient: func(mockCtl *gomock.Controller) awsClientKMSKeyAliasCreator {
				mock := NewMockawsClientKMSKeyAliasCreator(mockCtl)
				mock.EXPECT().CreateAlias(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errCreateKMSKeyAlias)
				return mock
			},
			errExpected: errCreateKMSKeyAlias,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)
			awsClient := test.awsClient(mockCtl)

			impl := &kmsKeyAliasCreatorImpl{}
			err := impl.createKMSKeyAlias(ctx, awsClient, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}
