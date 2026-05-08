package builds

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	queuebuild "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/queuebuild"
	componentdeploysyncandplan "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeworkflowstepgroup"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const buildBatchSize = 5

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Load the run to get the AppConfigID (set by the appconfig step)
	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if run.AppConfigID == "" {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

	// Get app config with component IDs
	appConfig, err := activities.AwaitGetAppConfigByIDByAppConfigID(ctx, run.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	l.Info("triggering builds",
		"app_branch_id", s.AppBranchID,
		"app_config_id", run.AppConfigID,
		"component_count", len(appConfig.ComponentIDs))

	if len(appConfig.ComponentIDs) == 0 {
		l.Info("no components to build")
		return nil
	}

	// Step 1: Build all components via step groups (must complete before sandbox deploys)
	if err := s.buildComponents(ctx, l, appConfig, run.AppConfigID); err != nil {
		return fmt.Errorf("component builds failed: %w", err)
	}

	// Step 2: Sandbox build — deploy components to sandbox installs using built artifacts
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	if len(branch.Configs) == 0 {
		l.Info("no branch config found, skipping sandbox build")
		return nil
	}

	if err := s.sandboxBuild(ctx, l, appConfig.ComponentIDs, &branch.Configs[0]); err != nil {
		return fmt.Errorf("sandbox builds failed: %w", err)
	}

	l.Info("all builds completed successfully")
	return nil
}

// buildComponents creates WorkflowStepGroups (batches of 5, parallel) with one
// WorkflowStep per component build, then dispatches each group through the flow
// engine and awaits completion.
func (s *Signal) buildComponents(ctx workflow.Context, l log.Logger, appConfig *app.AppConfig, appConfigID string) error {
	if s.FlowID == "" {
		return fmt.Errorf("builds signal has no flow ID — cannot create step groups")
	}

	componentIDs := appConfig.ComponentIDs

	// Look up component queue IDs for step routing.
	componentQueues := make(map[string]*componenthelpers.ComponentQueueIDs, len(componentIDs))
	for _, componentID := range componentIDs {
		queues, err := activities.AwaitGetComponentQueueIDsByComponentID(ctx, componentID)
		if err != nil {
			return fmt.Errorf("component %s: get queue IDs failed: %w", componentID, err)
		}
		componentQueues[componentID] = queues
	}

	// Look up component names by ID (sequential, 1-by-1).
	componentNames := make(map[string]string, len(componentIDs))
	for _, componentID := range componentIDs {
		cmp, err := activities.AwaitGetComponentByIDByComponentID(ctx, componentID)
		if err != nil {
			continue
		}
		componentNames[componentID] = cmp.Name
	}

	// Batch component IDs into groups of buildBatchSize.
	var groups []*app.WorkflowStepGroup
	var steps []*app.WorkflowStep
	groupIdx := 0

	for batchStart := 0; batchStart < len(componentIDs); batchStart += buildBatchSize {
		batchEnd := batchStart + buildBatchSize
		if batchEnd > len(componentIDs) {
			batchEnd = len(componentIDs)
		}
		batch := componentIDs[batchStart:batchEnd]

		groupID := domains.NewWorkflowStepGroupID()
		group := &app.WorkflowStepGroup{
			ID:         groupID,
			WorkflowID: s.FlowID,
			GroupIdx:   groupIdx,
			Parallel:   true,
			Name:       fmt.Sprintf("build components %d-%d", batchStart+1, batchEnd),
			Status:     app.CompositeStatus{Status: app.StatusPending},
		}
		groups = append(groups, group)

		for _, componentID := range batch {
			name := fmt.Sprintf("build component %s", componentID)
			if n, ok := componentNames[componentID]; ok {
				name = fmt.Sprintf("build %s", n)
			}

			queues := componentQueues[componentID]
			stepID := domains.NewWorkflowStepID()
			step := &app.WorkflowStep{
				ID:                  stepID,
				InstallWorkflowID:   s.FlowID,
				OwnerID:             componentID,
				OwnerType:           "components",
				Name:                name,
				ExecutionType:       app.WorkflowStepExecutionTypeSystem,
				Status:              app.NewCompositeTemporalStatus(ctx, app.StatusPending),
				GroupIdx:            groupIdx,
				WorkflowStepGroupID: groupID,
				Retryable:           true,
				Skippable:           true,
				StepQueueID:         queues.WorkflowStepsQueueID,
				TargetQueueID:       queues.DefaultQueueID,
				QueueSignal: &signaldb.SignalData{
					Signal: &queuebuild.Signal{
						ComponentID: componentID,
						AppConfigID: appConfigID,
					},
				},
			}

			// Inject step context so the inner signal knows its step and flow IDs.
			signal.ApplyStepContext(step.QueueSignal.Signal, stepID, s.FlowID)

			steps = append(steps, step)
		}

		groupIdx++
	}

	// Persist groups via activity.
	groupsReq := workflowactivities.CreateFlowStepGroupsRequest{
		Groups: make([]workflowactivities.CreateFlowStepGroup, 0, len(groups)),
	}
	for _, g := range groups {
		groupsReq.Groups = append(groupsReq.Groups, workflowactivities.CreateFlowStepGroup{
			ID:         g.ID,
			WorkflowID: s.FlowID,
			GroupIdx:   g.GroupIdx,
			Parallel:   g.Parallel,
			Name:       g.Name,
			Status:     g.Status,
		})
	}
	if _, err := workflowactivities.AwaitPkgWorkflowsFlowCreateFlowStepGroups(ctx, groupsReq); err != nil {
		return fmt.Errorf("unable to persist step groups: %w", err)
	}

	// Persist steps via activity.
	stepsReq := workflowactivities.CreateFlowStepsRequest{
		Steps: make([]workflowactivities.CreateFlowStep, 0, len(steps)),
	}
	for idx, step := range steps {
		stepsReq.Steps = append(stepsReq.Steps, workflowactivities.CreateFlowStep{
			ID:                  step.ID,
			FlowID:              s.FlowID,
			OwnerID:             step.OwnerID,
			OwnerType:           step.OwnerType,
			Status:              step.Status,
			Name:                step.Name,
			QueueSignal:         step.QueueSignal,
			Idx:                 idx * 100,
			ExecutionType:       step.ExecutionType,
			Retryable:           step.Retryable,
			Skippable:           step.Skippable,
			GroupIdx:            step.GroupIdx,
			WorkflowStepGroupID: step.WorkflowStepGroupID,
			StepQueueID:         step.StepQueueID,
			TargetQueueID:       step.TargetQueueID,
		})
	}
	if _, err := workflowactivities.AwaitPkgWorkflowsFlowCreateFlowSteps(ctx, stepsReq); err != nil {
		return fmt.Errorf("unable to persist workflow steps: %w", err)
	}

	// Dispatch each group and await completion sequentially.
	for _, group := range groups {
		l.Info("dispatching build group",
			"group_idx", group.GroupIdx,
			"group_name", group.Name,
			"step_group_id", group.ID,
			"parallel", group.Parallel)

		if err := s.dispatchAndAwaitGroup(ctx, group); err != nil {
			return fmt.Errorf("build group %q failed: %w", group.Name, err)
		}
	}

	l.Info("all component builds completed successfully")
	return nil
}

// dispatchAndAwaitGroup enqueues an executeworkflowstepgroup signal for a build
// group and waits for it to complete.
func (s *Signal) dispatchAndAwaitGroup(ctx workflow.Context, group *app.WorkflowStepGroup) error {
	sig := &executeworkflowstepgroup.Signal{
		WorkflowID:  s.FlowID,
		StepGroupID: group.ID,
		GroupIdx:    group.GroupIdx,
		OwnerID:     s.AppBranchID,
		OwnerType:   "app_branches",
		QueueName:   "app-branch-workflow-steps",
		Parallel:    group.Parallel,
		// TargetQueueName is left empty — per-step routing in dispatchStep
		// uses the step's OwnerID (componentID) and dispatches to its default queue.
	}

	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         s.AppBranchID,
		OwnerType:       "app_branches",
		QueueName:       "app-branch-workflow-step-groups",
		Signal:          sig,
		SignalOwnerID:   group.ID,
		SignalOwnerType: "workflow_step_groups",
	})
	if err != nil {
		return fmt.Errorf("unable to enqueue group signal: %w", err)
	}

	_, err = queueclient.AwaitQueueSignal(ctx, enqueueResp.QueueSignalID)
	if err != nil {
		if ctx.Err() != nil {
			cancelCtx, cancelCtxCancel := workflow.NewDisconnectedContext(ctx)
			defer cancelCtxCancel()
			queueclient.AwaitCancelSignal(cancelCtx, enqueueResp.QueueSignalID)
		}
		return fmt.Errorf("group signal failed: %w", err)
	}

	return nil
}

