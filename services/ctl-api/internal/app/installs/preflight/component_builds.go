package preflight

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func plannedComponentBuilds(generated *app.GenerateStepsResult) []activities.PlannedComponentBuild {
	if generated == nil {
		return nil
	}

	requirements := make([]activities.PlannedComponentBuild, 0)
	seen := make(map[string]struct{})
	for _, step := range generated.Steps {
		if step == nil || step.ExecutionType == app.WorkflowStepExecutionTypeSkipped || step.QueueSignal == nil {
			continue
		}

		var requirement activities.PlannedComponentBuild
		switch sig := step.QueueSignal.Signal.(type) {
		case *componentdeploysyncandplan.Signal:
			requirement = activities.PlannedComponentBuild{
				ComponentID:                 sig.ComponentID,
				BuildID:                     sig.BuildID,
				ComponentConfigConnectionID: sig.ComponentConfigConnectionID,
				DeployID:                    sig.DeployID,
				WaitForBuild:                true,
			}
		case *componentsyncimage.Signal:
			requirement = activities.PlannedComponentBuild{
				ComponentID:                 sig.ComponentID,
				BuildID:                     sig.BuildID,
				ComponentConfigConnectionID: sig.ComponentConfigConnectionID,
				DeployID:                    sig.DeployID,
			}
		default:
			continue
		}

		key := fmt.Sprintf("%s|%s|%s|%s", requirement.ComponentID, requirement.BuildID, requirement.ComponentConfigConnectionID, requirement.DeployID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		requirements = append(requirements, requirement)
	}
	return requirements
}
