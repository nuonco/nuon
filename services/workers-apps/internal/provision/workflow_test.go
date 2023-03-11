package provision

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/proto"

	"github.com/powertoolsdev/mono/pkg/generics"
	appv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	projectv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1/project/v1"
	repov1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1/repository/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	"github.com/powertoolsdev/mono/services/workers-apps/internal"
	"github.com/powertoolsdev/mono/services/workers-apps/internal/provision/project"
	"github.com/powertoolsdev/mono/services/workers-apps/internal/provision/repository"
)

func Test_Workflow(t *testing.T) {
	cfg := generics.GetFakeObj[internal.Config]()
	req := generics.GetFakeObj[*appv1.ProvisionRequest]()

	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	repoWkflow := repository.NewWorkflow(cfg)
	env.RegisterWorkflow(repoWkflow.ProvisionRepository)

	projectWkflow := project.NewWorkflow(cfg)
	env.RegisterWorkflow(projectWkflow.ProvisionProject)

	wf := NewWorkflow(cfg)

	a := NewActivities()

	prResp := generics.GetFakeObj[*repov1.ProvisionRepositoryResponse]()
	env.OnWorkflow(repoWkflow.ProvisionRepository, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *repov1.ProvisionRepositoryRequest) (*repov1.ProvisionRepositoryResponse, error) {
			assert.Nil(t, r.Validate())
			assert.Equal(t, req.OrgId, r.OrgId)
			assert.Equal(t, req.AppId, r.AppId)
			return prResp, nil
		})

	env.OnWorkflow(projectWkflow.ProvisionProject, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *projectv1.ProvisionProjectRequest) (*projectv1.ProvisionProjectResponse, error) {
			resp := &projectv1.ProvisionProjectResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, req.OrgId, r.OrgId)
			assert.Equal(t, req.AppId, r.AppId)
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
	assert.True(t, proto.Equal(resp.Repository, prResp))
}