// sandboxBuild deploys built components to sandbox installs across all install groups in parallel.
func (s *Signal) sandboxBuild(ctx workflow.Context, l log.Logger, componentIDs []string, config *app.AppBranchConfig) error {
	if len(config.InstallGroups) == 0 {
		l.Info("no install groups, skipping sandbox build")
		return nil
	}

	for _, group := range config.InstallGroups {
		if len(group.InstallIDs) == 0 {
			l.Info("no installs in group, skipping", "group_name", group.Name)
			continue
		}

		errCh := workflow.NewChannel(ctx)
		pending := len(group.InstallIDs)

		for _, installID := range group.InstallIDs {
			installID := installID
			workflow.Go(ctx, func(gCtx workflow.Context) {
				deployErr := s.sandboxBuildForInstall(gCtx, l, installID, componentIDs)
				errCh.Send(gCtx, deployErr)
			})
		}

		var errs []error
		for i := 0; i < pending; i++ {
			var deployErr error
			errCh.Receive(ctx, &deployErr)
			if deployErr != nil {
				errs = append(errs, deployErr)
			}
		}

		if len(errs) > 0 {
			return fmt.Errorf("sandbox build group %q had %d error(s): %v", group.Name, len(errs), errs)
		}
	}

	l.Info("sandbox builds completed successfully")
	return nil
}

