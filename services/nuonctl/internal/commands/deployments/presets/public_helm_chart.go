package presets

import (
	buildv1 "github.com/powertoolsdev/mono/pkg/types/components/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
)

func (p *preset) publicHelmChart() (*componentv1.Component, error) {
	return &componentv1.Component{
		Id: p.ID,
		BuildCfg: &buildv1.Config{
			Cfg: &buildv1.Config_Noop{},
		},
		DeployCfg: &deployv1.Config{
			Cfg: &deployv1.Config_ExternalHelm{
				ExternalHelm: &deployv1.ExternalHelmConfig{
					Name:     "public-helm-chart",
					ChartUrl: "matheusfm/httpbin",
					// TODO(jm): add ability to specify a custom repo
					//$ helm repo add matheusfm https://matheusfm.dev/charts
				},
			},
		},
	}, nil
}
