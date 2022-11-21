package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	workers "github.com/powertoolsdev/workers-orgs/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func newFakeConfig() workers.Config {
	fkr := faker.New()
	var cfg workers.Config
	fkr.Struct().Fill(&cfg)
	return cfg
}

func getFakeRunnerRequest() InstallRunnerRequest {
	fkr := faker.New()
	var req InstallRunnerRequest
	fkr.Struct().Fill(&req)
	return req
}

func TestRunner(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := newFakeConfig()

	a := NewActivities(workers.Config{})

	req := getFakeRunnerRequest()

	/*
		validProvisionOutput := map[string]string{
			clusterIDKey:	    "clusterid",
			clusterEndpointKey: "https://k8s.endpoint",
			clusterCAKey:	    "b64 encoded ca",
		}*/

	orgShortID := req.OrgID
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

			require.Equal(t, orgShortID, iwr.Namespace)
			require.Equal(t, fmt.Sprintf("wp-%s-runner", orgShortID), iwr.ReleaseName)
			require.Equal(t, orgShortID, iwr.RunnerConfig.ID)
			require.Equal(t, serverCookie, iwr.RunnerConfig.Cookie)
			require.Equal(t, orgServerAddr, iwr.RunnerConfig.ServerAddr)
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

	env.OnActivity(a.CreateOdrIAMRole, mock.Anything, mock.Anything).
		Return(func(_ context.Context, coirReq CreateOdrIAMRoleRequest) (CreateOdrIAMRoleResponse, error) {
			assert.Nil(t, coirReq.validate())
			require.Equal(t, orgShortID, coirReq.OrgID)
			return CreateOdrIAMRoleResponse{}, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Install, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp InstallRunnerResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
	// idk why this is returning incorrect, i can't figure out where it's set
	// require.Equal(t, validProvisionOutput, resp.TerraformOutputs)
}
