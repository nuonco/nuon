package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GenerateEagerStepGroups generates and persists the eager step groups, allowing
// execution to begin on them while remaining groups are still generating.
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
