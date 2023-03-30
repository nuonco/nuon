package presets

import (
	"time"

	buildv1 "github.com/powertoolsdev/mono/pkg/types/components/build/v1"
	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/components/deploy/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (p *preset) publicHelmChart() (*componentv1.Component, error) {
	return &componentv1.Component{
		Id: p.ID,
		BuildCfg: &buildv1.Config{
			Cfg: &buildv1.Config_Noop{},
		},
		DeployCfg: &deployv1.Config{
			Timeout: durationpb.New(time.Minute * 5),
			Cfg: &deployv1.Config_HelmRepo{
				HelmRepo: &deployv1.HelmRepoConfig{
					ChartName:    "public-helm-chart",
					ChartRepo:    "matheusfm/httpbin",
					ChartVersion: "v0.0.0",

					// TODO(jm): add ability to specify a custom repo
					//$ helm repo add matheusfm https://matheusfm.dev/charts
				},
			},
		},
	}, nil
}
