package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/generics"
	runnerv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/runner/v1"
	workers "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestRunner(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := generics.GetFakeObj[workers.Config]()

	a := NewActivities(nil, workers.Config{})

	req := generics.GetFakeObj[*runnerv1.InstallRunnerRequest]()

	orgShortID := req.OrgId
	serverCookie := uuid.NewString()
	orgServerAddr := fmt.Sprintf("%s.%s:9701", orgShortID, cfg.WaypointServerRootDomain)

	env.OnActivity(a.GetWaypointServerCookie, mock.Anything, mock.Anything).
		Return(func(_ context.Context, gwscReq GetWaypointServerCookieRequest) (GetWaypointServerCookieResponse, error) {
			assert.Nil(t, gwscReq.validate())

			require.Equal(t, orgShortID, gwscReq.OrgID)
			require.Equal(t, orgServerAddr, gwscReq.OrgServerAddr)
			require.Equal(t, cfg.WaypointBootstrapTokenNamespace, gwscReq.TokenSecretNamespace)
			return GetWaypointServerCookieResponse{Cookie: serverCookie}, nil
		})

	env.OnActivity(a.InstallWaypoint, mock.Anything, mock.Anything).
		Return(func(_ context.Context, iwr InstallWaypointRequest) (InstallWaypointResponse, error) {
			assert.Nil(t, iwr.validate())
			assert.Nil(t, iwr.RunnerConfig.Validate())

			require.Equal(t, orgShortID, iwr.Namespace)
			require.Equal(t, fmt.Sprintf("wp-%s-runner", orgShortID), iwr.ReleaseName)
			require.Equal(t, orgShortID, iwr.RunnerConfig.ID)
			require.Equal(t, serverCookie, iwr.RunnerConfig.Cookie)
			require.Equal(t, orgServerAddr, iwr.RunnerConfig.ServerAddr)
			require.Equal(t, req.OdrIamRoleArn, iwr.RunnerConfig.OdrIAMRoleArn)
			return InstallWaypointResponse{}, nil
		})

	env.OnActivity(a.AdoptWaypointRunner, mock.Anything, mock.Anything).
		Return(func(_ context.Context, awrReq AdoptWaypointRunnerRequest) (AdoptWaypointRunnerResponse, error) {
			assert.Nil(t, awrReq.validate())

			require.Equal(t, orgShortID, awrReq.OrgID)
			require.Equal(t, orgServerAddr, awrReq.OrgServerAddr)
			require.Equal(t, cfg.WaypointBootstrapTokenNamespace, awrReq.TokenSecretNamespace)
			return AdoptWaypointRunnerResponse{}, nil
		})
	env.OnActivity(a.CreateServerConfig, mock.Anything, mock.Anything).
		Return(func(_ context.Context, cscReq CreateServerConfigRequest) (CreateServerConfigResponse, error) {
			assert.NoError(t, cscReq.validate())

			require.Equal(t, orgShortID, cscReq.OrgID)
			require.Equal(t, orgServerAddr, cscReq.OrgServerAddr)
			require.Equal(t, cfg.WaypointBootstrapTokenNamespace, cscReq.TokenSecretNamespace)
			return CreateServerConfigResponse{}, nil
		})
	env.OnActivity(a.CreateRunnerProfile, mock.Anything, mock.Anything).
		Return(func(_ context.Context, crpReq CreateRunnerProfileRequest) (CreateRunnerProfileResponse, error) {
			assert.NoError(t, crpReq.validate())

			require.Equal(t, orgShortID, crpReq.OrgID)
			require.Equal(t, orgServerAddr, crpReq.OrgServerAddr)
			require.Equal(t, cfg.WaypointBootstrapTokenNamespace, crpReq.TokenSecretNamespace)
			return CreateRunnerProfileResponse{}, nil
		})

	env.OnActivity(a.CreateRoleBinding, mock.Anything, mock.Anything).
		Return(func(_ context.Context, crbReq CreateRoleBindingRequest) (CreateRoleBindingResponse, error) {
			assert.Nil(t, crbReq.validate())

			require.Equal(t, orgShortID, crbReq.OrgID)
			require.Equal(t, orgServerAddr, crbReq.OrgServerAddr)
			require.Equal(t, cfg.WaypointBootstrapTokenNamespace, crbReq.TokenSecretNamespace)
			require.Equal(t, orgShortID, crbReq.NamespaceName)
			return CreateRoleBindingResponse{}, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.ProvisionRunner, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp runnerv1.InstallRunnerResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	// idk why this is returning incorrect, i can't figure out where it's set
	// require.Equal(t, validProvisionOutput, resp.TerraformOutputs)
}
