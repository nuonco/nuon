package testworker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

const (
	pollTimeout  = 120 * time.Second
	pollInterval = 150 * time.Millisecond
)

// testQueueCache is per-case (each FlowTestSuite value is one case), so no
// locking is needed and unrelated cases never wait on each other's readiness
// polling.
func (e *FlowTestSuite) createTestQueue(ctx context.Context, ownerID, ownerType, queueName string) *app.Queue {
	key := ownerID + "/" + ownerType + "/" + queueName
	if q, ok := e.queueCache[key]; ok {
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

	e.queueCache[key] = q
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

// workflowStatus reads only the workflow row (no Preload) for polling loops.
func (e *FlowTestSuite) workflowStatus(ctx context.Context, id string) (app.Status, bool) {
	var flw app.Workflow
	if err := e.service.DB.WithContext(ctx).First(&flw, "id = ?", id).Error; err != nil {
		return "", false
	}
	return flw.Status.Status, true
}

// stepsByWorkflow is the error-returning read for use inside poll callbacks.
func (e *FlowTestSuite) stepsByWorkflow(ctx context.Context, workflowID string) ([]app.WorkflowStep, error) {
	var steps []app.WorkflowStep
	err := e.service.DB.WithContext(ctx).
		Where("install_workflow_id = ?", workflowID).
		Order("idx ASC").
		Find(&steps).Error
	return steps, err
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
	steps, err := e.tryStepsByWorkflow(ctx, workflowID)
	require.Nil(e.T(), err)
	return steps
}

// tryStepsByWorkflow is the error-tolerant form of getStepsByWorkflow for
// polling conditions: a failed query (e.g. context canceled during cleanup)
// must not fail the test from a goroutine that outlived it.
func (e *FlowTestSuite) tryStepsByWorkflow(ctx context.Context, workflowID string) ([]app.WorkflowStep, error) {
	var steps []app.WorkflowStep
	err := e.service.DB.WithContext(ctx).
		Where("install_workflow_id = ?", workflowID).
		Order("idx ASC").
		Find(&steps).Error
	return steps, err
}

// fakeString serializes go-faker access: faker mutates package-global state
// and is not safe for concurrent cases.
var fakeMu sync.Mutex

func fakeString() string {
	fakeMu.Lock()
	defer fakeMu.Unlock()
	return generics.GetFakeObj[string]()
}

func newTestOwner() (string, string) {
	return fakeString(), "test_installs"
}

// The waitFor* helpers poll with pure-bool conditions: testify runs
// Eventually callbacks on a separate goroutine, so a require.* FailNow inside
// a callback fires on the wrong goroutine (and can outlive the timeout).
// Transient read errors are treated as "not yet" and retried until timeout.

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
		var queueSignal app.QueueSignal
		err := e.service.DB.WithContext(ctx).
			Where(app.QueueSignal{
				OwnerID:   ownerID,
				OwnerType: ownerType,
				Type:      signalType,
			}).
			Order("created_at DESC").
			First(&queueSignal).Error
		if err != nil {
			return false
		}
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
// Status-only read: no Preload — the poller hits this every 150ms per case.
func (e *FlowTestSuite) waitForWorkflowStatus(ctx context.Context, workflowID string, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		status, ok := e.workflowStatus(ctx, workflowID)
		return ok && status == expected
	}, pollTimeout, pollInterval, "workflow %s did not reach status %s", workflowID, expected)
}

// waitForStepStatus polls until the step reaches the expected status.
func (e *FlowTestSuite) waitForStepStatus(ctx context.Context, stepID string, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		step := &app.WorkflowStep{}
		if err := e.service.DB.WithContext(ctx).First(step, "id = ?", stepID).Error; err != nil {
			return false
		}
		return step.Status.Status == expected
	}, pollTimeout, pollInterval, "step %s did not reach status %s", stepID, expected)
}

