package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/generics"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1/runner/v1"
	shared "github.com/powertoolsdev/mono/services/workers-installs/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestProvisionRunner(t *testing.T) {
	return
	cfg := generics.GetFakeObj[shared.Config]()
	req := generics.GetFakeObj[*runnerv1.ProvisionRunnerRequest]()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	assert.Nil(t, cfg.Validate())

	a := NewActivities(nil, cfg)

	serverCookie := uuid.NewString()
	orgServerAddr := fmt.Sprintf("%s.%s:9701", req.OrgId, cfg.OrgServerRootDomain)

	// Mock activity implementation
	env.OnActivity(a.CreateWaypointProject, mock.Anything, mock.Anything).
		Return(func(_ context.Context, cwpReq CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
			assert.Nil(t, cwpReq.validate())

			require.Equal(t, req.OrgId, cwpReq.OrgID)
			require.Equal(t, req.InstallId, cwpReq.InstallID)
			require.Equal(t, orgServerAddr, cwpReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, cwpReq.TokenSecretNamespace)
			return CreateWaypointProjectResponse{}, nil
		})

	env.OnActivity(a.CreateWaypointWorkspace, mock.Anything, mock.Anything).
		Return(func(_ context.Context, cwwReq CreateWaypointWorkspaceRequest) (CreateWaypointWorkspaceResponse, error) {
			assert.Nil(t, cwwReq.validate())

			require.Equal(t, req.OrgId, cwwReq.OrgID)
			require.Equal(t, req.InstallId, cwwReq.InstallID)
			require.Equal(t, orgServerAddr, cwwReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, cwwReq.TokenSecretNamespace)
			return CreateWaypointWorkspaceResponse{}, nil
		})

	env.OnActivity(a.GetWaypointServerCookie, mock.Anything, mock.Anything).
		Return(func(_ context.Context, gwscReq GetWaypointServerCookieRequest) (GetWaypointServerCookieResponse, error) {
			assert.Nil(t, gwscReq.validate())

			require.Equal(t, req.OrgId, gwscReq.OrgID)
			require.Equal(t, orgServerAddr, gwscReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, gwscReq.TokenSecretNamespace)
			return GetWaypointServerCookieResponse{Cookie: serverCookie}, nil
		})

	env.OnActivity(a.InstallWaypoint, mock.Anything, mock.Anything).
		Return(func(_ context.Context, iwr InstallWaypointRequest) (InstallWaypointResponse, error) {
			assert.Nil(t, iwr.validate())

			require.Equal(t, req.InstallId, iwr.InstallID)
			require.Equal(t, req.InstallId, iwr.Namespace)
			require.Equal(t, fmt.Sprintf("wp-%s", req.InstallId), iwr.ReleaseName)
			require.Equal(t, req.InstallId, iwr.RunnerConfig.ID)
			require.Equal(t, serverCookie, iwr.RunnerConfig.Cookie)
			require.Equal(t, orgServerAddr, iwr.RunnerConfig.ServerAddr)
			return InstallWaypointResponse{}, nil
		})

	env.OnActivity(a.AdoptWaypointRunner, mock.Anything, mock.Anything).
		Return(func(_ context.Context, awrReq AdoptWaypointRunnerRequest) (AdoptWaypointRunnerResponse, error) {
			assert.Nil(t, awrReq.validate())

			require.Equal(t, req.OrgId, awrReq.OrgID)
			require.Equal(t, orgServerAddr, awrReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, awrReq.TokenSecretNamespace)
			return AdoptWaypointRunnerResponse{}, nil
		})
	env.OnActivity(a.CreateRoleBinding, mock.Anything, mock.Anything).
		Return(func(_ context.Context, crbReq CreateRoleBindingRequest) (CreateRoleBindingResponse, error) {
			assert.Nil(t, crbReq.validate())

			require.Equal(t, req.InstallId, crbReq.InstallID)
			require.Equal(t, orgServerAddr, crbReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, crbReq.TokenSecretNamespace)
			return CreateRoleBindingResponse{}, nil
		})

	env.OnActivity(a.CreateWaypointRunnerProfile, mock.Anything, mock.Anything).
		Return(func(_ context.Context, cwrpReq CreateWaypointRunnerProfileRequest) (CreateWaypointRunnerProfileResponse, error) {
			assert.Nil(t, cwrpReq.validate())
			require.Equal(t, req.InstallId, cwrpReq.InstallID)
			require.Equal(t, orgServerAddr, cwrpReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, cwrpReq.TokenSecretNamespace)
			return CreateWaypointRunnerProfileResponse{}, nil
		})

	wkflow := NewWorkflow(nil, cfg)
	env.ExecuteWorkflow(wkflow.ProvisionRunner, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp *runnerv1.ProvisionRunnerResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
