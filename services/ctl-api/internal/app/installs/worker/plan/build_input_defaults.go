package plan

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

const buildInputDefaultsVersion = "deploy-plan-build-input-defaults-v1"

func (p *Planner) applyBuildInputDefaults(ctx workflow.Context, install *app.Install, build *app.ComponentBuild, st *state.State) error {
	buildCfgID := build.ComponentConfigConnection.AppConfigID
	if buildCfgID == "" || buildCfgID == install.AppConfigID {
		return nil
	}
	if workflow.GetVersion(ctx, buildInputDefaultsVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return nil
	}

	buildCfg, err := activities.AwaitGetAppConfigByID(ctx, buildCfgID)
	if err != nil {
		return errors.Wrap(err, "unable to get build app config")
	}

	FillInputDefaults(st, buildCfg)
	return nil
}

func FillInputDefaults(st *state.State, cfg *app.AppConfig) {
	if st == nil || cfg == nil {
		return
	}

	for _, inp := range cfg.InputConfig.AppInputs {
		if inp.Default == "" {
			continue
		}
		if st.Inputs == nil {
			st.Inputs = state.NewInputsState()
			st.Inputs.Populated = true
		}
		if st.Inputs.Inputs[inp.Name] != "" {
			continue
		}

		st.Inputs.Inputs[inp.Name] = inp.Default
		if st.Install != nil {
			if st.Install.Inputs == nil {
				st.Install.Inputs = make(map[string]string)
			}
			st.Install.Inputs[inp.Name] = inp.Default
		}
	}
}
