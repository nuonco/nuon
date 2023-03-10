package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

func TestNewOrgProvider(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := map[string]struct {
		v           *validator.Validate
		opts        []orgProviderOption
		errExpected error
	}{
		"happy path": {
			v: v,
			opts: []orgProviderOption{
				WithOrgConfig(Config{Address: "happy path", Token: Token{Namespace: "happy path", Name: "happy path"}}),
			},
		},
		"missing validator": {
			v: nil,
			opts: []orgProviderOption{
				WithOrgConfig(Config{Address: "happy path", Token: Token{Namespace: "happy path", Name: "happy path"}}),
			},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"missing address": {
			v: v,
			opts: []orgProviderOption{
				WithOrgConfig(Config{Token: Token{Namespace: "happy path", Name: "happy path"}}),
			},
			errExpected: fmt.Errorf("Field validation for 'Address' failed"),
		},
		"missing namespace": {
			v: v,
			opts: []orgProviderOption{
				WithOrgConfig(Config{Address: "happy path", Token: Token{Name: "happy path"}}),
			},
			errExpected: fmt.Errorf("Field validation for 'SecretNamespace' failed"),
		},
		"missing name": {
			v: v,
			opts: []orgProviderOption{
				WithOrgConfig(Config{Address: "happy path", Token: Token{Namespace: "happy path"}}),
			},
			errExpected: fmt.Errorf("Field validation for 'SecretName' failed"),
		},
		"error during config": {
			v: v,
			opts: []orgProviderOption{
				func(op *orgProvider) error { return fmt.Errorf("error during config") },
			},
			errExpected: fmt.Errorf("error during config"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wp, err := NewOrgProvider(test.v, test.opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, name, wp.Address)
			assert.NoError(t, wp.Close())
		})
	}
}

type mockTokenGetter struct {
	mock.Mock
}

func (t *mockTokenGetter) GetOrgToken(ctx context.Context) (string, error) {
	args := t.Called(ctx)
	return args.String(0), args.Error(1)
}

func TestOrgProvider_GetClient(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		provider    func(*testing.T) *orgProvider
		errExpected error
	}{
		"happy path": {
			provider: func(t *testing.T) *orgProvider {
				cc := &grpc.ClientConn{}

				mtg := &mockTokenGetter{}
				mtg.On("GetOrgToken", mock.Anything).Return(t.Name(), nil)

				mcg := &mockClientGetter{}
				mcg.On("getClient", mock.Anything, t.Name(), t.Name()).Return(cc, pb.NewWaypointClient(cc), nil)

				return &orgProvider{Address: t.Name(), clientGetter: mcg, tokenGetter: mtg}
			},
		},
		"error getting token": {
			provider: func(t *testing.T) *orgProvider {
				mtg := &mockTokenGetter{}
				mtg.On("GetOrgToken", mock.Anything).Return("", fmt.Errorf("error getting token"))

				return &orgProvider{Address: t.Name(), clientGetter: nil, tokenGetter: mtg}
			},
			errExpected: fmt.Errorf("error getting token"),
		},
		"error getting client": {
			provider: func(t *testing.T) *orgProvider {
				cc := &grpc.ClientConn{}

				mtg := &mockTokenGetter{}
				mtg.On("GetOrgToken", mock.Anything).Return(t.Name(), nil)

				mcg := &mockClientGetter{}
				mcg.
					On("getClient", mock.Anything, t.Name(), t.Name()).
					Return(cc, pb.NewWaypointClient(cc), fmt.Errorf("error getting client"))

				return &orgProvider{Address: t.Name(), clientGetter: mcg, tokenGetter: mtg}
			},
			errExpected: fmt.Errorf("error getting client"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := test.provider(t)

			c, err := p.GetClient(context.Background())
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, c)
		})
	}
}
