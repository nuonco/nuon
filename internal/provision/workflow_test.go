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
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-sender"
	workers "github.com/powertoolsdev/workers-installs/internal"
	"github.com/powertoolsdev/workers-installs/internal/provision/runner"
	"github.com/powertoolsdev/workers-installs/internal/provision/sandbox"
)

func newFakeConfig() workers.Config {
	return workers.Config{
		OrgServerRootDomain:           "test.nuon.co",
		TokenSecretNamespace:          "default",
		InstallationStateBucket:       "s3://nuon-installations",
		InstallationStateBucketRegion: "us-west-2",
		SandboxBucket:                 "s3://nuon-sandboxes",
		NuonAccessRoleArn:             "arn:124355/role",
		OrgInstanceRoleTemplate:       "arn:aws:123456789:iam:role/org/%[1]s/org-instance-role-%[1]s",
		OrgInstallerRoleTemplate:      "arn:aws:123456789:iam:role/org/%[1]s/org-installer-role-%[1]s",
		OrgInstallationsRoleTemplate:  "arn:aws:123456789:iam:role/org/%[1]s/org-installations-role-%[1]s",
	}
}

func getFakeProvisionRequest() ProvisionRequest {
	return ProvisionRequest{
		InstallID: uuid.New().String(),
		OrgID:     uuid.New().String(),
		AppID:     uuid.New().String(),
		AccountSettings: &sandbox.AccountSettings{
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

	errChildWorkflow := fmt.Errorf("unable to complete workflow")

	env.OnActivity(act.StartWorkflow, mock.Anything, mock.Anything).
		Return(func(_ context.Context, _ StartWorkflowRequest) (StartWorkflowResponse, error) {
			var resp StartWorkflowResponse
			return resp, nil
		})

	sWkflow := sandbox.NewWorkflow(cfg)
	env.RegisterWorkflow(sWkflow.ProvisionSandbox)
	env.OnWorkflow(sWkflow.ProvisionSandbox, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr sandbox.ProvisionRequest) (sandbox.ProvisionResponse, error) {
			return sandbox.ProvisionResponse{}, errChildWorkflow
		})

	env.OnActivity(act.Finish, mock.Anything, mock.Anything).
		Return(func(_ context.Context, fr FinishRequest) (FinishResponse, error) {
			assert.Equal(t, fr.ErrorStep, "provision_sandbox")
			assert.Contains(t, fr.ErrorMessage, errChildWorkflow.Error())
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
	rWkflow := runner.NewWorkflow(cfg)
	env.RegisterWorkflow(rWkflow.ProvisionRunner)
	sWkflow := sandbox.NewWorkflow(cfg)
	env.RegisterWorkflow(sWkflow.ProvisionSandbox)

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

	// Mock activity implementation
	env.OnWorkflow(sWkflow.ProvisionSandbox, mock.Anything, mock.Anything).
		Return(func(_ workflow.Context, pr sandbox.ProvisionRequest) (sandbox.ProvisionResponse, error) {
			assert.Nil(t, pr.Validate())

			assert.Equal(t, orgShortID, pr.OrgID)
			assert.Equal(t, appShortID, pr.AppID)
			assert.Equal(t, installShortID, pr.InstallID)
			assert.Equal(t, req.AccountSettings, pr.AccountSettings)
			assert.Equal(t, req.SandboxSettings, pr.SandboxSettings)
			return sandbox.ProvisionResponse{TerraformOutputs: validProvisionOutput}, nil
		})

	env.OnActivity(a.StartWorkflow, mock.Anything, mock.Anything).
		Return(func(_ context.Context, ssnReq StartWorkflowRequest) (StartWorkflowResponse, error) {
			var resp StartWorkflowResponse
			assert.Nil(t, ssnReq.validate())
			return resp, nil
		})

	env.OnWorkflow(rWkflow.ProvisionRunner, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r runner.ProvisionRequest) (runner.ProvisionResponse, error) {
			var resp runner.ProvisionResponse
			assert.Nil(t, r.Validate())
			assert.Equal(t, orgShortID, r.OrgID)
			assert.Equal(t, appShortID, r.AppID)
			assert.Equal(t, installShortID, r.InstallID)
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
