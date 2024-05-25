package protos

import (
	"fmt"
	"time"

	buildv1 "github.com/powertoolsdev/mono/pkg/types/components/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
	variablesv1 "github.com/powertoolsdev/mono/pkg/types/components/variables/v1"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	defaultHelmDeployTimeout time.Duration = time.Minute * 15
)

func (c *Adapter) toHelmValues(inputVals map[string]*string) *variablesv1.HelmValues {
	vals := make([]*variablesv1.HelmValue, 0)
	for k, v := range inputVals {
		if v == nil {
			continue
		}

		vals = append(vals, &variablesv1.HelmValue{
			Name:      k,
			Value:     *v,
			Sensitive: true,
		})
	}

	return &variablesv1.HelmValues{
		Values: vals,
	}
}

func (c *Adapter) ToHelmComponentConfig(cfg *app.HelmComponentConfig, connections []app.InstallDeploy, gitRef string) (*componentv1.Component, error) {
	vcsCfg, err := c.ToVCSConfig(gitRef, cfg.PublicGitVCSConfig, cfg.ConnectedGithubVCSConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get vcs config: %w", err)
	}

	return &componentv1.Component{
		Id: cfg.ComponentConfigConnection.ComponentID,
		BuildCfg: &buildv1.Config{
			Timeout: durationpb.New(defaultBuildTimeout),
			Cfg: &buildv1.Config_HelmChartCfg{
				HelmChartCfg: &buildv1.HelmChartConfig{
					VcsCfg:    vcsCfg,
					ChartName: cfg.ChartName,
				},
			},
		},
		DeployCfg: &deployv1.Config{
			Timeout: durationpb.New(defaultHelmDeployTimeout),
			Cfg: &deployv1.Config_HelmChart{
				HelmChart: &deployv1.HelmChartConfig{
					Values:      c.toHelmValues(cfg.Values),
					ValuesFiles: cfg.ValuesFiles,
				},
			},
		},
		Connections: c.toConnections(connections),
	}, nil
}
