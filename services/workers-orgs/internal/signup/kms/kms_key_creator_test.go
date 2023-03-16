package kms

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	kms_types "github.com/aws/aws-sdk-go-v2/service/kms/types"
	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_createKMSKey(t *testing.T) {
	errCreateKMSKey := fmt.Errorf("error creating kms key")
	req := generics.GetFakeObj[CreateKMSKeyRequest]()
	req.KeyTags = [][2]string{{"key", "value"}}

	tests := map[string]struct {
		awsClient   func(*gomock.Controller) awsClientKMSKeyCreator
		assertFn    func(*testing.T, *kms_types.KeyMetadata)
		errExpected error
	}{
		"happy path": {
			awsClient: func(mockCtl *gomock.Controller) awsClientKMSKeyCreator {
				mock := NewMockawsClientKMSKeyCreator(mockCtl)

				mockReq := &kms.CreateKeyInput{
					CustomerMasterKeySpec: kms_types.CustomerMasterKeySpecSymmetricDefault,
					KeyUsage:              kms_types.KeyUsageTypeEncryptDecrypt,
					Origin:                kms_types.OriginTypeAwsKms,
					Tags: []kms_types.Tag{
						{
							TagKey:   generics.ToPtr("key"),
							TagValue: generics.ToPtr("value"),
						},
					},
				}
				mockResp := &kms.CreateKeyOutput{
					KeyMetadata: &kms_types.KeyMetadata{
						KeyId: generics.ToPtr("key-id"),
						Arn:   generics.ToPtr("key-arn"),
					},
				}
				mock.EXPECT().CreateKey(gomock.Any(), mockReq, gomock.Any()).
					Return(mockResp, nil)

				return mock
			},
			assertFn: func(t *testing.T, resp *kms_types.KeyMetadata) {
				assert.Equal(t, "key-id", *resp.KeyId)
				assert.Equal(t, "key-arn", *resp.Arn)
			},
		},
		"error": {
			awsClient: func(mockCtl *gomock.Controller) awsClientKMSKeyCreator {
				mock := NewMockawsClientKMSKeyCreator(mockCtl)
				mock.EXPECT().CreateKey(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, errCreateKMSKey)
				return mock
			},
			errExpected: errCreateKMSKey,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)
			awsClient := test.awsClient(mockCtl)

			impl := &kmsKeyCreatorImpl{}
			resp, err := impl.createKMSKey(ctx, awsClient, req)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, resp)
		})
	}
}
