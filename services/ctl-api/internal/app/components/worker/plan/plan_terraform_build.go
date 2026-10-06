package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (p *Planner) createTerraformBuildPlan(_ workflow.Context, bld *app.ComponentBuild) (*plantypes.TerraformBuildPlan, error) {
	plan := &plantypes.TerraformBuildPlan{
		Labels: map[string]string{
			"component_id":       bld.ComponentID,
			"component_build_id": bld.ID,
		},
		VendorProviders: true,
	}

	// Plumb the configured terraform version through to the build
	// runner so it can install the matching CLI to vendor providers
	// via `terraform providers mirror`. Empty values cause the build
	// runner to fall back to its default version.
	if cfg := bld.ComponentConfigConnection.TerraformModuleComponentConfig; cfg != nil {
		plan.TerraformVersion = cfg.Version
	}

	return plan, nil
}
