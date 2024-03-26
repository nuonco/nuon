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
)

type testWaypointClientProjectUpserter struct {
	mock.Mock
}

func (t *testWaypointClientProjectUpserter) UpsertProject(ctx context.Context, req *gen.UpsertProjectRequest, opts ...grpc.CallOption) (*gen.UpsertProjectResponse, error) {
	args := t.Called(ctx, req, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*gen.UpsertProjectResponse), args.Error(1)
	}

	return nil, args.Error(1)
}

func TestCreateWaypointProject_validateRequest(t *testing.T) {
	clusterInfo := generics.GetFakeObj[kube.ClusterInfo]()
	clusterInfo.KubeConfig = ""

	tests := map[string]struct {
		reqFn       func() CreateWaypointProjectRequest
		errExpected error
	}{
		"happy path": {
			reqFn: func() CreateWaypointProjectRequest {
				return CreateWaypointProjectRequest{
					OrgID:                uuid.NewString(),
					TokenSecretNamespace: "default",
					OrgServerAddr:        fmt.Sprintf("%s.nuon.co", uuid.NewString()),
					InstallID:            uuid.NewString(),
					ClusterInfo:          clusterInfo,
				}
			},
		},
		"no-org-id": {
			reqFn: func() CreateWaypointProjectRequest {
				return CreateWaypointProjectRequest{
					OrgID:                "",
					TokenSecretNamespace: "default",
					OrgServerAddr:        fmt.Sprintf("%s.nuon.co", uuid.NewString()),
					InstallID:            uuid.NewString(),
					ClusterInfo:          clusterInfo,
				}
			},
			errExpected: fmt.Errorf("CreateWaypointProjectRequest.OrgID"),
		},
		"no-namespace-id": {
			reqFn: func() CreateWaypointProjectRequest {
				return CreateWaypointProjectRequest{
					OrgID:                uuid.NewString(),
					TokenSecretNamespace: "",
					OrgServerAddr:        fmt.Sprintf("%s.nuon.co", uuid.NewString()),
					InstallID:            uuid.NewString(),
					ClusterInfo:          clusterInfo,
				}
			},
			errExpected: fmt.Errorf("CreateWaypointProjectRequest.TokenSecretNamespace"),
		},
		"no-server-addr": {
			reqFn: func() CreateWaypointProjectRequest {
				return CreateWaypointProjectRequest{
					OrgID:                uuid.NewString(),
					TokenSecretNamespace: "default",
					OrgServerAddr:        "",
					InstallID:            uuid.NewString(),
					ClusterInfo:          clusterInfo,
				}
			},
			errExpected: fmt.Errorf("CreateWaypointProjectRequest.OrgServerAddr"),
		},
		"no-install-id": {
			reqFn: func() CreateWaypointProjectRequest {
				return CreateWaypointProjectRequest{
					OrgID:                uuid.NewString(),
					TokenSecretNamespace: "default",
					OrgServerAddr:        fmt.Sprintf("%s.nuon.co", uuid.NewString()),
					InstallID:            "",
					ClusterInfo:          clusterInfo,
				}
			},
			errExpected: fmt.Errorf("CreateWaypointProjectRequest.InstallID"),
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

func TestCreateWaypointProject_createWaypointProject(t *testing.T) {
	installID := uuid.NewString()
	testErr := fmt.Errorf("test-error")

	tests := map[string]struct {
		clientFn    func() waypointClientProjectUpserter
		assertFn    func(*testing.T, waypointClientProjectUpserter)
		errExpected error
	}{
		"happy path": {
			clientFn: func() waypointClientProjectUpserter {
				client := &testWaypointClientProjectUpserter{}
				client.On("UpsertProject", mock.Anything, mock.Anything, mock.Anything).Return(&gen.UpsertProjectResponse{}, nil)
				return client
			},
			assertFn: func(t *testing.T, client waypointClientProjectUpserter) {
				obj := client.(*testWaypointClientProjectUpserter)
				obj.AssertNumberOfCalls(t, "UpsertProject", 1)
				req := obj.Calls[0].Arguments[1].(*gen.UpsertProjectRequest)
				assert.Equal(t, req.Project.Name, installID)

				assert.True(t, req.Project.RemoteEnabled)
				assert.NotNil(t, req.Project.DataSource.Source)
				assert.False(t, req.Project.DataSourcePoll.Enabled)

				byts, err := getProjectWaypointConfig(installID)
				assert.NoError(t, err)
				assert.Equal(t, byts, req.Project.WaypointHcl)
				assert.Equal(t, gen.Hcl_JSON, req.Project.WaypointHclFormat)
			},
		},
		"error": {
			clientFn: func() waypointClientProjectUpserter {
				client := &testWaypointClientProjectUpserter{}
				client.On("UpsertProject", mock.Anything, mock.Anything, mock.Anything).Return(nil, testErr)
				return client
			},
			errExpected: testErr,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			projectCreator := wpProjectCreator{}
			client := test.clientFn()
			err := projectCreator.createWaypointProject(context.Background(), client, installID)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			test.assertFn(t, client)
		})
	}
}
