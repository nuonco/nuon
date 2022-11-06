package server

import (
	"context"
	"fmt"
	"testing"

	"github.com/jaswdr/faker"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"github.com/powertoolsdev/workers-orgs/internal/signup/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func getFakeConfig() workers.Config {
	fkr := faker.New()
	var cfg workers.Config
	fkr.Struct().Fill(&cfg)
	return cfg
}

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := getFakeConfig()

	wkfl := runner.NewWorkflow(cfg)
	env.RegisterWorkflow(wkfl.Install)

	wf := NewWorkflow(cfg)
	a := NewActivities()

	req := ProvisionRequest{OrgID: "0hihjnf1znsaa2j7w5hz1jx7te", Region: "us-east-2"}

	// Mock activity implementations
	env.OnActivity(a.CreateNamespace, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, cnr CreateNamespaceRequest) (CreateNamespaceResponse, error) {
			err := cnr.validate()
			assert.Nil(t, err)
			require.Equal(t, req.OrgID, cnr.NamespaceName)
			return CreateNamespaceResponse{}, nil
		})

	env.OnActivity(a.InstallWaypointServer, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, iwr InstallWaypointServerRequest) (InstallWaypointServerResponse, error) {
			err := iwr.validate()
			assert.Nil(t, err)
			require.Equal(t, fmt.Sprintf("wp-%s", req.OrgID), iwr.ReleaseName)
			return InstallWaypointServerResponse{}, nil
		})

	env.OnActivity(a.ExposeWaypointServer, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, ewsr ExposeWaypointServerRequest) (ExposeWaypointServerResponse, error) {
			err := ewsr.validate()
			assert.Nil(t, err)
			require.Equal(t, req.OrgID, ewsr.NamespaceName)
			require.Equal(t, req.OrgID, ewsr.ShortID)
			return ExposeWaypointServerResponse{}, nil
		})

	env.OnActivity(a.PingWaypointServer, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, pwsr PingWaypointServerRequest) (PingWaypointServerResponse, error) {
			err := validatePingWaypointServerRequest(pwsr)
			assert.Nil(t, err)
			require.Equal(t, pwsr.Addr, fmt.Sprintf("%s.%s:%d", req.OrgID, cfg.WaypointServerRootDomain, defaultWaypointServerPort))
			return PingWaypointServerResponse{}, nil
		})

	env.OnActivity(a.BootstrapWaypointServer, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, bwsr BootstrapWaypointServerRequest) (BootstrapWaypointServerResponse, error) {
			err := bwsr.validate()
			assert.Nil(t, err)
			require.Equal(
				t,
				bwsr.ServerAddr,
				fmt.Sprintf("%s.%s:%d", req.OrgID, cfg.WaypointServerRootDomain, defaultWaypointServerPort),
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
				fmt.Sprintf("%s.%s:%d", req.OrgID, cfg.WaypointServerRootDomain, defaultWaypointServerPort),
			)
			return CreateWaypointProjectResponse{}, nil
		})

	env.ExecuteWorkflow(wf.Provision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp ProvisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
