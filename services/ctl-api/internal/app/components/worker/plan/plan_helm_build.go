package plan

import (
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildsource"
)

func (p *Planner) createHelmBuildPlan(_ workflow.Context, bld *app.ComponentBuild, helmCompCfg *app.HelmComponentConfig) (*plantypes.HelmBuildPlan, error) {
	connection := bld.ComponentConfigConnection
	connection.HelmComponentConfig = helmCompCfg
	inputs, err := buildsource.ComponentInputs(&connection)
	if err != nil {
		return nil, err
	}
	plan := inputs.HelmBuildPlan
	plan.Labels = map[string]string{
		"component_id":       bld.ComponentID,
		"component_build_id": bld.ID,
	}

	return plan, nil
}
