package workflows

import (
	"crypto/sha256"
	"encoding/hex"
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
// This workflow creates a single parallel step group with one build signal per component.
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
	sg.nextParallelGroup("build components")

	for _, componentID := range appConfig.ComponentIDs {
		build, err := activities.AwaitCreateComponentBuild(ctx, activities.CreateComponentBuildRequest{
			BuildID:     appConfigBuildComponentBuildID(flw.ID, componentID),
			ComponentID: componentID,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create build record for component %s", componentID)
		}

		name := fmt.Sprintf("build component %s", componentID)
		if n, ok := componentNames[componentID]; ok {
			name = fmt.Sprintf("build %s", n)
		}

		queues := componentQueues[componentID]
		step, err := sg.signalStep(ctx, componentID, "components", name, pgtype.Hstore{}, &inlinebuild.Signal{
			ComponentID: componentID,
			BuildID:     build.ID,
		},
			WithStepQueueID(queues.WorkflowStepsQueueID),
			WithTargetQueueID(queues.DefaultQueueID),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create build step for component %s", componentID)
		}
		steps = append(steps, step)
	}

	return sg.Result(steps), nil
}

func appConfigBuildComponentBuildID(workflowID, componentID string) string {
	sum := sha256.Sum256([]byte(workflowID + ":" + componentID))
	return "bld" + hex.EncodeToString(sum[:])[:23]
}
