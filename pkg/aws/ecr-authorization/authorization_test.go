package ecr

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecr_types "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	gomock "github.com/golang/mock/gomock"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func Test_ecrAuthorizer_getAuthorizationData(t *testing.T) {
	testErr := fmt.Errorf("testErr")
	authData := generics.GetFakeObj[ecr_types.AuthorizationData]()
	defaultRegistryID := generics.GetFakeObj[string]()

	tests := map[string]struct {
		clientFn    func(*gomock.Controller) awsECRClient
		assertFn    func(*testing.T, *ecr_types.AuthorizationData)
		registryID  string
		errExpected error
	}{
		"happy path": {
			registryID: defaultRegistryID,
			clientFn: func(mockCtl *gomock.Controller) awsECRClient {
				mock := NewMockawsECRClient(mockCtl)
				req := &ecr.GetAuthorizationTokenInput{}
				resp := &ecr.GetAuthorizationTokenOutput{
					AuthorizationData: []ecr_types.AuthorizationData{
						authData,
					},
				}
				mock.EXPECT().GetAuthorizationToken(gomock.Any(), req, gomock.Any()).Return(resp, nil)
				return mock
			},
			assertFn: func(t *testing.T, resp *ecr_types.AuthorizationData) {
				assert.Equal(t, authData, *resp)
			},
			errExpected: nil,
		},
		"client err": {
			registryID: defaultRegistryID,
			clientFn: func(mockCtl *gomock.Controller) awsECRClient {
				mock := NewMockawsECRClient(mockCtl)
				req := &ecr.GetAuthorizationTokenInput{}
				mock.EXPECT().GetAuthorizationToken(gomock.Any(), req, gomock.Any()).Return(nil, testErr)
				return mock
			},
			errExpected: testErr,
		},
		"invalid response": {
			registryID: defaultRegistryID,
			clientFn: func(mockCtl *gomock.Controller) awsECRClient {
				mock := NewMockawsECRClient(mockCtl)
				req := &ecr.GetAuthorizationTokenInput{}
				resp := &ecr.GetAuthorizationTokenOutput{
					AuthorizationData: []ecr_types.AuthorizationData{},
				}
				mock.EXPECT().GetAuthorizationToken(gomock.Any(), req, gomock.Any()).Return(resp, nil)
				return mock
			},
			errExpected: fmt.Errorf("invalid get authorization token response"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			client := test.clientFn(mockCtl)

			authorizer := &ecrAuthorizer{
				RegistryID: test.registryID,
			}
			resp, err := authorizer.getAuthorizationData(ctx, client)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, resp)
		})
	}
}
