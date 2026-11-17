package executeworkflowstepgroup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func TestNotifyOnSkipRecordsSkipOnTarget(t *testing.T) {
	deny := []any{map[string]any{"policy_id": "pol-1", "message": "public buckets are not allowed"}}
	for name, tt := range map[string]struct {
		step *app.WorkflowStep
		want map[string]any
	}{
		"skipped after a policy deny": {
			step: &app.WorkflowStep{ID: "step-1", StepTargetID: "dpl-1", StepTargetType: "install_deploys",
				Status: app.CompositeStatus{Status: app.StatusError, Metadata: map[string]any{"deny_violations": deny}}},
			want: map[string]any{"deny_violations": deny},
		},
		"skipped sandbox reprovision": {
			step: &app.WorkflowStep{ID: "step-2", StepTargetID: "run-1", StepTargetType: "install_sandbox_runs",
				Status: app.CompositeStatus{Status: app.AwaitingApproval}},
			want: map[string]any{},
		},
		"stack version awaiting the customer": {
			step: &app.WorkflowStep{ID: "step-1", StepTargetID: "isv-1", StepTargetType: "install_stack_versions",
				Status: app.CompositeStatus{Status: app.InstallStackVersionStatusPendingUser}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.SetDataConverter(converter.NewCompositeDataConverter(
				converter.NewNilPayloadConverter(),
				converter.NewByteSlicePayloadConverter(),
				pkgdataconverter.NewJSONConverter(),
			))
			var updates []activities.UpdateFlowStepTargetStatusRequest
			env.RegisterActivityWithOptions(func(_ context.Context, req activities.UpdateFlowStepTargetStatusRequest) error {
				updates = append(updates, req)
				return nil
			}, activity.RegisterOptions{Name: "PkgWorkflowsFlowUpdateFlowStepTargetStatus"})

			env.ExecuteWorkflow(func(ctx workflow.Context) error {
				NotifyOnSkip(ctx, tt.step)
				return nil
			})
			require.NoError(t, env.GetWorkflowError())

			if tt.want == nil {
				require.Empty(t, updates)
				return
			}
			require.Len(t, updates, 1)
			require.Equal(t, tt.step.ID, updates[0].StepID)
			require.Equal(t, app.StatusUserSkipped, updates[0].Status)
			require.Equal(t, tt.want, updates[0].Metadata)
		})
	}
}
