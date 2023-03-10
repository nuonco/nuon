package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-generics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type testWaypointClientServerConfigGetter struct {
	mock.Mock
}

func (t *testWaypointClientServerConfigGetter) GetServerConfig(
	ctx context.Context,
	req *emptypb.Empty,
	opts ...grpc.CallOption,
) (*gen.GetServerConfigResponse, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*gen.GetServerConfigResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestGetWaypointServerCookie_validateRequest(t *testing.T) {
	tests := map[string]struct {
		reqFn       func() GetWaypointServerCookieRequest
		errExpected error
	}{
		"happy path": {
			reqFn: generics.GetFakeObj[GetWaypointServerCookieRequest],
		},
		"no-org-id": {
			reqFn: func() GetWaypointServerCookieRequest {
				req := generics.GetFakeObj[GetWaypointServerCookieRequest]()
				req.OrgID = ""
				return req
			},
			errExpected: fmt.Errorf("GetWaypointServerCookieRequest.OrgID"),
		},
		"no-namespace": {
			reqFn: func() GetWaypointServerCookieRequest {
				req := generics.GetFakeObj[GetWaypointServerCookieRequest]()
				req.TokenSecretNamespace = ""
				return req
			},
			errExpected: fmt.Errorf("GetWaypointServerCookieRequest.TokenSecretNamespace"),
		},
		"no-server-addr": {
			reqFn: func() GetWaypointServerCookieRequest {
				req := generics.GetFakeObj[GetWaypointServerCookieRequest]()
				req.OrgServerAddr = ""
				return req
			},
			errExpected: fmt.Errorf("GetWaypointServerCookieRequest.OrgServerAddr"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req := test.reqFn()
			err := req.validate()
			if test.errExpected == nil {
				assert.Nil(t, err)
			} else {
				assert.ErrorContains(t, err, test.errExpected.Error())
			}
		})
	}
}

func TestGetWaypointServerCookie_getWaypointServerCookie(t *testing.T) {
	testErr := fmt.Errorf("test-error")

	tests := map[string]struct {
		clientFn       func(*testing.T) waypointClientServerConfigGetter
		assertFn       func(*testing.T, waypointClientServerConfigGetter)
		errExpected    error
		expectedCookie string
	}{
		"happy path": {
			clientFn: func(t *testing.T) waypointClientServerConfigGetter {
				client := &testWaypointClientServerConfigGetter{}
				client.On("GetServerConfig", mock.Anything, mock.Anything, mock.Anything).Return(&gen.GetServerConfigResponse{
					Config: &gen.ServerConfig{
						Cookie: "cookie",
					},
				}, nil)
				return client
			},
			assertFn: func(t *testing.T, client waypointClientServerConfigGetter) {
				obj := client.(*testWaypointClientServerConfigGetter)
				obj.AssertNumberOfCalls(t, "UpsertProject", 1)
			},
			expectedCookie: "cookie",
		},
		"error": {
			clientFn: func(t *testing.T) waypointClientServerConfigGetter {
				client := &testWaypointClientServerConfigGetter{}
				client.On("GetServerConfig", mock.Anything, mock.Anything, mock.Anything).Return(nil, testErr)
				return client
			},
			errExpected: testErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			projectCreator := wpServerCookieGetter{}
			client := test.clientFn(t)
			cookie, err := projectCreator.getWaypointServerCookie(context.Background(), client)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Equal(t, test.expectedCookie, cookie)
		})
	}
}
