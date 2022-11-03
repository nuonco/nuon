package waypoint

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/waypoint/pkg/serverclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

var errClientTest = fmt.Errorf("test-error")

type testServerConnector struct {
	mock.Mock
}

func (m *testServerConnector) Connect(ctx context.Context, opts ...serverclient.ConnectOption) (*grpc.ClientConn, error) {
	calledArgs := []interface{}{ctx}
	for _, opt := range opts {
		calledArgs = append(calledArgs, opt)
	}
	args := m.Called(calledArgs...)
	if args.Get(0) != nil {
		return args.Get(0).(*grpc.ClientConn), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestGetUnauthenticatedClient(t *testing.T) {
	addr := fmt.Sprintf("%s.nuon.co", uuid.NewString())

	tests := map[string]struct {
		connectorFn func(*testing.T) serverConnector
		assertFn    func(*testing.T, serverConnector)
		errExpected error
	}{
		"happy path": {
			connectorFn: func(t *testing.T) serverConnector {
				connector := &testServerConnector{}
				connector.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(&grpc.ClientConn{}, nil)
				return connector
			},
			assertFn: func(t *testing.T, client serverConnector) {
				obj := client.(*testServerConnector)
				obj.AssertNumberOfCalls(t, "Connect", 1)

				connectOpts := obj.Calls[0].Arguments[1].(serverclient.ConnectOption)
				logOpts := obj.Calls[0].Arguments[2].(serverclient.ConnectOption)

				// NOTE(jm): it's currently not that trivial to test the actual options we're passing to
				// the client, because the serverclient internalizes a lot here. The option functions it
				// accepts are actually of the signature type `ConnectOption func(*connectConfig)
				// error`.
				//
				// We could refactor the option building in a future change, to split this out, but it
				// still won't give us certainty that we're configuring the server right without
				// actually using it.
				//
				// For ref https://github.com/hashicorp/waypoint/blob/main/pkg/serverclient/client.go
				assert.NotNil(t, connectOpts)
				assert.NotNil(t, logOpts)
			},
		},
		"error": {
			connectorFn: func(t *testing.T) serverConnector {
				connector := &testServerConnector{}
				connector.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(&grpc.ClientConn{}, errClientTest)
				return connector
			},
			errExpected: errClientTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			connector := test.connectorFn(t)
			clientProvider := &wpClientProvider{
				connector: connector,
			}

			_, err := clientProvider.GetUnauthenticatedWaypointClient(context.Background(), addr)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			if test.assertFn != nil {
				test.assertFn(t, connector)
			}
		})
	}
}

func TestGetOrgWaypointClient(t *testing.T) {
	orgID := uuid.NewString()
	addr := fmt.Sprintf("%s.nuon.co", orgID)
	unableToGetTokenErr := fmt.Errorf("unable-to-get-token")

	tests := map[string]struct {
		tokenGetterFn func(*testing.T) tokenGetter
		connectorFn   func(*testing.T) serverConnector
		assertFn      func(*testing.T, serverConnector, tokenGetter)
		errExpected   error
	}{
		"happy path": {
			tokenGetterFn: func(t *testing.T) tokenGetter {
				tg := &testTokenGetter{}
				tg.On("getOrgToken", mock.Anything, mock.Anything, mock.Anything).Return("token", nil)
				return tg
			},
			connectorFn: func(t *testing.T) serverConnector {
				connector := &testServerConnector{}
				connector.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(&grpc.ClientConn{}, nil)
				return connector
			},
			assertFn: func(t *testing.T, client serverConnector, tg tokenGetter) {},
		},
		"error getting token": {
			tokenGetterFn: func(t *testing.T) tokenGetter {
				tg := &testTokenGetter{}
				tg.On("getOrgToken", mock.Anything, mock.Anything, mock.Anything).Return("", unableToGetTokenErr)
				return tg
			},
			connectorFn: func(t *testing.T) serverConnector {
				connector := &testServerConnector{}
				connector.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(&grpc.ClientConn{}, errClientTest)
				return connector
			},
			assertFn: func(t *testing.T, client serverConnector, tg tokenGetter) {
				mc := client.(*testServerConnector)
				mc.AssertNumberOfCalls(t, "Connect", 0)
			},
			errExpected: unableToGetTokenErr,
		},
		"error connecting": {
			tokenGetterFn: func(t *testing.T) tokenGetter {
				tg := &testTokenGetter{}
				tg.On("getOrgToken", mock.Anything, mock.Anything, mock.Anything).Return("token", nil)
				return tg
			},
			connectorFn: func(t *testing.T) serverConnector {
				connector := &testServerConnector{}
				connector.On("Connect", mock.Anything, mock.Anything, mock.Anything).Return(&grpc.ClientConn{}, errClientTest)
				return connector
			},
			errExpected: errClientTest,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			connector := test.connectorFn(t)
			tokenGetter := test.tokenGetterFn(t)
			clientProvider := &wpClientProvider{
				connector:   connector,
				tokenGetter: tokenGetter,
			}

			_, err := clientProvider.GetOrgWaypointClient(context.Background(), "default", orgID, addr)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
			if test.assertFn != nil {
				test.assertFn(t, connector, tokenGetter)
			}
		})
	}
}

type testTokenGetter struct {
	mock.Mock
}

func (t *testTokenGetter) getOrgToken(ctx context.Context, ns, orgID string) (string, error) {
	args := t.Called(ctx, ns, orgID)
	return args.String(0), args.Error(1)
}
