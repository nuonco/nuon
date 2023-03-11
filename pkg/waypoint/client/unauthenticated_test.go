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

func TestNewUnauthenticatedProvider(t *testing.T) {
	t.Parallel()

	v := validator.New()

	tests := map[string]struct {
		v           *validator.Validate
		opts        []unauthedProviderOption
		errExpected error
	}{
		"happy path": {
			v:    v,
			opts: []unauthedProviderOption{WithUnauthenticatedConfig(Config{Address: "happy path"})},
		},
		"missing validator": {
			v:           nil,
			opts:        []unauthedProviderOption{WithUnauthenticatedConfig(Config{Address: "happy path"})},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"missing address": {
			v:           v,
			opts:        []unauthedProviderOption{},
			errExpected: fmt.Errorf("Field validation for 'Address' failed"),
		},
		"error on conifg": {
			v:           v,
			opts:        []unauthedProviderOption{func(*unauthedProvider) error { return fmt.Errorf("error on config") }},
			errExpected: fmt.Errorf("error on config"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wp, err := NewUnauthenticatedProvider(test.v, test.opts...)
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

func TestUnauthedProvider_GetClient(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		provider    func(*testing.T) *unauthedProvider
		errExpected error
	}{
		"happy path": {
			provider: func(t *testing.T) *unauthedProvider {
				cc := &grpc.ClientConn{}

				m := &mockClientGetter{}
				m.On("getClient", mock.Anything, t.Name(), "").Return(cc, pb.NewWaypointClient(cc), nil)

				return &unauthedProvider{Address: t.Name(), clientGetter: m}
			},
		},
		"on error": {
			provider: func(t *testing.T) *unauthedProvider {
				cc := &grpc.ClientConn{}

				m := &mockClientGetter{}
				m.On("getClient", mock.Anything, t.Name(), "").Return(cc, pb.NewWaypointClient(cc), fmt.Errorf("error getting client"))

				return &unauthedProvider{Address: t.Name(), clientGetter: m}
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
