package testworker

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stretchr/testify/require"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateworkflowsteps"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	flowclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

const (
	pollTimeout  = 120 * time.Second
	pollInterval = 150 * time.Millisecond

	testResidentIdleTimeout = 5 * time.Second
)

func (e *FlowTestSuite) createTestQueue(ctx context.Context, ownerID, ownerType, queueName string) *app.Queue {
	return e.createTestQueueWithLimits(ctx, ownerID, ownerType, queueName, 20, 500)
}

func (e *FlowTestSuite) createTestQueueWithLimits(ctx context.Context, ownerID, ownerType, queueName string, maxInFlight, maxDepth int) *app.Queue {
	key := ownerID + "/" + ownerType + "/" + queueName
	if q, ok := e.queueCache[key]; ok {
		return q
	}

	q, err := e.service.QueueClient.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     ownerID,
		OwnerType:   ownerType,
		Namespace:   defaultNamespace,
		Name:        queueName,
		MaxInFlight: maxInFlight,
		MaxDepth:    maxDepth,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), q)

	require.Eventually(e.T(), func() bool {
		return e.service.QueueClient.QueueReady(ctx, q.ID) == nil
	}, pollTimeout, pollInterval, "queue %s did not become ready", q.ID)

	e.queueCache[key] = q
	return q
}

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

func (e *FlowTestSuite) getWorkflow(ctx context.Context, id string) *app.Workflow {
	var flw app.Workflow
	res := e.service.DB.WithContext(ctx).Preload("Steps").First(&flw, "id = ?", id)
	require.Nil(e.T(), res.Error)
	return &flw
}

func (e *FlowTestSuite) workflowStatus(ctx context.Context, id string) (app.Status, bool) {
	var flw app.Workflow
	if err := e.service.DB.WithContext(ctx).First(&flw, "id = ?", id).Error; err != nil {
		return "", false
	}
	return flw.Status.Status, true
}

func (e *FlowTestSuite) stepsByWorkflow(ctx context.Context, workflowID string) ([]app.WorkflowStep, error) {
	var steps []app.WorkflowStep
	err := e.service.DB.WithContext(ctx).
		Where("install_workflow_id = ?", workflowID).
		Order("idx ASC").
		Find(&steps).Error
	return steps, err
}

func (e *FlowTestSuite) getStep(ctx context.Context, id string) *app.WorkflowStep {
	var step app.WorkflowStep
	res := e.service.DB.WithContext(ctx).First(&step, "id = ?", id)
	require.Nil(e.T(), res.Error)
	return &step
}

func (e *FlowTestSuite) getStepGroup(ctx context.Context, id string) *app.WorkflowStepGroup {
	var group app.WorkflowStepGroup
	res := e.service.DB.WithContext(ctx).Where(app.WorkflowStepGroup{ID: id}).First(&group)
	require.Nil(e.T(), res.Error)
	return &group
}

func (e *FlowTestSuite) getStepsByWorkflow(ctx context.Context, workflowID string) []app.WorkflowStep {
	steps, err := e.tryStepsByWorkflow(ctx, workflowID)
	require.Nil(e.T(), err)
	return steps
}

func (e *FlowTestSuite) tryStepsByWorkflow(ctx context.Context, workflowID string) ([]app.WorkflowStep, error) {
	var steps []app.WorkflowStep
	err := e.service.DB.WithContext(ctx).
		Where("install_workflow_id = ?", workflowID).
		Order("idx ASC").
		Find(&steps).Error
	return steps, err
}

func fakeString() string {
	return generics.GetFakeObj[string]()
}

func newTestOwner() (string, string) {
	return fakeString(), "test_installs"
}

var testGenerators = struct {
	sync.Mutex
	byOwner map[string]map[app.WorkflowType]flow.WorkflowStepGenerator
}{byOwner: map[string]map[app.WorkflowType]flow.WorkflowStepGenerator{}}

var testGeneratorOwnerTypes = []string{
	"test_installs",
	"app_branches",
	"test_deadlock_installs",
	"test_layered_app_branches",
}

func registerTestGeneratorOwnerTypes() {
	testGenerators.Lock()
	defer testGenerators.Unlock()
	for _, ownerType := range testGeneratorOwnerTypes {
		testGenerators.byOwner[ownerType] = map[app.WorkflowType]flow.WorkflowStepGenerator{}
		generateworkflowsteps.RegisterGenerators(ownerType, func() map[app.WorkflowType]flow.WorkflowStepGenerator {
			testGenerators.Lock()
			defer testGenerators.Unlock()
			return maps.Clone(testGenerators.byOwner[ownerType])
		})
	}
}

func registerTestGenerator(ownerType string, workflowType app.WorkflowType, gen flow.WorkflowStepGenerator) {
	testGenerators.Lock()
	defer testGenerators.Unlock()
	gens, ok := testGenerators.byOwner[ownerType]
	if !ok {
		panic(fmt.Sprintf("owner type %q is not in testGeneratorOwnerTypes; add it so its factory is registered before cases run", ownerType))
	}
	gens[workflowType] = gen
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

func (e *FlowTestSuite) waitForWorkflowStatus(ctx context.Context, workflowID string, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		status, ok := e.workflowStatus(ctx, workflowID)
		return ok && status == expected
	}, pollTimeout, pollInterval, "workflow %s did not reach status %s", workflowID, expected)
}

func (e *FlowTestSuite) waitForWorkflowParked(ctx context.Context, workflowID string) {
	require.Eventually(e.T(), func() bool {
		flw := e.getWorkflow(ctx, workflowID)
		awaitingRetry, _ := flw.Status.Metadata["awaiting_retry"].(bool)
		return flw.Status.Status == app.StatusFailedPendingRetry && awaitingRetry
	}, pollTimeout, pollInterval, "workflow %s did not park awaiting retry", workflowID)
}

func (e *FlowTestSuite) waitForStepStatus(ctx context.Context, stepID string, expected app.Status) {
	require.Eventually(e.T(), func() bool {
		step := &app.WorkflowStep{}
		if err := e.service.DB.WithContext(ctx).First(step, "id = ?", stepID).Error; err != nil {
			return false
		}
		return step.Status.Status == expected
	}, pollTimeout, pollInterval, "step %s did not reach status %s", stepID, expected)
}

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

const ceilingWait = 45 * time.Second

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

func isTerminal(status app.Status) bool {
	switch status {
	case app.StatusSuccess, app.StatusError, app.StatusCancelled,
		app.StatusDiscarded, app.StatusUserSkipped, app.StatusAutoSkipped:
		return true
	}
	return false
}
