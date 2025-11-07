package plan

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/config"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (p *Planner) createHelmBuildPlan(ctx workflow.Context, bld *app.ComponentBuild, helmCompCfg *app.HelmComponentConfig) (*plantypes.HelmBuildPlan, error) {

	var helmCfg *config.HelmRepoConfig
	if helmCompCfg.HelmConfig.HelmRepoConfig != nil {
		helmCfg = &config.HelmRepoConfig{
			Chart:   helmCompCfg.HelmConfig.HelmRepoConfig.Chart,
			RepoURL: helmCompCfg.HelmConfig.HelmRepoConfig.RepoURL,
			Version: helmCompCfg.HelmConfig.HelmRepoConfig.Version,
		}
	}

	return &plantypes.HelmBuildPlan{
		Labels: map[string]string{
			"component_id":       bld.ComponentID,
			"component_build_id": bld.ID,
		},
		HelmRepoConfig: helmCfg,
	}, nil
}
