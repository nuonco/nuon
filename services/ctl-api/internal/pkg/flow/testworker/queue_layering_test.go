package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateworkflowsteps"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type flowQueueNames struct {
	stepGroups    string
	steps         string
	stepTargets   string
	generateSteps string
}

func (e *FlowTestSuite) enqueueFlowWithQueues(ctx context.Context, queueID string, flw *app.Workflow, names flowQueueNames) {
	resp, err := e.service.QueueClient.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queueID,
		Signal: &executeflow.Signal{
			WorkflowID:             flw.ID,
			StepGroupQueueName:     names.stepGroups,
			StepQueueName:          names.steps,
			StepTargetQueueName:    names.stepTargets,
			GenerateStepsQueueName: names.generateSteps,
			OwnerID:                flw.OwnerID,
			OwnerType:              flw.OwnerType,
		},
		OwnerID:   flw.ID,
		OwnerType: "install_workflows",
	})
	require.NoError(e.T(), err)
	require.NotNil(e.T(), resp)
}

func registerLayeringGenerator(ownerType string, workflowType app.WorkflowType) {
	generateworkflowsteps.RegisterGenerators(ownerType, func() map[app.WorkflowType]flow.WorkflowStepGenerator {
		return map[app.WorkflowType]flow.WorkflowStepGenerator{
			workflowType: func(workflow.Context, *app.Workflow) (*app.GenerateStepsResult, error) {
				return &app.GenerateStepsResult{
					Groups: []*app.WorkflowStepGroup{{
						GroupIdx: 1,
						Status:   app.CompositeStatus{Status: app.StatusPending},
					}},
					Steps: []*app.WorkflowStep{{
						Name:          "layered-success",
						Idx:           100,
						GroupIdx:      1,
						ExecutionType: app.WorkflowStepExecutionTypeSystem,
						Status:        app.CompositeStatus{Status: app.StatusPending},
						QueueSignal:   &signaldb.SignalData{Signal: &SuccessSignal{}},
					}},
				}, nil
			},
		}
	})
}

func (e *FlowTestSuite) TestOldQueueLayoutWedgesAtCapacity() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := fakeString(), "test_deadlock_installs"
	workflowType := app.WorkflowType("test_old_queue_layout_deadlock")
	registerLayeringGenerator(ownerType, workflowType)

	sharedQueue := e.createTestQueueWithLimits(ctx, ownerID, ownerType, "shared-workflows-and-generation", 2, 50)
	names := flowQueueNames{
		stepGroups:    "test-step-groups",
		steps:         "test-steps",
		stepTargets:   "test-signals",
		generateSteps: sharedQueue.Name,
	}
	e.createTestQueueWithLimits(ctx, ownerID, ownerType, names.stepGroups, 1, 50)
	e.createTestQueueWithLimits(ctx, ownerID, ownerType, names.steps, 1, 50)
	e.createTestQueueWithLimits(ctx, ownerID, ownerType, names.stepTargets, 1, 50)

	flows := []*app.Workflow{
		e.createTestWorkflow(ctx, ownerID, ownerType, workflowType, &generateworkflowsteps.Signal{}),
		e.createTestWorkflow(ctx, ownerID, ownerType, workflowType, &generateworkflowsteps.Signal{}),
	}
	for _, flw := range flows {
		e.enqueueFlowWithQueues(ctx, sharedQueue.ID, flw, names)
	}

	require.Eventually(e.T(), func() bool {
		var executing int64
		var generationQueued int64
		err := e.service.DB.WithContext(ctx).Model(&app.QueueSignal{}).
			Where("queue_id = ? AND type = ? AND status->>'status' = ?", sharedQueue.ID, executeflow.SignalType, app.StatusInProgress).
			Count(&executing).Error
		if err != nil {
			return false
		}
		err = e.service.DB.WithContext(ctx).Model(&app.QueueSignal{}).
			Where("queue_id = ? AND type = ? AND status->>'status' = ?", sharedQueue.ID, generateworkflowsteps.SignalType, app.StatusQueued).
			Count(&generationQueued).Error
		return err == nil && executing == 2 && generationQueued == 2
	}, 30*time.Second, pollInterval)

	require.Never(e.T(), func() bool {
		var count int64
		err := e.service.DB.WithContext(ctx).Model(&app.WorkflowStep{}).
			Where("install_workflow_id IN ?", []string{flows[0].ID, flows[1].ID}).
			Count(&count).Error
		return err == nil && count > 0
	}, 3*time.Second, pollInterval)

	cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, flw := range flows {
		refs, err := e.flowTemporalRefs(cleanupCtx, flw.ID)
		require.NoError(e.T(), err)
		for _, ref := range refs {
			namespaceClient, err := e.service.TClient.GetNamespaceClient(ref.Namespace)
			require.NoError(e.T(), err)
			_ = namespaceClient.TerminateWorkflow(cleanupCtx, ref.ID, "", "deadlock reproduction cleanup")
		}
	}
	namespaceClient, err := e.service.TClient.GetNamespaceClient(sharedQueue.Workflow.Namespace)
	require.NoError(e.T(), err)
	require.NoError(e.T(), namespaceClient.TerminateWorkflow(cleanupCtx, sharedQueue.Workflow.ID, "", "deadlock reproduction cleanup"))
}

func (e *FlowTestSuite) TestLayeredQueuesDrainAtSingleConcurrency() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())
	ownerID, ownerType := fakeString(), "test_layered_app_branches"
	workflowType := app.WorkflowType("test_layered_queue_layout")
	registerLayeringGenerator(ownerType, workflowType)

	names := flowQueueNames{
		stepGroups:    queuenames.AppBranchWorkflowStepGroupsQueueName,
		steps:         queuenames.AppBranchWorkflowStepsQueueName,
		stepTargets:   queuenames.AppBranchSignalsQueueName,
		generateSteps: queuenames.AppBranchGenerateStepsQueueName,
	}
	workflowQueue := e.createTestQueueWithLimits(ctx, ownerID, ownerType, queuenames.AppBranchWorkflowsQueueName, 1, 50)
	e.createTestQueueWithLimits(ctx, ownerID, ownerType, names.stepGroups, 1, 50)
	e.createTestQueueWithLimits(ctx, ownerID, ownerType, names.steps, 1, 50)
	e.createTestQueueWithLimits(ctx, ownerID, ownerType, names.stepTargets, 1, 50)
	e.createTestQueueWithLimits(ctx, ownerID, ownerType, names.generateSteps, 1, 50)

	flows := make([]*app.Workflow, 8)
	for idx := range flows {
		flows[idx] = e.createTestWorkflow(ctx, ownerID, ownerType, workflowType, &generateworkflowsteps.Signal{})
		e.enqueueFlowWithQueues(ctx, workflowQueue.ID, flows[idx], names)
	}

	for _, flw := range flows {
		e.waitForWorkflowStatus(ctx, flw.ID, app.StatusSuccess)
		e.assertTemporalDrained(ctx, flw.ID)
	}
}
