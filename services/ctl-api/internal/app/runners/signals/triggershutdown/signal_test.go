package triggershutdown

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const (
	runnerID  = "runacme"
	freshPID  = "rprfresh"
	stalePID  = "rprstale"
	queueID   = "questale"
	queueName = "runner-process-" + stalePID
)

func activeProcess(id string) *app.RunnerProcess {
	return &app.RunnerProcess{
		ID:              id,
		RunnerID:        runnerID,
		Type:            app.RunnerProcessTypeInstall,
		CompositeStatus: app.CompositeStatus{Status: app.Status(app.RunnerProcessStatusActive)},
	}
}

func stoppedProcess(id string) *app.RunnerProcess {
	p := activeProcess(id)
	p.CompositeStatus.Status = app.Status(app.RunnerProcessStatusShutDown)
	return p
}

func runSignal(t *testing.T, sig *Signal, setup func(env *testsuite.TestWorkflowEnvironment)) *testsuite.TestWorkflowEnvironment {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.SetDataConverter(converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pkgdataconverter.NewJSONConverter(),
	))
	env.RegisterActivity(&activities.Activities{})
	env.RegisterActivity(&queueclient.Client{})
	setup(env)

	env.ExecuteWorkflow(func(ctx workflow.Context) error {
		ctx = cctx.SetQueueIDWorkflowContext(ctx, queueID)
		return sig.Execute(ctx)
	})
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	return env
}

func TestExecute_TargetsPinnedProcess(t *testing.T) {
	var shutdownFor string
	env := runSignal(t, &Signal{RunnerID: runnerID, ProcessType: "install", ProcessID: stalePID}, func(env *testsuite.TestWorkflowEnvironment) {
		env.OnActivity("GetRunnerProcess", mock.Anything, activities.GetRunnerProcessRequest{ProcessID: stalePID}).
			Return(activeProcess(stalePID), nil).Once()
		env.OnActivity("CreateRunnerProcessShutdown", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				shutdownFor = args.Get(1).(activities.CreateRunnerProcessShutdownRequest).RunnerProcessID
			}).
			Return(&app.RunnerProcessShutdown{}, nil).Once()
	})
	env.AssertExpectations(t)
	require.Equal(t, stalePID, shutdownFor)
	env.AssertNotCalled(t, "GetCurrentRunnerProcess", mock.Anything, mock.Anything)
}

func TestExecute_PinnedProcessAlreadyStopped_NoShutdown(t *testing.T) {
	env := runSignal(t, &Signal{RunnerID: runnerID, ProcessType: "install", ProcessID: stalePID}, func(env *testsuite.TestWorkflowEnvironment) {
		env.OnActivity("GetRunnerProcess", mock.Anything, activities.GetRunnerProcessRequest{ProcessID: stalePID}).
			Return(stoppedProcess(stalePID), nil).Once()
	})
	env.AssertExpectations(t)
	env.AssertNotCalled(t, "CreateRunnerProcessShutdown", mock.Anything, mock.Anything)
	env.AssertNotCalled(t, "GetCurrentRunnerProcess", mock.Anything, mock.Anything)
}

// Legacy template: no process_id, but the emitter sits on the dead process's queue.
// The fresh current process must survive.
func TestExecute_LegacyTemplateResolvesFromQueue_DoesNotKillCurrent(t *testing.T) {
	env := runSignal(t, &Signal{RunnerID: runnerID, ProcessType: "install"}, func(env *testsuite.TestWorkflowEnvironment) {
		env.OnActivity("GetQueue", mock.Anything, queueID).
			Return(&app.Queue{ID: queueID, Name: queueName, OwnerID: runnerID}, nil).Once()
		env.OnActivity("GetRunnerProcess", mock.Anything, activities.GetRunnerProcessRequest{ProcessID: stalePID}).
			Return(stoppedProcess(stalePID), nil).Once()
	})
	env.AssertExpectations(t)
	env.AssertNotCalled(t, "CreateRunnerProcessShutdown", mock.Anything, mock.Anything)
	env.AssertNotCalled(t, "GetCurrentRunnerProcess", mock.Anything, mock.Anything)
}

func TestExecute_LegacyTemplateResolvesFromQueue_ShutsDownOwnProcess(t *testing.T) {
	var shutdownFor string
	env := runSignal(t, &Signal{RunnerID: runnerID, ProcessType: "install"}, func(env *testsuite.TestWorkflowEnvironment) {
		env.OnActivity("GetQueue", mock.Anything, queueID).
			Return(&app.Queue{ID: queueID, Name: queueName, OwnerID: runnerID}, nil).Once()
		env.OnActivity("GetRunnerProcess", mock.Anything, activities.GetRunnerProcessRequest{ProcessID: stalePID}).
			Return(activeProcess(stalePID), nil).Once()
		env.OnActivity("CreateRunnerProcessShutdown", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				shutdownFor = args.Get(1).(activities.CreateRunnerProcessShutdownRequest).RunnerProcessID
			}).
			Return(&app.RunnerProcessShutdown{}, nil).Once()
	})
	env.AssertExpectations(t)
	require.Equal(t, stalePID, shutdownFor)
}

func TestExecute_LegacyTemplateOnUnknownQueue_Noop(t *testing.T) {
	env := runSignal(t, &Signal{RunnerID: runnerID, ProcessType: "install"}, func(env *testsuite.TestWorkflowEnvironment) {
		env.OnActivity("GetQueue", mock.Anything, queueID).
			Return(&app.Queue{ID: queueID, Name: "management", OwnerID: runnerID}, nil).Once()
	})
	env.AssertExpectations(t)
	env.AssertNotCalled(t, "GetRunnerProcess", mock.Anything, mock.Anything)
	env.AssertNotCalled(t, "CreateRunnerProcessShutdown", mock.Anything, mock.Anything)
}

func TestExecute_ProcessBelongsToOtherRunner_Noop(t *testing.T) {
	env := runSignal(t, &Signal{RunnerID: "runother", ProcessType: "install", ProcessID: freshPID}, func(env *testsuite.TestWorkflowEnvironment) {
		env.OnActivity("GetRunnerProcess", mock.Anything, activities.GetRunnerProcessRequest{ProcessID: freshPID}).
			Return(activeProcess(freshPID), nil).Once()
	})
	env.AssertExpectations(t)
	env.AssertNotCalled(t, "CreateRunnerProcessShutdown", mock.Anything, mock.Anything)
}
