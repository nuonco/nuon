package plantypes

import "github.com/powertoolsdev/mono/pkg/config"

type HelmBuildPlan struct {
	Labels         map[string]string
	HelmRepoConfig *config.HelmRepoConfig
}
