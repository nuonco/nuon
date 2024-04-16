package registry

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	ecrauthorization "github.com/powertoolsdev/mono/pkg/aws/ecr-authorization"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_getAccessInfo(t *testing.T) {
	errGetAccessInfo := fmt.Errorf("error get access info")
	authorization := generics.GetFakeObj[*ecrauthorization.Authorization]()
	cfg := generics.GetFakeObj[configs.OCIRegistry]()

	tests := map[string]struct {
		clientFn    func(*gomock.Controller) ecrauthorization.Client
		errExpected error
		assertFn    func(*testing.T, *ociv1.AccessInfo)
	}{
		"happy path": {
			clientFn: func(mockCtl *gomock.Controller) ecrauthorization.Client {
				mock := ecrauthorization.NewMockClient(mockCtl)
				mock.EXPECT().GetAuthorization(gomock.Any()).Return(authorization, nil)
				return mock
			},
			assertFn: func(t *testing.T, accessInfo *ociv1.AccessInfo) {
				assert.Equal(t, authorization.RegistryToken, accessInfo.Auth.Password)
				assert.Equal(t, authorization.Username, accessInfo.Auth.Username)

				assert.Equal(t, cfg.Repository, accessInfo.Image)
				assert.Equal(t, cfg.Tag, accessInfo.Tag)
			},
		},
		"error": {
			clientFn: func(mockCtl *gomock.Controller) ecrauthorization.Client {
				mock := ecrauthorization.NewMockClient(mockCtl)
				mock.EXPECT().GetAuthorization(gomock.Any()).Return(nil, errGetAccessInfo)
				return mock
			},
			errExpected: fmt.Errorf("unable to get authorization"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)

			client := test.clientFn(mockCtl)
			registry := &Registry{
				config: cfg,
			}

			accessInfo, err := registry.getECRAccessInfo(ctx, client)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			test.assertFn(t, accessInfo)
		})
	}
}
