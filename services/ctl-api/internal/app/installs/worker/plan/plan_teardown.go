package plan

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/pkg/terraform/teardown"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
)

// teardownRetain returns the retain entries for a component teardown. The
// install's current app config wins so a vendor can change what is retained
// without redeploying; the deployed build's config covers components that
// have since been removed from the app config.
func teardownRetain(appCfg *app.AppConfig, build *app.ComponentBuild) []string {
	if appCfg != nil {
		for _, ccc := range appCfg.ComponentConfigConnections {
			if ccc.ComponentID == build.ComponentConfigConnection.ComponentID {
				return ccc.GetTeardownRetain()
			}
		}
	}
	return build.ComponentConfigConnection.GetTeardownRetain()
}

func (p *Planner) createTeardownPlan(
	ctx workflow.Context,
	appCfg *app.AppConfig,
	build *app.ComponentBuild,
	installState *state.State,
) (*plantypes.TeardownPlan, error) {
	retain := teardownRetain(appCfg, build)
	if len(retain) == 0 {
		return nil, nil
	}

	stateData, err := installState.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get state")
	}

	return RenderTeardownPlan(retain, stateData)
}

// RenderTeardownPlan renders each retain entry, dropping entries that render
// to an empty string so templates can conditionally retain resources.
func RenderTeardownPlan(retain []string, stateData map[string]any) (*plantypes.TeardownPlan, error) {
	rendered := make([]string, 0, len(retain))
	seen := make(map[string]struct{}, len(retain))
	for idx, entry := range retain {
		out, err := render.RenderTextV2(entry, stateData)
		if err != nil {
			return nil, deployerrors.NewDeployPlanRenderFailed(err, fmt.Sprintf("unable to render teardown.retain[%d]", idx))
		}

		out = strings.TrimSpace(out)
		if out == "" {
			continue
		}
		if err := teardown.ValidateAddress(out); err != nil {
			return nil, deployerrors.NewDeployPlanRenderFailed(err, fmt.Sprintf("invalid teardown.retain[%d]", idx))
		}
		if _, ok := seen[out]; ok {
			continue
		}
		seen[out] = struct{}{}
		rendered = append(rendered, out)
	}

	if len(rendered) == 0 {
		return nil, nil
	}

	return &plantypes.TeardownPlan{Retain: rendered}, nil
}
