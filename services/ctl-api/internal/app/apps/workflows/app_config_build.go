package workflows

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	queuebuild "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/queuebuild"
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

	// Parse component state from app config to get names
	componentNames := parseComponentNames(appConfig.State)

	steps := make([]*app.WorkflowStep, 0, len(appConfig.ComponentIDs))
	sg := newStepGroup()

	for _, componentID := range appConfig.ComponentIDs {
		sg.nextGroup()

		name := fmt.Sprintf("build component %s", componentID)
		if n, ok := componentNames[componentID]; ok {
			name = fmt.Sprintf("build %s", n)
		}

		step, err := sg.signalStep(ctx, flw.OwnerID, "apps", name, pgtype.Hstore{}, &queuebuild.Signal{
			ComponentID: componentID,
			AppConfigID: appConfigID,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create build step for component %s", componentID)
		}
		steps = append(steps, step)
	}

	return sg.Result(steps), nil
}

// componentState is a minimal representation of the component state stored in app config.
type componentState struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// parseComponentNames extracts component names from the app config state JSON.
func parseComponentNames(state string) map[string]string {
	if state == "" {
		return nil
	}

	var components []componentState
	if err := json.Unmarshal([]byte(state), &components); err != nil {
		return nil
	}

	names := make(map[string]string, len(components))
	for _, c := range components {
		if c.ID != "" && c.Name != "" {
			names[c.ID] = c.Name
		}
	}
	return names
}