// sandboxBuildForInstall triggers componentdeploysyncandplan (sandbox mode) for each component in an install.
func (s *Signal) sandboxBuildForInstall(ctx workflow.Context, l log.Logger, installID string, componentIDs []string) error {
	mappingResp, err := activities.AwaitGetInstallComponentsByComponentIDs(ctx, activities.GetInstallComponentsByComponentIDsRequest{
		Req: &activities.GetInstallComponentsByComponentIDsInput{
			InstallID:    installID,
			ComponentIDs: componentIDs,
		},
	})
	if err != nil {
		return fmt.Errorf("install %s: unable to get install components: %w", installID, err)
	}

	installComponentMap := make(map[string]string, len(mappingResp.Mappings))
	for _, m := range mappingResp.Mappings {
		installComponentMap[m.ComponentID] = m.InstallComponentID
	}

	for _, componentID := range componentIDs {
		installComponentID, ok := installComponentMap[componentID]
		if !ok {
			l.Warn("install component mapping not found, skipping",
				"install_id", installID,
				"component_id", componentID)
			continue
		}

		enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   installID,
			OwnerType: "installs",
			QueueName: "install-signals",
			Signal: &componentdeploysyncandplan.Signal{
				SandboxMode:        true,
				ComponentID:        componentID,
				InstallComponentID: installComponentID,
			},
		})
		if err != nil {
			return fmt.Errorf("install %s component %s: enqueue failed: %w", installID, componentID, err)
		}

		l.Info("waiting for sandbox deploy to complete",
			"install_id", installID,
			"component_id", componentID,
			"install_component_id", installComponentID,
			"queue_signal_id", enqueueResp.QueueSignalID)

		if _, err = queueclient.AwaitQueueSignal(ctx, enqueueResp.QueueSignalID); err != nil {
			return fmt.Errorf("install %s component %s: sandbox deploy failed: %w", installID, componentID, err)
		}

		l.Info("sandbox deploy completed", "install_id", installID, "component_id", componentID)
	}

	return nil
}
