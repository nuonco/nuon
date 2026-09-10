package signal_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	stackversion "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processjob"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstep"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstepgroup"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// TestInlineValidateSignalsScheduleNoActivities enforces the
// SignalWithInlineValidate contract. The queue skips the validate update for
// these signals and the handler folds Validate into the execute phase, which is
// only safe while Validate stays activity-free. An author adding a DB call to
// one of these Validate methods would break that silently at runtime; this
// fails the build instead.
func TestInlineValidateSignalsScheduleNoActivities(t *testing.T) {
	cases := map[string]signal.Signal{
		"execute-workflow-step-group": &executeworkflowstepgroup.Signal{
			WorkflowID: "wf-1", OwnerID: "own-1", OwnerType: "installs",
		},
		"execute-workflow-step": &executeworkflowstep.Signal{
			StepID: "stp-1", WorkflowID: "wf-1", OwnerID: "own-1", OwnerType: "installs",
		},
		"generate-install-stack-version": &stackversion.Signal{
			InstallStackID: "ist-1",
		},
		"process_job": &processjob.Signal{
			JobID: "job-1", RunnerID: "run-1",
		},
	}

	for name, sig := range cases {
		sig := sig
		t.Run(name, func(t *testing.T) {
			require.True(t, signal.IsInlineValidate(sig),
				"signal is in this table but does not declare InlineValidate")

			var ts testsuite.WorkflowTestSuite
			env := ts.NewTestWorkflowEnvironment()
			env.SetWorkerOptions(worker.Options{DeadlockDetectionTimeout: time.Minute})

			var scheduled []string
			env.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, _ converter.EncodedValues) {
				scheduled = append(scheduled, info.ActivityType.Name)
			})

			env.ExecuteWorkflow(func(ctx workflow.Context) error {
				return sig.Validate(ctx)
			})

			require.NoError(t, env.GetWorkflowError())
			require.Empty(t, scheduled,
				"Validate must not schedule activities while the signal declares InlineValidate")
		})
	}
}
