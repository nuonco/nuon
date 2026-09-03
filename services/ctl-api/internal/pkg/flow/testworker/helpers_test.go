package testworker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

const (
	pollTimeout  = 60 * time.Second
	pollInterval = 150 * time.Millisecond
)

// testQueueCache memoizes queues so each is created and readiness-polled once per run.
var (
	testQueueCacheMu sync.Mutex
	testQueueCache   = map[string]*app.Queue{}
)

func (e *FlowTestSuite) createTestQueue(ctx context.Context, ownerID, ownerType, queueName string) *app.Queue {
	key := ownerID + "/" + ownerType + "/" + queueName
	testQueueCacheMu.Lock()
	defer testQueueCacheMu.Unlock()
	if q, ok := testQueueCache[key]; ok {
		return q
	}

	q, err := e.service.QueueClient.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     ownerID,
		OwnerType:   ownerType,
		Namespace:   defaultNamespace,
		Name:        queueName,
		MaxInFlight: 20,
		MaxDepth:    500,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), q)

	// QueueReady may fail transiently while the queue workflow registers its
	// query handlers. Retry until it succeeds or the timeout expires.
	require.Eventually(e.T(), func() bool {
		return e.service.QueueClient.QueueReady(ctx, q.ID) == nil
	}, pollTimeout, pollInterval, "queue %s did not become ready", q.ID)

	testQueueCache[key] = q
	return q
}

// createTestWorkflow creates a workflow with a generate-steps signal.
func (e *FlowTestSuite) createTestWorkflow(ctx context.Context, ownerID, ownerType string, wfType app.WorkflowType, genSignal signal.Signal) *app.Workflow {
	flw := app.Workflow{
		OwnerID:   ownerID,
		OwnerType: ownerType,
		Type:      wfType,
		Status:    app.NewCompositeStatus(ctx, app.StatusPending),
		GenerateStepsSignal: &signaldb.SignalData{
			Signal: genSignal,
		},
	}
	res := e.service.DB.WithContext(ctx).Create(&flw)
	require.Nil(e.T(), res.Error)
	return &flw
}

// createTestSteps creates workflow steps for the given workflow.
func (e *FlowTestSuite) createTestSteps(ctx context.Context, flw *app.Workflow, steps []app.WorkflowStep) {
	groups := make(map[int]string)
	for i := range steps {
		steps[i].InstallWorkflowID = flw.ID
		steps[i].OwnerID = flw.OwnerID
		steps[i].OwnerType = flw.OwnerType
		if steps[i].Status.Status == "" {
			steps[i].Status = app.NewCompositeStatus(ctx, app.StatusPending)
		}
		if steps[i].WorkflowStepGroupID == "" {
			groupID, ok := groups[steps[i].GroupIdx]
			if !ok {
				group := app.WorkflowStepGroup{
					WorkflowID: flw.ID,
					GroupIdx:   steps[i].GroupIdx,
					Parallel:   steps[i].GroupParallel,
					Status:     app.NewCompositeStatus(ctx, app.StatusPending),
				}
				res := e.service.DB.WithContext(ctx).Create(&group)
				require.Nil(e.T(), res.Error)
				groupID = group.ID
				groups[steps[i].GroupIdx] = groupID
			}
			steps[i].WorkflowStepGroupID = groupID
		}
	}
	res := e.service.DB.WithContext(ctx).Create(&steps)
	require.Nil(e.T(), res.Error)
}

// getWorkflow re-fetches a workflow from DB.
func (e *FlowTestSuite) getWorkflow(ctx context.Context, id string) *app.Workflow {
	var flw app.Workflow
	res := e.service.DB.WithContext(ctx).Preload("Steps").First(&flw, "id = ?", id)
	require.Nil(e.T(), res.Error)
	return &flw
}

// getStep re-fetches a workflow step from DB.
func (e *FlowTestSuite) getStep(ctx context.Context, id string) *app.WorkflowStep {
	var step app.WorkflowStep
	res := e.service.DB.WithContext(ctx).First(&step, "id = ?", id)
	require.Nil(e.T(), res.Error)
	return &step
}

// getStepsByWorkflow fetches all steps for a workflow ordered by Idx.
func (e *FlowTestSuite) getStepsByWorkflow(ctx context.Context, workflowID string) []app.WorkflowStep {
	var steps []app.WorkflowStep
	res := e.service.DB.WithContext(ctx).
		Where("install_workflow_id = ?", workflowID).
		Order("idx ASC").
		Find(&steps)
	require.Nil(e.T(), res.Error)
	return steps
}

// Dispatch resolves queues by (flow owner, queue name), so one run-scoped owner lets all tests reuse one queue set; assertions key on workflow/step IDs, never owner.
var sharedTestOwnerID = generics.GetFakeObj[string]()

func newTestOwner() (string, string) {
	return sharedTestOwnerID, "test_installs"
}

func (e *FlowTestSuite) getLatestQueueSignal(ctx context.Context, ownerID, ownerType string, signalType signal.SignalType) *app.QueueSignal {
	var queueSignal app.QueueSignal
	res := e.service.DB.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   ownerID,
			OwnerType: ownerType,
			Type:      signalType,
		}).
		Order("created_at DESC").
		First(&queueSignal)
	require.Nil(e.T(), res.Error)
	return &queueSignal
}

