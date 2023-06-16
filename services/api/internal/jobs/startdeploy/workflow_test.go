package startdeploy

import (
	"context"
	"fmt"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid"
	jobsv1 "github.com/powertoolsdev/mono/pkg/types/api/jobs/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	buildsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/builds/v1"
	deploysv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deploys/v1"
	executev1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/execute/v1"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	sharedactivitiesv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1/activities/v1"
	sharedactivities "github.com/powertoolsdev/mono/pkg/workflows/activities"
	meta "github.com/powertoolsdev/mono/pkg/workflows/meta"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	"github.com/powertoolsdev/mono/services/api/internal/jobs/startdeploy/activities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/anypb"
)

// NOTE(jm): unfortunately, the only way to register these workflows in the test env is to do it using the same exact
// signature. Given we'll be using these workflows from just about every domain, we should probably make a library to
// wrap these calls, so we don't have to maintain them everywhere like this.
func CreatePlan(workflow.Context, *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
	return nil, nil
}

func ExecutePlan(workflow.Context, *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
	return nil, nil
}

func TestStartDeploy(t *testing.T) {
	cfg := generics.GetFakeObj[Config]()
	wf := New(nil, cfg)
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	req := generics.GetFakeObj[*jobsv1.StartDeployRequest]()
	idResp := generics.GetFakeObj[*activities.GetIDsResponse]()
	planRef := generics.GetFakeObj[*planv1.PlanRef]()
	id := shortid.NewNanoID("cmp")
	wpMetadata := generics.GetFakeObj[*planv1.Metadata]()
	plan := &planv1.Plan{
		Actual: &planv1.Plan_WaypointPlan{
			WaypointPlan: &planv1.WaypointPlan{
				Metadata: wpMetadata,
				Component: &componentv1.Component{
					Id: id,
				},
			},
		},
	}

	deployReq := &deploysv1.DeployRequest{
		DeployId: idResp.DeployID,
		BuildId:  idResp.BuildID,
		OrgId:    wpMetadata.OrgId,
		AppId:    wpMetadata.AppId,
	}

	// register child workflows
	env.RegisterWorkflow(CreatePlan)
	env.OnWorkflow(CreatePlan, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *planv1.CreatePlanRequest) (*planv1.CreatePlanResponse, error) {
			assert.NoError(t, r.Validate())
			return &planv1.CreatePlanResponse{Plan: planRef}, nil
		})

	env.RegisterWorkflow(ExecutePlan)
	env.OnWorkflow(ExecutePlan, mock.Anything, mock.Anything).
		Return(func(ctx workflow.Context, r *executev1.ExecutePlanRequest) (*executev1.ExecutePlanResponse, error) {
			assert.NoError(t, r.Validate())
			return &executev1.ExecutePlanResponse{}, nil
		})

	// register activities
	a := activities.NewActivities(nil, nil)
	env.OnActivity(a.GetIDs, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r string) (*activities.GetIDsResponse, error) {
			return idResp, nil
		})

	env.OnActivity(a.FetchBuildPlanJob, mock.Anything, mock.Anything).
		Return(func(_ context.Context, planref *planv1.PlanRef) (*planv1.Plan, error) {
			assert.Equal(t, planRef.Bucket, planref.Bucket)
			assert.Equal(t, planRef.BucketKey, planref.BucketKey)
			assert.Equal(t, planRef.BucketAssumeRoleArn, planref.BucketAssumeRoleArn)
			return plan, nil
		})

	env.OnActivity(a.UpsertInstanceJob, mock.Anything, mock.Anything).
		Return(func(_ context.Context, deployID string) (*activities.UpsertInstanceResponse, error) {
			assert.Equal(t, req.DeployId, deployID)
			return &activities.UpsertInstanceResponse{}, nil
		})

	sharedActs := sharedactivities.Activities{}
	env.OnActivity(sharedActs.PollWorkflow, mock.Anything, mock.Anything).
		Return(func(_ context.Context, pwf *sharedactivitiesv1.PollWorkflowRequest) (*sharedactivitiesv1.PollWorkflowResponse, error) {
			wkflwResp := &buildsv1.BuildResponse{
				BuildPlan: planRef,
			}
			resp, wkflowErr := anypb.New(wkflwResp)
			assert.NoError(t, wkflowErr)
			return &sharedactivitiesv1.PollWorkflowResponse{
				Response: resp,
			}, nil
		})

	startAct := meta.NewStartActivity()
	env.OnActivity(startAct.StartRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
			resp := &sharedv1.StartActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.DeploymentsBucket, r.MetadataBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgsDeploymentsRoleTemplate, deployReq.OrgId)
			assert.Equal(t, expectedRoleARN, r.MetadataBucketAssumeRoleArn)
			// we use the InstallID from the idResp because that will represent the
			// InstallID of the build from postgres, because we don't set an
			// InstallID on a Build workflow
			expectedPrefix := prefix.InstancePath(deployReq.OrgId, deployReq.AppId, id, wpMetadata.DeploymentId, idResp.InstallID)
			assert.Equal(t, expectedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})
	finishAct := meta.NewFinishActivity()
	env.OnActivity(finishAct.FinishRequest, mock.Anything, mock.Anything).
		Return(func(_ context.Context, r *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
			resp := &sharedv1.FinishActivityResponse{}
			assert.Nil(t, r.Validate())
			assert.Equal(t, cfg.DeploymentsBucket, r.MetadataBucket)

			expectedRoleARN := fmt.Sprintf(cfg.OrgsDeploymentsRoleTemplate, deployReq.OrgId)
			assert.Equal(t, expectedRoleARN, r.MetadataBucketAssumeRoleArn)
			expectedPrefix := prefix.InstancePath(deployReq.OrgId, deployReq.AppId, id, wpMetadata.DeploymentId, idResp.InstallID)
			assert.Equal(t, expectedPrefix, r.MetadataBucketPrefix)
			return resp, nil
		})

	// exec and assert workflow
	env.ExecuteWorkflow(wf.StartDeploy, req)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	resp := &jobsv1.StartDeployResponse{}
	require.NoError(t, env.GetWorkflowResult(&resp))
	require.NotNil(t, resp)
}
