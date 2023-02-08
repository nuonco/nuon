package provision

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/go-common/shortid"
	appv1 "github.com/powertoolsdev/protos/workflows/generated/types/apps/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	workers "github.com/powertoolsdev/workers-apps/internal"
	"github.com/powertoolsdev/workers-apps/internal/provision/project"
	"github.com/powertoolsdev/workers-apps/internal/provision/repository"
)

func getFakeConfig() workers.Config {
	fkr := faker.New()
	var cfg workers.Config
	fkr.Struct().Fill(&cfg)
	return cfg
}

func getFakeProvisionRequest() *appv1.ProvisionRequest {
	fkr := faker.New()
	var req appv1.ProvisionRequest
	fkr.Struct().Fill(&req)

	req.OrgId = uuid.NewString()
	req.AppId = uuid.NewString()
	return &req
}

func Test_Workflow(t *testing.T) {
	cfg := getFakeConfig()
	req := getFakeProvisionRequest()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	repoWkflow := repository.NewWorkflow(cfg)
	env.RegisterWorkflow(repoWkflow.ProvisionRepository)

	projectWkflow := project.NewWorkflow(cfg)
	env.RegisterWorkflow(projectWkflow.ProvisionProject)

	wf := NewWorkflow(cfg)

	a := NewActivities()

	orgShortID, err := shortid.ParseString(req.OrgId)
	require.NoError(t, err)

	appShortID, err := shortid.ParseString(req.AppId)
	require.NoError(t, err)

	env.OnWorkflow(repoWkflow.ProvisionRepository, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r repository.ProvisionRepositoryRequest) (repository.ProvisionRepositoryResponse, error) {
			var resp repository.ProvisionRepositoryResponse
			assert.Nil(t, r.Validate())
			assert.Equal(t, orgShortID, r.OrgID)
			assert.Equal(t, appShortID, r.AppID)
			return resp, nil
		})

	env.OnWorkflow(projectWkflow.ProvisionProject, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r project.ProvisionProjectRequest) (project.ProvisionProjectResponse, error) {
			var resp project.ProvisionProjectResponse
			assert.Nil(t, r.Validate())
			assert.Equal(t, orgShortID, r.OrgID)
			assert.Equal(t, appShortID, r.AppID)
			return resp, nil
		})
	env.OnActivity(a.StartProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			return &sharedv1.StartActivityResponse{}, nil
		})
	env.OnActivity(a.FinishProvisionRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			return &sharedv1.FinishActivityResponse{}, nil
		})

	env.ExecuteWorkflow(wf.Provision, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var resp appv1.ProvisionResponse
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, &resp)
}
