package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	workflowsflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow"
)

// GenerateSteps generates workflow steps and persists them.
//
// If the workflow has a GenerateStepsSignal, steps are generated via the signal
// path (see step_generate_signal.go). Otherwise falls back to the generators map
// for backward compatibility with conductor callers that haven't migrated yet.
func GenerateSteps(ctx workflow.Context, cfg StepConfig, flw *app.Workflow, generators map[app.WorkflowType]WorkflowStepGenerator) (*app.Workflow, error) {
	if flw.GenerateStepsSignal != nil && flw.GenerateStepsSignal.Signal != nil {
		return generateStepsViaSignal(ctx, cfg, flw)
	}

	return generateStepsViaGeneratorMap(ctx, flw, generators)
}

// GenerateEagerStepGroups generates and persists the eager step groups, allowing
// execution to begin on them while remaining groups are still generating.
// Only works for the signal-based path.
func GenerateEagerStepGroups(ctx workflow.Context, cfg StepConfig, flw *app.Workflow) (*EagerStepGroupsResult, error) {
	if flw.GenerateStepsSignal == nil || flw.GenerateStepsSignal.Signal == nil {
		return nil, errors.New("GenerateEagerStepGroups requires GenerateStepsSignal")
	}
	return generateEagerStepGroups(ctx, cfg, flw)
}

// CompleteStepGeneration fetches and persists all remaining step groups after
// eager generation. Safe to call even if eager generation was not used (no-op).
func CompleteStepGeneration(ctx workflow.Context, cfg StepConfig, flw *app.Workflow, queueSignalID string) (*app.Workflow, error) {
	return completeStepGeneration(ctx, cfg, flw, queueSignalID)
}

// generateStepsViaGeneratorMap uses the hardcoded Generators map to produce steps,
// then persists them via the GenerateWorkflowSteps child workflow.
// This is the legacy path used by conductor callers (apps, app-branches) that
// haven't migrated to GenerateStepsSignal yet.
func generateStepsViaGeneratorMap(ctx workflow.Context, flw *app.Workflow, generators map[app.WorkflowType]WorkflowStepGenerator) (*app.Workflow, error) {
	if generators == nil {
		return nil, errors.Errorf("no step generation method available for workflow %s", flw.ID)
	}

	gen, has := generators[flw.Type]
	if !has {
		return nil, errors.Errorf("no workflow step generator registered for workflow type %s", flw.Type)
	}

	result, err := gen(ctx, flw)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to generate steps for workflow %s", flw.ID)
	}

	steps, err := workflowsflow.AwaitGenerateWorkflowSteps(ctx, &workflowsflow.GenerateWorkflowStepsRequest{
		WorkflowID: flw.ID,
		Steps:      result.Steps,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create steps for workflow")
	}

	flw.Steps = make([]app.WorkflowStep, len(steps))
	for i, step := range steps {
		flw.Steps[i] = *step
	}

	return flw, nil
}
