package v2

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func TestDeprovisionSkipsNeverProvisionedSandboxBeforeDeletingManagedStack(t *testing.T) {
	var workflowSuite testsuite.WorkflowTestSuite
	env := workflowSuite.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

	a := &activities.Activities{}
	for name, fn := range map[string]any{
		"Get":                a.Get,
		"GetAppConfig":       a.GetAppConfig,
		"GetActionWorkflows": a.GetActionWorkflows,
		"GetAppGraph":        a.GetAppGraph,
		"GetInstallSandbox":  a.GetInstallSandbox,
		"GetInstallStack":    a.GetInstallStack,
	} {
		env.RegisterActivityWithOptions(fn, activity.RegisterOptions{Name: name})
	}

	connectionID := "con_example"
	env.OnActivity("Get", mock.Anything, mock.Anything).Return(&app.Install{
		ID:                "ins_example",
		AppConfigID:       "cfg_example",
		CloudConnectionID: &connectionID,
	}, nil)
	env.OnActivity("GetAppConfig", mock.Anything, mock.Anything).Return(&app.AppConfig{
		RunnerConfig: app.AppRunnerConfig{Type: app.AppRunnerTypeAWS},
	}, nil)
	env.OnActivity("GetActionWorkflows", mock.Anything, mock.Anything).Return([]*app.InstallActionWorkflow{}, nil)
	env.OnActivity("GetAppGraph", mock.Anything, mock.Anything).Return([]string{}, nil)
	env.OnActivity("GetInstallSandbox", mock.Anything, mock.Anything).Return(&app.InstallSandbox{
		ID:     "sbox_example",
		Status: app.InstallSandboxStatusQueued,
	}, nil)
	env.OnActivity("GetInstallStack", mock.Anything, mock.Anything).Return(&app.InstallStack{ID: "stk_example"}, nil)

	env.ExecuteWorkflow(Deprovision, &app.Workflow{
		ID: "flw_example",
		Metadata: pgtype.Hstore{
			"install_id": generics.ToPtr("ins_example"),
		},
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result app.GenerateStepsResult
	require.NoError(t, env.GetWorkflowResult(&result))

	steps := make(map[string]*app.WorkflowStep, len(result.Steps))
	for _, step := range result.Steps {
		steps[step.Name] = step
	}
	require.Contains(t, steps, "deprovision sandbox")
	require.Contains(t, steps, "delete install stack")
	require.Equal(t, app.WorkflowStepExecutionTypeSkipped, steps["deprovision sandbox"].ExecutionType)
	require.Equal(t, "sandbox is not provisioned", generics.FromPtrStr(steps["deprovision sandbox"].Metadata["reason"]))
	require.Equal(t, app.WorkflowStepExecutionTypeUser, steps["delete install stack"].ExecutionType)
	require.Greater(t, steps["delete install stack"].GroupIdx, steps["deprovision sandbox"].GroupIdx)
	require.NotContains(t, steps, "deprovision sandbox plan")
	require.NotContains(t, steps, "deprovision sandbox apply")
}
