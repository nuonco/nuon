package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	componentsactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/components/worker/activities"
)

func (p *Planner) createTerraformBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.TerraformBuildPlan, error) {
	plan := &plantypes.TerraformBuildPlan{
		Labels: map[string]string{
			"component_id":       bld.ComponentID,
			"component_build_id": bld.ID,
		},
	}

	mirrorEnabled, err := componentsactivities.AwaitOrgHasFeature(ctx, componentsactivities.OrgHasFeatureRequest{
		OrgID:   bld.OrgID,
		Feature: string(app.OrgFeatureTerraformProviderMirror),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to check terraform provider mirror feature flag")
	}
	if mirrorEnabled {
		plan.VendorProviders = true

		if cfg := bld.ComponentConfigConnection.TerraformModuleComponentConfig; cfg != nil {
			plan.TerraformVersion = cfg.Version
		}
	}

	return plan, nil
}
