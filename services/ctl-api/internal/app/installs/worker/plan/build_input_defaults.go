package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func (p *Planner) applyBuildInputDefaults(ctx workflow.Context, install *app.Install, build *app.ComponentBuild, st *state.State) error {
	buildCfgID := build.ComponentConfigConnection.AppConfigID
	if buildCfgID == "" || buildCfgID == install.AppConfigID {
		return nil
	}

	buildCfg, err := activities.AwaitGetAppConfigByID(ctx, buildCfgID)
	if err != nil {
		return errors.Wrap(err, "unable to get build app config")
	}

	helpers.FillInputDefaults(st, buildCfg)
	return nil
}
