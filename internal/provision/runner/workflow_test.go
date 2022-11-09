package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	workers "github.com/powertoolsdev/workers-installs/internal"
)

func newFakeConfig() workers.Config {
	fkr := faker.New()
	var cfg workers.Config
	fkr.Struct().Fill(&cfg)
	return cfg
}

func getFakeProvisionRequest() ProvisionRequest {
	fkr := faker.New()
	var req ProvisionRequest
	fkr.Struct().Fill(&req)
	return req
}

func TestProvisionRunner(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := newFakeConfig()
	assert.Nil(t, cfg.Validate())

	a := NewActivities(cfg)

	req := getFakeProvisionRequest()

	serverCookie := uuid.NewString()
	orgServerAddr := fmt.Sprintf("%s.%s:9701", req.OrgID, cfg.OrgServerRootDomain)

	// Mock activity implementation
	env.OnActivity(a.CreateWaypointProject, mock.Anything, mock.Anything).
		Return(func(_ context.Context, cwpReq CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
			assert.Nil(t, cwpReq.validate())

			require.Equal(t, req.OrgID, cwpReq.OrgID)
			require.Equal(t, req.InstallID, cwpReq.InstallID)
			require.Equal(t, orgServerAddr, cwpReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, cwpReq.TokenSecretNamespace)
			return CreateWaypointProjectResponse{}, nil
		})

	env.OnActivity(a.CreateWaypointWorkspace, mock.Anything, mock.Anything).
		Return(func(_ context.Context, cwwReq CreateWaypointWorkspaceRequest) (CreateWaypointWorkspaceResponse, error) {
			assert.Nil(t, cwwReq.validate())

			require.Equal(t, req.OrgID, cwwReq.OrgID)
			require.Equal(t, req.InstallID, cwwReq.InstallID)
			require.Equal(t, orgServerAddr, cwwReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, cwwReq.TokenSecretNamespace)
			return CreateWaypointWorkspaceResponse{}, nil
		})

	env.OnActivity(a.GetWaypointServerCookie, mock.Anything, mock.Anything).
		Return(func(_ context.Context, gwscReq GetWaypointServerCookieRequest) (GetWaypointServerCookieResponse, error) {
			assert.Nil(t, gwscReq.validate())

			require.Equal(t, req.OrgID, gwscReq.OrgID)
			require.Equal(t, orgServerAddr, gwscReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, gwscReq.TokenSecretNamespace)
			return GetWaypointServerCookieResponse{Cookie: serverCookie}, nil
		})

	env.OnActivity(a.InstallWaypoint, mock.Anything, mock.Anything).
		Return(func(_ context.Context, iwr InstallWaypointRequest) (InstallWaypointResponse, error) {
			assert.Nil(t, iwr.validate())

			require.Equal(t, req.InstallID, iwr.Namespace)
			require.Equal(t, fmt.Sprintf("wp-%s", req.InstallID), iwr.ReleaseName)
			require.Equal(t, req.InstallID, iwr.RunnerConfig.ID)
			require.Equal(t, serverCookie, iwr.RunnerConfig.Cookie)
			require.Equal(t, orgServerAddr, iwr.RunnerConfig.ServerAddr)
			return InstallWaypointResponse{}, nil
		})

	env.OnActivity(a.AdoptWaypointRunner, mock.Anything, mock.Anything).
		Return(func(_ context.Context, awrReq AdoptWaypointRunnerRequest) (AdoptWaypointRunnerResponse, error) {
			assert.Nil(t, awrReq.validate())

			require.Equal(t, req.OrgID, awrReq.OrgID)
			require.Equal(t, orgServerAddr, awrReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, awrReq.TokenSecretNamespace)
			return AdoptWaypointRunnerResponse{}, nil
		})
	env.OnActivity(a.CreateRoleBinding, mock.Anything, mock.Anything).
		Return(func(_ context.Context, crbReq CreateRoleBindingRequest) (CreateRoleBindingResponse, error) {
			assert.Nil(t, crbReq.validate())

			require.Equal(t, req.InstallID, crbReq.InstallID)
			require.Equal(t, orgServerAddr, crbReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, crbReq.TokenSecretNamespace)
			return CreateRoleBindingResponse{}, nil
		})

	env.OnActivity(a.CreateWaypointRunnerProfile, mock.Anything, mock.Anything).
		Return(func(_ context.Context, cwrpReq CreateWaypointRunnerProfileRequest) (CreateWaypointRunnerProfileResponse, error) {
			assert.Nil(t, cwrpReq.validate())
			require.Equal(t, req.InstallID, cwrpReq.InstallID)
			require.Equal(t, orgServerAddr, cwrpReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, cwrpReq.TokenSecretNamespace)
			return CreateWaypointRunnerProfileResponse{}, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.ProvisionRunner, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp ProvisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
