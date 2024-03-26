package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type testWaypointClientRunnerAdopter struct {
	mock.Mock
}

func (t *testWaypointClientRunnerAdopter) AdoptRunner(ctx context.Context, req *gen.AdoptRunnerRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*emptypb.Empty), args.Error(1)
	}

	return nil, args.Error(1)
}

func getFakeAdoptWaypointRunnerRequest() AdoptWaypointRunnerRequest {
	return AdoptWaypointRunnerRequest{
		OrgID:                uuid.NewString(),
		InstallID:            uuid.NewString(),
		TokenSecretNamespace: "default",
		OrgServerAddr:        fmt.Sprintf("%s.nuon.co", uuid.NewString()),
		ClusterInfo:          generics.GetFakeObj[kube.ClusterInfo](),
	}
}

func TestAdoptWaypointRunner_validateRequest(t *testing.T) {
	tests := map[string]struct {
		reqFn       func() AdoptWaypointRunnerRequest
		errExpected error
	}{
		"happy path": {
			reqFn: getFakeAdoptWaypointRunnerRequest,
		},
		"no-org-id": {
			reqFn: func() AdoptWaypointRunnerRequest {
				req := getFakeAdoptWaypointRunnerRequest()
				req.OrgID = ""
				return req
			},
			errExpected: fmt.Errorf("AdoptWaypointRunnerRequest.OrgID"),
		},
		"no-namespace": {
			reqFn: func() AdoptWaypointRunnerRequest {
				req := getFakeAdoptWaypointRunnerRequest()
				req.TokenSecretNamespace = ""
				return req
			},
			errExpected: fmt.Errorf("AdoptWaypointRunnerRequest.TokenSecretNamespace"),
		},
		"no-server-addr": {
			reqFn: func() AdoptWaypointRunnerRequest {
				req := getFakeAdoptWaypointRunnerRequest()
				req.OrgServerAddr = ""
				return req
			},
			errExpected: fmt.Errorf("AdoptWaypointRunnerRequest.OrgServerAddr"),
		},
		"no-install-id": {
			reqFn: func() AdoptWaypointRunnerRequest {
				req := getFakeAdoptWaypointRunnerRequest()
				req.InstallID = ""
				return req
			},
			errExpected: fmt.Errorf("AdoptWaypointRunnerRequest.InstallID"),
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

func TestAdoptWaypointRunner_adoptWaypointRunner(t *testing.T) {
	installID := uuid.NewString()
	testErr := fmt.Errorf("test-error")

	tests := map[string]struct {
		clientFn    func(*testing.T) waypointClientRunnerAdopter
		assertFn    func(*testing.T, waypointClientRunnerAdopter)
		errExpected error
	}{
		"happy path": {
			clientFn: func(t *testing.T) waypointClientRunnerAdopter {
				client := &testWaypointClientRunnerAdopter{}
				client.On("AdoptRunner", mock.Anything, mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil)
				return client
			},
			assertFn: func(t *testing.T, client waypointClientRunnerAdopter) {
				obj := client.(*testWaypointClientRunnerAdopter)
				obj.AssertNumberOfCalls(t, "AdoptRunner", 1)

				req := obj.Calls[0].Arguments[1].(*gen.AdoptRunnerRequest)

				assert.Equal(t, req.RunnerId, installID)
				assert.Equal(t, req.Adopt, true)
			},
		},
		"error": {
			clientFn: func(t *testing.T) waypointClientRunnerAdopter {
				client := &testWaypointClientRunnerAdopter{}
				client.On("AdoptRunner", mock.Anything, mock.Anything, mock.Anything).Return(nil, testErr)
				return client
			},
			errExpected: testErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			adopter := wpRunnerAdopter{}
			client := test.clientFn(t)
			err := adopter.adoptWaypointRunner(context.Background(), client, installID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.Nil(t, err)
		})
	}
}
