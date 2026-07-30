package triggerevent

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func routedResponse() *activities.RouteTriggerEventResponse {
	return &activities.RouteTriggerEventResponse{
		Dispatches: []activities.TriggerEventDispatchRef{
			{ID: "dispatch-1", AppID: "app-1", GenerationToken: "gen-1"},
			{ID: "dispatch-2", AppID: "app-2", GenerationToken: "gen-2"},
		},
		Waiters: []activities.EventRunbookWaiterRef{
			{ID: "waiter-1", OrgID: "org-1", QueueSignalID: "qs-1"},
			{ID: "waiter-2", OrgID: "org-1", QueueSignalID: "qs-2"},
		},
	}
}

func executeSignal(t *testing.T, env *testsuite.TestWorkflowEnvironment) error {
	t.Helper()
	sig := &Signal{EventID: "event-1"}
	env.ExecuteWorkflow(func(ctx workflow.Context) error { return sig.Execute(ctx) })
	require.True(t, env.IsWorkflowCompleted())
	return env.GetWorkflowError()
}

func TestExecuteFansOutEveryDispatchAndWaiter(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	env.OnActivity((*activities.Activities).RouteTriggerEvent, mock.Anything, mock.Anything, mock.Anything).
		Return(routedResponse(), nil).Once()
	env.OnActivity((*sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.Anything, mock.Anything).
		Return(&sharedactivities.EnqueueSignalToOwnerResponse{}, nil).Twice()
	env.OnActivity((*activities.Activities).NotifyEventRunbookWaiter, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Twice()

	require.NoError(t, executeSignal(t, env))
	env.AssertExpectations(t)
}

func TestExecuteFailedDispatchEnqueueDoesNotBlockSiblingsOrWaiters(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	env.OnActivity((*activities.Activities).RouteTriggerEvent, mock.Anything, mock.Anything, mock.Anything).
		Return(routedResponse(), nil).Once()
	env.OnActivity((*sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.Anything, mock.Anything).
		Return((*sharedactivities.EnqueueSignalToOwnerResponse)(nil), fmt.Errorf("enqueue failed"))
	env.OnActivity((*activities.Activities).NotifyEventRunbookWaiter, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Twice()

	err := executeSignal(t, env)
	require.ErrorContains(t, err, "enqueue failed")
	env.AssertExpectations(t)
	env.AssertNumberOfCalls(t, "EnqueueSignalToOwner", 10)
}

func TestExecuteFailedWaiterDoesNotBlockSiblingWaitersOrDispatches(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	// Register with a string name so the test env decodes the request payload
	// into (ctx, request) args; method-expression mocks treat the receiver as
	// arg 0 and leave the request zero-valued, breaking MatchedBy.
	env.RegisterActivityWithOptions((&activities.Activities{}).NotifyEventRunbookWaiter,
		activity.RegisterOptions{Name: "NotifyEventRunbookWaiter"})

	env.OnActivity((*activities.Activities).RouteTriggerEvent, mock.Anything, mock.Anything, mock.Anything).
		Return(routedResponse(), nil).Once()
	env.OnActivity((*sharedactivities.Activities).EnqueueSignalToOwner, mock.Anything, mock.Anything, mock.Anything).
		Return(&sharedactivities.EnqueueSignalToOwnerResponse{}, nil).Twice()
	env.OnActivity("NotifyEventRunbookWaiter", mock.Anything,
		mock.MatchedBy(func(req activities.NotifyEventRunbookWaiterRequest) bool { return req.WaiterID == "waiter-1" })).
		Return(fmt.Errorf("waiter notify failed"))
	env.OnActivity("NotifyEventRunbookWaiter", mock.Anything,
		mock.MatchedBy(func(req activities.NotifyEventRunbookWaiterRequest) bool { return req.WaiterID == "waiter-2" })).
		Return(nil).Once()

	err := executeSignal(t, env)
	require.ErrorContains(t, err, "waiter notify failed")
	env.AssertExpectations(t)
}
