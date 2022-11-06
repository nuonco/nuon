package provision

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-common/shortid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	"github.com/powertoolsdev/go-sender"
	workers "github.com/powertoolsdev/workers-installs/internal"
)

func newFakeConfig() workers.Config {
	return workers.Config{
		OrgServerRootDomain:           "test.nuon.co",
		TokenSecretNamespace:          "default",
		InstallationStateBucket:       "s3://nuon-installations",
		InstallationStateBucketRegion: "us-west-2",
		SandboxBucket:                 "s3://nuon-sandboxes",
		NuonAccessRoleArn:             "arn:124355/role",
	}
}

func getFakeProvisionRequest() ProvisionRequest {
	return ProvisionRequest{
		InstallID: uuid.New().String(),
		OrgID:     uuid.New().String(),
		AppID:     uuid.New().String(),
		AccountSettings: &AccountSettings{
			AwsAccountID: uuid.New().String(),
			AwsRegion:    validInstallRegions()[0],
			AwsRoleArn:   uuid.New().String(),
		},
		SandboxSettings: struct {
			Name    string `json:"name" validate:"required"`
			Version string `json:"version" validate:"required"`
		}{
			Name:    "aws-eks",
			Version: "v0.0.1",
		},
	}
}

func TestProvision_finishWithErr(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := newFakeConfig()
	req := getFakeProvisionRequest()
	act := NewProvisionActivities(workers.Config{}, nil)

	errActivity := fmt.Errorf("unable to complete activity")

	env.OnActivity(act.StartWorkflow, mock.Anything, mock.Anything).
		Return(func(_ context.Context, _ StartWorkflowRequest) (StartWorkflowResponse, error) {
			var resp StartWorkflowResponse
			return resp, nil
		})

	env.OnActivity(act.ProvisionSandbox, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr ProvisionSandboxRequest) (ProvisionSandboxResponse, error) {
			return ProvisionSandboxResponse{}, errActivity
		})

	env.OnActivity(act.Finish, mock.Anything, mock.Anything).
		Return(func(_ context.Context, fr FinishRequest) (FinishResponse, error) {
			assert.Equal(t, fr.ErrorStep, "provision_sandbox")
			assert.Contains(t, fr.ErrorMessage, errActivity.Error())
			assert.False(t, fr.Success)
			return FinishResponse{}, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Provision, req)

	var resp ProvisionResponse
	require.Error(t, env.GetWorkflowResult(&resp))
}

func Test_validateProvisionRequest(t *testing.T) {
	tests := map[string]struct {
		errExpected error
		buildReq    func() ProvisionRequest
	}{
		"should error when install id is empty": {
			errExpected: ErrInvalidInstall,
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				req.InstallID = ""
				return req
			},
		},
		"should error when account id is empty": {
			errExpected: ErrInvalidAccountSettings,
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				req.AccountSettings.AwsAccountID = ""
				return req
			},
		},
		"should error when account region is empty or invalid": {
			errExpected: ErrInvalidAccountSettings,
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				req.AccountSettings.AwsRegion = "invalid"
				return req
			},
		},
		"should error when account role arn is empty or invalid": {
			errExpected: ErrInvalidAccountSettings,
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				req.AccountSettings.AwsRoleArn = ""
				return req
			},
		},
		"should error when sandbox version is empty": {
			errExpected: ErrInvalidSandboxSettings,
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				req.SandboxSettings.Version = ""
				return req
			},
		},
		"should error when sandbox name is empty": {
			errExpected: ErrInvalidSandboxSettings,
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				req.SandboxSettings.Name = ""
				return req
			},
		},
		"should not error when properly set": {
			buildReq: func() ProvisionRequest {
				req := getFakeProvisionRequest()
				return req
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			err := validateProvisionRequest(test.buildReq())

			if test.errExpected != nil {
				assert.True(t, errors.Is(err, test.errExpected))
			}
		})
	}
}

