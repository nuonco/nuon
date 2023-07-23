package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube"
	waypoint "github.com/powertoolsdev/mono/pkg/waypoint/client"
)

func getFakeCreateWaypointWorkspaceRequest() CreateWaypointWorkspaceRequest {
	orgID := uuid.NewString()

	return CreateWaypointWorkspaceRequest{
		OrgServerAddr:        waypoint.DefaultOrgServerAddress("stage.nuon.co", orgID),
		TokenSecretNamespace: "default",
		OrgID:                orgID,
		InstallID:            uuid.NewString(),
		ClusterInfo:          generics.GetFakeObj[kube.ClusterInfo](),
	}
}

func Test_validateCreateWaypointWorkspaceRequest(t *testing.T) {
	tests := map[string]struct {
		reqFn       func() CreateWaypointWorkspaceRequest
		errExpected error
	}{
		"happy path": {
			reqFn: func() CreateWaypointWorkspaceRequest {
				return getFakeCreateWaypointWorkspaceRequest()
			},
		},
		"missing org id": {
			reqFn: func() CreateWaypointWorkspaceRequest {
				req := getFakeCreateWaypointWorkspaceRequest()
				req.OrgID = ""
				return req
			},
			errExpected: fmt.Errorf("OrgID"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			req := tt.reqFn()
			err := req.validate()

			if tt.errExpected != nil {
				assert.ErrorContains(t, err, tt.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

type testWaypointClientWorkspaceUpserter struct {
	mock.Mock
}

func (t *testWaypointClientWorkspaceUpserter) UpsertWorkspace(ctx context.Context, req *gen.UpsertWorkspaceRequest, opts ...grpc.CallOption) (*gen.UpsertWorkspaceResponse, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*gen.UpsertWorkspaceResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func Test_wpWorkspaceCreator_createWaypointWorkspace(t *testing.T) {
	req := getFakeCreateWaypointWorkspaceRequest()

	tests := map[string]struct {
		clientFn    func() waypointClientWorkspaceUpserter
		assertFn    func(*testing.T, waypointClientWorkspaceUpserter)
		errExpected error
	}{
		"happy path": {
			clientFn: func() waypointClientWorkspaceUpserter {
				client := &testWaypointClientWorkspaceUpserter{}
				client.On("UpsertWorkspace", mock.Anything, mock.Anything, mock.Anything).Return(&gen.UpsertWorkspaceResponse{}, nil)
				return client
			},
			assertFn: func(t *testing.T, client waypointClientWorkspaceUpserter) {
				obj := client.(*testWaypointClientWorkspaceUpserter)
				obj.AssertNumberOfCalls(t, "WorkspaceProject", 1)
				uReq := obj.Calls[0].Arguments[1].(*gen.UpsertWorkspaceRequest)
				assert.Equal(t, uReq.Workspace.Name, req.OrgID)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			w := &wpWorkspaceCreator{}
			client := tt.clientFn()

			err := w.createWaypointWorkspace(context.Background(), client, req.InstallID)
			if tt.errExpected != nil {
				assert.ErrorContains(t, err, tt.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
