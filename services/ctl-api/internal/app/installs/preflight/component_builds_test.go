package preflight

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

func TestPlannedComponentBuilds(t *testing.T) {
	first := componentBuildStep(&componentdeploysyncandplan.Signal{
		ComponentID:                 "cmp123",
		ComponentConfigConnectionID: "ccc123",
	})
	generated := &app.GenerateStepsResult{Steps: []*app.WorkflowStep{
		first,
		first,
		componentBuildStep(&componentsyncimage.Signal{
			ComponentID: "cmp456",
			BuildID:     "bld456",
		}),
		{
			ExecutionType: app.WorkflowStepExecutionTypeSkipped,
			QueueSignal: &signaldb.SignalData{Signal: &componentsyncimage.Signal{
				ComponentID: "cmp-skipped",
				BuildID:     "bld-skipped",
			}},
		},
	}}

	planned := plannedComponentBuilds(generated)
	require.Equal(t, "ccc123", planned[0].ComponentConfigConnectionID)
	require.True(t, planned[0].WaitForBuild)
	require.Equal(t, "bld456", planned[1].BuildID)
	require.False(t, planned[1].WaitForBuild)
	require.Len(t, planned, 2)
}

func TestInstallPreflightRequestUsesDesiredBranchConfig(t *testing.T) {
	desiredAppConfigID := "appcfg-new"
	flw := &app.Workflow{
		ID:      "iwf123",
		OwnerID: "inl123",
		Type:    app.WorkflowTypeAppBranchConfigUpdate,
		Metadata: map[string]*string{
			"new_app_config_id": &desiredAppConfigID,
		},
	}

	req := installPreflightRequest(flw, &app.GenerateStepsResult{})
	require.Equal(t, desiredAppConfigID, req.DesiredAppConfigID)
	require.True(t, req.CheckStackOutdated)
}

func TestInstallPreflightRequestSkipsStackCheckWhenWorkflowRegeneratesIt(t *testing.T) {
	flw := &app.Workflow{Type: app.WorkflowTypeAppBranchConfigUpdate}
	step := componentBuildStep(&generateinstallstackversion.Signal{InstallStackID: "stk123"})

	require.False(t, installPreflightRequest(flw, &app.GenerateStepsResult{Steps: []*app.WorkflowStep{step}}).CheckStackOutdated)

	step.ExecutionType = app.WorkflowStepExecutionTypeSkipped
	require.True(t, installPreflightRequest(flw, &app.GenerateStepsResult{Steps: []*app.WorkflowStep{step}}).CheckStackOutdated)
}

func TestInstallPreflightRequestSkipsStackCheckForDeprovision(t *testing.T) {
	flw := &app.Workflow{Type: app.WorkflowTypeDeprovision}
	require.False(t, installPreflightRequest(flw, &app.GenerateStepsResult{}).CheckStackOutdated)
}

func componentBuildStep(sig signal.Signal) *app.WorkflowStep {
	return &app.WorkflowStep{QueueSignal: &signaldb.SignalData{Signal: sig}}
}