func (e *FlowTestSuite) waitForQueueSignalStatus(ctx context.Context, ownerID, ownerType string, signalType signal.SignalType, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		queueSignal := e.getLatestQueueSignal(ctx, ownerID, ownerType, signalType)
		return queueSignal.Status.Status == expected
	}, pollTimeout, pollInterval, "queue signal %s for %s did not reach %s", signalType, ownerID, expected)
}

func (e *FlowTestSuite) getWorkflowRuns(ctx context.Context, workflowID string) []app.WorkflowRun {
	var runs []app.WorkflowRun
	res := e.service.DB.WithContext(ctx).
		Where(app.WorkflowRun{WorkflowID: workflowID}).
		Order("created_at ASC").
		Find(&runs)
	require.Nil(e.T(), res.Error)
	return runs
}

// waitForWorkflowStatus polls until the workflow reaches the expected status.
func (e *FlowTestSuite) waitForWorkflowStatus(ctx context.Context, workflowID string, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		flw := e.getWorkflow(ctx, workflowID)
		return flw.Status.Status == expected
	}, pollTimeout, pollInterval, "workflow %s did not reach status %s", workflowID, expected)
}

// waitForStepStatus polls until the step reaches the expected status.
func (e *FlowTestSuite) waitForStepStatus(ctx context.Context, stepID string, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		step := e.getStep(ctx, stepID)
		return step.Status.Status == expected
	}, pollTimeout, pollInterval, "step %s did not reach status %s", stepID, expected)
}

// waitForStepInProgress waits until a step with the given name is in-progress
// and returns its ID.
func (e *FlowTestSuite) waitForStepInProgress(ctx context.Context, workflowID, stepName string) string {
	var stepID string
	require.Eventually(e.T(), func() bool {
		steps := e.getStepsByWorkflow(ctx, workflowID)
		for _, s := range steps {
			if s.Name == stepName && s.Status.Status == app.StatusInProgress {
				stepID = s.ID
				return true
			}
		}
		return false
	}, pollTimeout, pollInterval)
	return stepID
}

// waitForWorkflowTerminal polls until the workflow reaches any terminal status.
func (e *FlowTestSuite) waitForWorkflowTerminal(ctx context.Context, workflowID string) {
	require.Eventually(e.T(), func() bool {
		flw := e.getWorkflow(ctx, workflowID)
		switch flw.Status.Status {
		case app.StatusSuccess, app.StatusError, app.StatusCancelled:
			return true
		}
		return false
	}, pollTimeout, pollInterval, "workflow %s did not reach a terminal status", workflowID)
}

// ceilingWait bounds asserts that depend on the MaxWaitCeiling override (15s
// in SetupSuite) firing, with margin for scheduling.
const ceilingWait = 45 * time.Second

// flowTemporalRefs collects the Temporal workflow refs of every queue signal
// owned by the flow, its groups, or its steps (including retry clones).
func (e *FlowTestSuite) flowTemporalRefs(ctx context.Context, workflowID string) []signaldb.WorkflowRef {
	ownerIDs := []string{workflowID}
	var groups []app.WorkflowStepGroup
	res := e.service.DB.WithContext(ctx).
		Where(app.WorkflowStepGroup{WorkflowID: workflowID}).
		Find(&groups)
	require.Nil(e.T(), res.Error)
	for _, g := range groups {
		ownerIDs = append(ownerIDs, g.ID)
	}
	for _, s := range e.getStepsByWorkflow(ctx, workflowID) {
		ownerIDs = append(ownerIDs, s.ID)
	}

	var queueSignals []app.QueueSignal
	res = e.service.DB.WithContext(ctx).
		Where("owner_id IN ?", ownerIDs).
		Find(&queueSignals)
	require.Nil(e.T(), res.Error)

	var refs []signaldb.WorkflowRef
	for _, qs := range queueSignals {
		if qs.Workflow.ID != "" {
			refs = append(refs, qs.Workflow)
		}
	}
	return refs
}

// assertTemporalDrained waits until every Temporal workflow backing the flow's
// queue signals is closed — a stopped flow must not hold handlers open.
func (e *FlowTestSuite) assertTemporalDrained(ctx context.Context, workflowID string) {
	require.Eventually(e.T(), func() bool {
		refs := e.flowTemporalRefs(ctx, workflowID)
		for _, ref := range refs {
			resp, err := e.service.TClient.DescribeWorkflowExecutionInNamespace(ctx, ref.Namespace, ref.ID, "")
			if err != nil {
				var notFound *serviceerror.NotFound
				if errors.As(err, &notFound) {
					continue
				}
				return false
			}
			if resp.GetWorkflowExecutionInfo().GetStatus() == enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
				return false
			}
		}
		return true
	}, ceilingWait, pollInterval, "temporal workflows for flow %s did not drain", workflowID)
}

// isTerminal returns true if the status is a terminal status for a step.
func isTerminal(status app.Status) bool {
	switch status {
	case app.StatusSuccess, app.StatusError, app.StatusCancelled,
		app.StatusDiscarded, app.StatusUserSkipped, app.StatusAutoSkipped:
		return true
	}
	return false
}