// waitForStepInProgress waits until a step with the given name is in-progress
// and returns its ID.
func (e *FlowTestSuite) waitForStepInProgress(ctx context.Context, workflowID, stepName string) string {
	var found atomic.Pointer[string]
	require.Eventually(e.T(), func() bool {
		steps, err := e.stepsByWorkflow(ctx, workflowID)
		if err != nil {
			return false
		}
		for _, s := range steps {
			if s.Name == stepName && s.Status.Status == app.StatusInProgress {
				id := s.ID
				found.Store(&id)
				return true
			}
		}
		return false
	}, pollTimeout, pollInterval)
	if id := found.Load(); id != nil {
		return *id
	}
	require.FailNow(e.T(), "step %s never reached in-progress", stepName)
	return ""
}

// waitForWorkflowTerminal polls until the workflow reaches any terminal status.
func (e *FlowTestSuite) waitForWorkflowTerminal(ctx context.Context, workflowID string) {
	require.Eventually(e.T(), func() bool {
		status, ok := e.workflowStatus(ctx, workflowID)
		if !ok {
			return false
		}
		switch status {
		case app.StatusSuccess, app.StatusError, app.StatusCancelled:
			return true
		}
		return false
	}, pollTimeout, pollInterval, "workflow %s did not reach a terminal status", workflowID)
}

func (e *FlowTestSuite) waitForWorkflowFinished(ctx context.Context, workflowID string) {
	require.Eventually(e.T(), func() bool {
		var flw app.Workflow
		if err := e.service.DB.WithContext(ctx).First(&flw, "id = ?", workflowID).Error; err != nil {
			return false
		}
		return !flw.FinishedAt.IsZero()
	}, pollTimeout, pollInterval, "workflow %s did not set finished_at", workflowID)
}

func (e *FlowTestSuite) cancelWorkflow(ctx context.Context, workflowID string) {
	_, err := e.service.FlowClient.CancelWorkflow(ctx, &flowclient.CancelWorkflowRequest{
		InstallWorkflowID: workflowID,
	})
	require.NoError(e.T(), err)
	e.waitForWorkflowStatus(ctx, workflowID, app.StatusCancelled)
}

// ceilingWait bounds asserts that depend on the MaxWaitCeiling override (5s
// in SetupSuite) firing, with margin for scheduling.
const ceilingWait = 45 * time.Second

// flowTemporalRefs collects the Temporal workflow refs of every queue signal
// owned by the flow, its groups, or its steps (including retry clones).
func (e *FlowTestSuite) flowTemporalRefs(ctx context.Context, workflowID string) ([]signaldb.WorkflowRef, error) {
	ownerIDs := []string{workflowID}
	var groups []app.WorkflowStepGroup
	if err := e.service.DB.WithContext(ctx).
		Where(app.WorkflowStepGroup{WorkflowID: workflowID}).
		Find(&groups).Error; err != nil {
		return nil, err
	}
	for _, g := range groups {
		ownerIDs = append(ownerIDs, g.ID)
	}
	steps, err := e.stepsByWorkflow(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	for _, s := range steps {
		ownerIDs = append(ownerIDs, s.ID)
	}

	var queueSignals []app.QueueSignal
	if err := e.service.DB.WithContext(ctx).
		Where("owner_id IN ?", ownerIDs).
		Find(&queueSignals).Error; err != nil {
		return nil, err
	}

	var refs []signaldb.WorkflowRef
	for _, qs := range queueSignals {
		if qs.Workflow.ID != "" {
			refs = append(refs, qs.Workflow)
		}
	}
	return refs, nil
}

// assertTemporalDrained waits until every Temporal workflow backing the flow's
// queue signals is closed — a stopped flow must not hold handlers open.
func (e *FlowTestSuite) assertTemporalDrained(ctx context.Context, workflowID string) {
	require.Eventually(e.T(), func() bool {
		refs, err := e.flowTemporalRefs(ctx, workflowID)
		if err != nil {
			return false
		}
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
