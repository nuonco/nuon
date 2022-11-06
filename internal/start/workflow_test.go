package start

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"github.com/powertoolsdev/go-common/shortid"
	workers "github.com/powertoolsdev/workers-deployments/internal"
	"github.com/powertoolsdev/workers-deployments/internal/start/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func getFakeStartRequest() StartRequest {
	fkr := faker.New()
	var req StartRequest
	fkr.Struct().Fill(&req)
	return req
}

func Test_validateStartRequest(t *testing.T) {
	tests := map[string]struct {
		errExpectedMsg string
		buildReq       func() StartRequest
	}{
		"should error when org id is empty": {
			errExpectedMsg: "StartRequest.OrgID",
			buildReq: func() StartRequest {
				req := getFakeStartRequest()
				req.OrgID = ""
				return req
			},
		},
		"should error when app id is empty": {
			errExpectedMsg: "StartRequest.AppID",
			buildReq: func() StartRequest {
				req := getFakeStartRequest()
				req.AppID = ""
				return req
			},
		},
		"should error when deployment id is empty": {
			errExpectedMsg: "StartRequest.DeploymentID",
			buildReq: func() StartRequest {
				req := getFakeStartRequest()
				req.DeploymentID = ""
				return req
			},
		},
		"should error when no installs provided": {
			errExpectedMsg: "StartRequest.InstallIDs",
			buildReq: func() StartRequest {
				req := getFakeStartRequest()
				req.InstallIDs = []string(nil)
				return req
			},
		},
		"should not error when properly set": {
			buildReq: func() StartRequest {
				req := getFakeStartRequest()
				return req
			},
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			req := test.buildReq()
			err := req.validate()

			if test.errExpectedMsg != "" {
				assert.ErrorContains(t, err, test.errExpectedMsg)
			}
		})
	}
}

func getFakeConfig() workers.Config {
	fkr := faker.New()
	var cfg workers.Config
	fkr.Struct().Fill(&cfg)
	return cfg
}

func TestProvision(t *testing.T) {
	cfg := getFakeConfig()
	wkflow := NewWorkflow(cfg)

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	act := NewActivities(cfg)
	bld := build.NewWorkflow(cfg)
	env.RegisterWorkflow(bld.Build)

	req := getFakeStartRequest()

	orgShortID, err := shortid.ParseString(req.OrgID)
	assert.NoError(t, err)
	appShortID, err := shortid.ParseString(req.AppID)
	assert.NoError(t, err)

	// Mock activity implementation
	env.OnActivity(act.ProvisionInstance, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pr ProvisionInstanceRequest) (ProvisionInstanceResponse, error) {
			assert.Nil(t, pr.validate())
			assert.Equal(t, orgShortID, pr.OrgID)
			assert.Equal(t, appShortID, pr.AppID)
			return ProvisionInstanceResponse{WorkflowID: uuid.NewString()}, nil
		})

	env.OnWorkflow(bld.Build, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, br build.BuildRequest) (build.BuildResponse, error) {
			var resp build.BuildResponse
			assert.Nil(t, br.Validate())
			assert.Equal(t, orgShortID, br.OrgID)
			return resp, nil
		})

	env.ExecuteWorkflow(wkflow.Start, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp StartResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
