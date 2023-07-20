package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	serverv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/server/v1"
	workers "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/workflows/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := generics.GetFakeObj[workers.Config]()

	wkfl := runner.NewWorkflow(cfg)
	env.RegisterWorkflow(wkfl.ProvisionRunner)

	wf := NewWorkflow(cfg)
	a := NewActivities(nil)

	req := &serverv1.ProvisionServerRequest{OrgId: "0hihjnf1znsaa2j7w5hz1jx7te", Region: "us-west-2"}

	// Mock activity implementations
	env.OnActivity(a.CreateNamespace, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, cnr CreateNamespaceRequest) (CreateNamespaceResponse, error) {
			err := cnr.validate()
			assert.Nil(t, err)
			require.Equal(t, req.OrgId, cnr.NamespaceName)
			return CreateNamespaceResponse{}, nil
		})

	env.OnActivity(a.InstallWaypointServer, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, iwr InstallWaypointServerRequest) (InstallWaypointServerResponse, error) {
			err := iwr.validate()
			assert.Nil(t, err)
			require.Equal(t, fmt.Sprintf("wp-%s", req.OrgId), iwr.ReleaseName)
			return InstallWaypointServerResponse{}, nil
		})

	env.OnActivity(a.PingWaypointServer, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, pwsr PingWaypointServerRequest) (PingWaypointServerResponse, error) {
			err := validatePingWaypointServerRequest(pwsr)
			assert.Nil(t, err)
			require.Equal(t, pwsr.Addr, fmt.Sprintf("%s.%s:%d", req.OrgId, cfg.WaypointServerRootDomain, defaultWaypointServerPort))
			return PingWaypointServerResponse{}, nil
		})

	env.OnActivity(a.BootstrapWaypointServer, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, bwsr BootstrapWaypointServerRequest) (BootstrapWaypointServerResponse, error) {
			err := bwsr.validate()
			assert.Nil(t, err)
			require.Equal(
				t,
				bwsr.ServerAddr,
				fmt.Sprintf("%s.%s:%d", req.OrgId, cfg.WaypointServerRootDomain, defaultWaypointServerPort),
			)
			return BootstrapWaypointServerResponse{}, nil
		})

	env.OnActivity(a.CreateWaypointProject, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, cwp CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
			err := cwp.validate()
			assert.Nil(t, err)
			require.Equal(
				t,
				cwp.OrgServerAddr,
				fmt.Sprintf("%s.%s:%d", req.OrgId, cfg.WaypointServerRootDomain, defaultWaypointServerPort),
			)
			return CreateWaypointProjectResponse{}, nil
		})

	env.ExecuteWorkflow(wf.ProvisionServer, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	// test out the response
	var resp *serverv1.ProvisionServerResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
	assert.NoError(t, resp.Validate())
}
