package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func GenerateEagerStepGroups(ctx workflow.Context, cfg StepConfig, flw *app.Workflow) (*EagerStepGroupsResult, error) {
	if flw.GenerateStepsSignal == nil || flw.GenerateStepsSignal.Signal == nil {
		return nil, errors.New("GenerateEagerStepGroups requires GenerateStepsSignal")
	}
	return generateEagerStepGroups(ctx, cfg, flw)
}

func CompleteStepGeneration(ctx workflow.Context, cfg StepConfig, flw *app.Workflow, queueSignalID string) (*app.Workflow, error) {
	return completeStepGeneration(ctx, cfg, flw, queueSignalID)
}
