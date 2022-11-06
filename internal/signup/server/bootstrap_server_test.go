package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

var errUnableToBootstrap = fmt.Errorf("test-error")

type testWpClientBootstrapper struct {
	mock.Mock
}

func (t *testWpClientBootstrapper) BootstrapToken(ctx context.Context, req *emptypb.Empty, opts ...grpc.CallOption) (*gen.NewTokenResponse, error) {
	resp := t.Called(ctx, req, opts)
	tokenResp := resp.Get(0)
	if tokenResp != nil {
		return tokenResp.(*gen.NewTokenResponse), resp.Error(1)
	}
	return nil, resp.Error(1)
}

func TestBootstrapServer_bootstrapWaypointServer(t *testing.T) {
	tests := map[string]struct {
		clientFn      func(*testing.T) waypointClientBootstrapper
		errExpected   error
		tokenExpected string
	}{
		"happy path": {
			tokenExpected: "token",
			clientFn: func(t *testing.T) waypointClientBootstrapper {
				client := &testWpClientBootstrapper{}
				client.On("BootstrapToken", mock.Anything, mock.Anything, mock.Anything).Return(&gen.NewTokenResponse{Token: "token"}, nil)
				return client
			},
		},
		"error": {
			clientFn: func(t *testing.T) waypointClientBootstrapper {
				client := &testWpClientBootstrapper{}
				client.On("BootstrapToken", mock.Anything, mock.Anything, mock.Anything).Return(&gen.NewTokenResponse{}, errUnableToBootstrap)
				return client
			},
			errExpected: errUnableToBootstrap,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bootstrapper := wpServerBootstrapper{}
			client := test.clientFn(t)
			token, err := bootstrapper.bootstrapWaypointServer(context.Background(), client)
			if test.errExpected != nil {
				assert.ErrorIs(t, test.errExpected, err)
				return
			}
			assert.NotEmpty(t, token)
		})
	}
}

type testKubeClientSecretStorer struct {
	mock.Mock
}

func (m *testKubeClientSecretStorer) Apply(ctx context.Context, secret *coreapplyv1.SecretApplyConfiguration, opts metav1.ApplyOptions) (*corev1.Secret, error) {
	args := m.Called(ctx, secret, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*corev1.Secret), nil
}

func TestBootstrapServer_storeToken(t *testing.T) {
	orgID := uuid.NewString()

	tests := map[string]struct {
		clientFn    func(*testing.T) kubeClientSecretStorer
		assertFn    func(*testing.T, kubeClientSecretStorer)
		token       string
		errExpected error
	}{
		"happy path": {
			token: "token",
			clientFn: func(t *testing.T) kubeClientSecretStorer {
				client := &testKubeClientSecretStorer{}
				client.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Secret{}, nil)
				return client
			},
			assertFn: func(t *testing.T, client kubeClientSecretStorer) {
				obj := client.(*testKubeClientSecretStorer)
				obj.AssertNumberOfCalls(t, "Apply", 1)
				secret := obj.Calls[0].Arguments[1].(*coreapplyv1.SecretApplyConfiguration)

				assert.Equal(t, corev1.SecretTypeOpaque, *secret.Type)
				assert.Equal(t, "v1", *secret.TypeMetaApplyConfiguration.APIVersion)
				assert.Equal(t, "Secret", *secret.TypeMetaApplyConfiguration.Kind)
				assert.Equal(t, "waypoint-bootstrap-token-"+orgID, *secret.ObjectMetaApplyConfiguration.Name)
				assert.Equal(t, "token", secret.StringData["token"])
			},
		},
		"error": {
			clientFn: func(t *testing.T) kubeClientSecretStorer {
				client := &testKubeClientSecretStorer{}
				client.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(nil, errUnableToBootstrap)
				return client
			},
			errExpected: errUnableToBootstrap,
			assertFn: func(t *testing.T, client kubeClientSecretStorer) {
				obj := client.(*testKubeClientSecretStorer)
				obj.AssertNumberOfCalls(t, "Apply", 1)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			bootstrapper := wpServerBootstrapper{}
			client := test.clientFn(t)
			err := bootstrapper.storeBootstrapToken(context.Background(), client, orgID, test.token)

			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)

			if test.assertFn != nil {
				test.assertFn(t, client)
			}
		})
	}
}