func TestProvision(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	cfg := newFakeConfig()
	assert.Nil(t, cfg.Validate())

	a := NewProvisionActivities(cfg, sender.NewNoopSender())

	req := getFakeProvisionRequest()

	validProvisionOutput := map[string]string{
		clusterIDKey:       "clusterid",
		clusterEndpointKey: "https://k8s.endpoint",
		clusterCAKey:       "b64 encoded ca",
	}
	appShortID, err := shortid.ParseString(req.AppID)
	assert.NoError(t, err)
	orgShortID, err := shortid.ParseString(req.OrgID)
	assert.NoError(t, err)
	installShortID, err := shortid.ParseString(req.InstallID)
	assert.NoError(t, err)
	serverCookie := uuid.NewString()
	orgServerAddr := fmt.Sprintf("%s.%s:9701", orgShortID, cfg.OrgServerRootDomain)

	// Mock activity implementation
	env.OnActivity(a.ProvisionSandbox, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr ProvisionSandboxRequest) (ProvisionSandboxResponse, error) {
			assert.Nil(t, pr.validate())

			assert.Equal(t, orgShortID, pr.OrgID)
			assert.Equal(t, appShortID, pr.AppID)
			assert.Equal(t, installShortID, pr.InstallID)
			assert.Equal(t, req.AccountSettings, pr.AccountSettings)
			assert.Equal(t, req.SandboxSettings, pr.SandboxSettings)
			return ProvisionSandboxResponse{Outputs: validProvisionOutput}, nil
		})

	env.OnActivity(a.CreateWaypointProject, mock.Anything, mock.Anything).
		Return(func(_ context.Context, cwpReq CreateWaypointProjectRequest) (CreateWaypointProjectResponse, error) {
			assert.Nil(t, validateCreateWaypointProjectRequest(cwpReq))

			require.Equal(t, orgShortID, cwpReq.OrgID)
			require.Equal(t, installShortID, cwpReq.InstallID)
			require.Equal(t, orgServerAddr, cwpReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, cwpReq.TokenSecretNamespace)
			return CreateWaypointProjectResponse{}, nil
		})

	env.OnActivity(a.CreateWaypointWorkspace, mock.Anything, mock.Anything).
		Return(func(_ context.Context, cwwReq CreateWaypointWorkspaceRequest) (CreateWaypointWorkspaceResponse, error) {
			assert.Nil(t, validateCreateWaypointWorkspaceRequest(cwwReq))

			require.Equal(t, orgShortID, cwwReq.OrgID)
			require.Equal(t, installShortID, cwwReq.InstallID)
			require.Equal(t, orgServerAddr, cwwReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, cwwReq.TokenSecretNamespace)
			return CreateWaypointWorkspaceResponse{}, nil
		})

	env.OnActivity(a.GetWaypointServerCookie, mock.Anything, mock.Anything).
		Return(func(_ context.Context, gwscReq GetWaypointServerCookieRequest) (GetWaypointServerCookieResponse, error) {
			assert.Nil(t, validateGetWaypointServerCookieRequest(gwscReq))

			require.Equal(t, orgShortID, gwscReq.OrgID)
			require.Equal(t, orgServerAddr, gwscReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, gwscReq.TokenSecretNamespace)
			return GetWaypointServerCookieResponse{Cookie: serverCookie}, nil
		})

	env.OnActivity(a.InstallWaypoint, mock.Anything, mock.Anything).
		Return(func(_ context.Context, iwr InstallWaypointRequest) (InstallWaypointResponse, error) {
			assert.Nil(t, validateInstallWaypointRequest(iwr))

			require.Equal(t, installShortID, iwr.Namespace)
			require.Equal(t, fmt.Sprintf("wp-%s", installShortID), iwr.ReleaseName)
			require.Equal(t, installShortID, iwr.RunnerConfig.ID)
			require.Equal(t, serverCookie, iwr.RunnerConfig.Cookie)
			require.Equal(t, orgServerAddr, iwr.RunnerConfig.ServerAddr)
			return InstallWaypointResponse{}, nil
		})

	env.OnActivity(a.AdoptWaypointRunner, mock.Anything, mock.Anything).
		Return(func(_ context.Context, awrReq AdoptWaypointRunnerRequest) (AdoptWaypointRunnerResponse, error) {
			assert.Nil(t, validateAdoptWaypointRunnerRequest(awrReq))

			require.Equal(t, orgShortID, awrReq.OrgID)
			require.Equal(t, orgServerAddr, awrReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, awrReq.TokenSecretNamespace)
			return AdoptWaypointRunnerResponse{}, nil
		})
	env.OnActivity(a.CreateRoleBinding, mock.Anything, mock.Anything).
		Return(func(_ context.Context, crbReq CreateRoleBindingRequest) (CreateRoleBindingResponse, error) {
			assert.Nil(t, crbReq.validate())

			require.Equal(t, installShortID, crbReq.InstallID)
			require.Equal(t, orgServerAddr, crbReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, crbReq.TokenSecretNamespace)
			return CreateRoleBindingResponse{}, nil
		})

	env.OnActivity(a.CreateWaypointRunnerProfile, mock.Anything, mock.Anything).
		Return(func(_ context.Context, cwrpReq CreateWaypointRunnerProfileRequest) (CreateWaypointRunnerProfileResponse, error) {
			assert.Nil(t, cwrpReq.validate())
			require.Equal(t, installShortID, cwrpReq.InstallID)
			require.Equal(t, orgServerAddr, cwrpReq.OrgServerAddr)
			require.Equal(t, cfg.TokenSecretNamespace, cwrpReq.TokenSecretNamespace)
			return CreateWaypointRunnerProfileResponse{}, nil
		})

	env.OnActivity(a.StartWorkflow, mock.Anything, mock.Anything).
		Return(func(_ context.Context, ssnReq StartWorkflowRequest) (StartWorkflowResponse, error) {
			var resp StartWorkflowResponse
			assert.Nil(t, ssnReq.validate())
			return resp, nil
		})

	env.OnActivity(a.Finish, mock.Anything, mock.Anything).
		Return(func(_ context.Context, fReq FinishRequest) (FinishResponse, error) {
			var resp FinishResponse
			assert.Nil(t, fReq.validate())
			assert.Equal(t, orgShortID, fReq.OrgID)
			assert.Equal(t, appShortID, fReq.AppID)
			assert.Equal(t, installShortID, fReq.InstallID)
			return resp, nil
		})

	wkflow := NewWorkflow(cfg)
	env.ExecuteWorkflow(wkflow.Provision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp ProvisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
	require.Equal(t, validProvisionOutput, resp.TerraformOutputs)
}
