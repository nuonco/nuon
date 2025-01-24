package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pkg/errors"

	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *planner) createStepPlan(ctx workflow.Context, step app.InstallActionWorkflowRunStep, installID string) (*plantypes.ActionWorkflowRunStepPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	plan := &plantypes.ActionWorkflowRunStepPlan{
		ID: step.ID,
		Attrs: map[string]string{
			"step.name": step.Step.Name,
			"step.id":   step.Step.ID,
		},
		InterpolatedEnvVars: make(map[string]string, 0),
	}

	// step 1 - fetch token for repo
	l.Debug("creating git source for config")
	// TODO(jm): not implemented, as requires changes from down stream executors, possibly.

	if len(step.Step.EnvVars) < 1 {
		return plan, nil
	}

	// step 2 - interpolate all variables in the set
	l.Debug("fetching install intermediate data")

	// NOTE(jm): this is no longer the best way to get this intermediate data. Jordan is adding a `get-state`
	// endpoint which we can use for this, and will replace this, long term.
	//
	intermediateData, err := activities.AwaitGetInstallIntermediateDataByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get intermediate data")
	}

	l.Debug(fmt.Sprintf("rendering %d env vars", len(step.Step.EnvVars)))
	intermediateDataPB, err := structpb.NewStruct(intermediateData.IntermediateData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create intermediate data protos")
	}

	for k, v := range step.Step.EnvVars {
		renderedVal, err := render.RenderString(*v, intermediateDataPB)
		if err != nil {
			l.Error(fmt.Sprintf("error rendering %s from intermediate data", *v),
				zap.Any("intermediate-data", intermediateData.IntermediateData))
			return nil, err
		}

		plan.InterpolatedEnvVars[k] = renderedVal
	}

	return plan, nil
}
