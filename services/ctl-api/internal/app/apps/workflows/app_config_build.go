package workflows

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	inlinebuild "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/inlinebuild"
)

// AppConfigBuild builds the workflow steps for an app config build.
// This workflow creates one step-group per component, each containing a build signal.
func AppConfigBuild(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	appConfigID := generics.FromPtrStr(flw.Metadata["app_config_id"])
	if appConfigID == "" {
		return nil, errors.New("app_config_id not found in workflow metadata")
	}

	appConfig, err := activities.AwaitGetAppConfigByIDByAppConfigID(ctx, appConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	if len(appConfig.ComponentIDs) == 0 {
		return &app.GenerateStepsResult{}, nil
	}

	// Look up component queue IDs for step routing.
	componentQueues := make(map[string]*componenthelpers.ComponentQueueIDs, len(appConfig.ComponentIDs))
	for _, componentID := range appConfig.ComponentIDs {
		queues, err := activities.AwaitGetComponentQueueIDsByComponentID(ctx, componentID)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get queue IDs for component %s", componentID)
		}
		componentQueues[componentID] = queues
	}

	// Look up component names by ID (sequential, 1-by-1).
	componentNames := make(map[string]string, len(appConfig.ComponentIDs))
	for _, componentID := range appConfig.ComponentIDs {
		cmp, err := activities.AwaitGetComponentByIDByComponentID(ctx, componentID)
		if err != nil {
			continue
		}
		componentNames[componentID] = cmp.Name
	}

	steps := make([]*app.WorkflowStep, 0, len(appConfig.ComponentIDs))
	sg := newStepGroup()

	// Batch components into parallel groups of 5.
	const batchSize = 5
	for batchStart := 0; batchStart < len(appConfig.ComponentIDs); batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > len(appConfig.ComponentIDs) {
			batchEnd = len(appConfig.ComponentIDs)
		}
		batch := appConfig.ComponentIDs[batchStart:batchEnd]

		sg.nextParallelGroup(fmt.Sprintf("build components %d-%d", batchStart+1, batchEnd))

		for _, componentID := range batch {
			name := fmt.Sprintf("build component %s", componentID)
			if n, ok := componentNames[componentID]; ok {
				name = fmt.Sprintf("build %s", n)
			}

			queues := componentQueues[componentID]
			step, err := sg.signalStep(ctx, componentID, "components", name, pgtype.Hstore{}, &inlinebuild.Signal{
				ComponentID: componentID,
			},
				WithStepQueueID(queues.WorkflowStepsQueueID),
				WithTargetQueueID(queues.DefaultQueueID),
			)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to create build step for component %s", componentID)
			}
			steps = append(steps, step)
		}
	}

	return sg.Result(steps), nil
}
