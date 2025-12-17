package plantypes

import "github.com/nuonco/nuon/pkg/config"

type HelmBuildPlan struct {
	Labels         map[string]string
	HelmRepoConfig *config.HelmRepoConfig
}
