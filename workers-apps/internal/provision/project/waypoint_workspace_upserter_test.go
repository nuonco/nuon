package project

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/go-generics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

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

var _ waypointClientWorkspaceUpserter = (*testWaypointClientWorkspaceUpserter)(nil)

func Test_wpWorkspaceUpserter_upsertWaypointWorkspace(t *testing.T) {
	req := generics.GetFakeObj[UpsertWaypointWorkspaceRequest]()
	errUpsertWaypointWorkspace := fmt.Errorf("upsert waypoint workspace")

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
				obj.AssertNumberOfCalls(t, "UpsertWorkspace", 1)
				uReq := obj.Calls[0].Arguments[1].(*gen.UpsertWorkspaceRequest)

				assert.Equal(t, uReq.Workspace.Name, req.AppID)
				assert.Equal(t, uReq.Workspace.Projects[0].Project.Project, req.AppID)
			},
		},
		"error-path": {
			clientFn: func() waypointClientWorkspaceUpserter {
				client := &testWaypointClientWorkspaceUpserter{}
				client.On("UpsertWorkspace", mock.Anything, mock.Anything, mock.Anything).Return(nil, errUpsertWaypointWorkspace)
				return client
			},
			errExpected: errUpsertWaypointWorkspace,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			w := &wpWorkspaceUpserter{}
			client := tt.clientFn()

			err := w.upsertWaypointWorkspace(context.Background(), client, req.AppID)
			if tt.errExpected != nil {
				assert.ErrorContains(t, err, tt.errExpected.Error())
			} else {
				assert.Nil(t, err)
			}

			if tt.assertFn != nil {
				tt.assertFn(t, client)
			}
		})
	}
}
