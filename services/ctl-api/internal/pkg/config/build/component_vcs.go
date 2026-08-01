package build

import "github.com/nuonco/nuon/pkg/config"

// ComponentRepos returns the git sources from whichever type config is set.
func ComponentRepos(comp *config.Component) (*config.ConnectedRepoConfig, *config.PublicRepoConfig) {
	switch {
	case comp.DockerBuild != nil:
		return comp.DockerBuild.ConnectedRepo, comp.DockerBuild.PublicRepo
	case comp.HelmChart != nil:
		return comp.HelmChart.ConnectedRepo, comp.HelmChart.PublicRepo
	case comp.TerraformModule != nil:
		return comp.TerraformModule.ConnectedRepo, comp.TerraformModule.PublicRepo
	case comp.KubernetesManifest != nil:
		return comp.KubernetesManifest.ConnectedRepo, comp.KubernetesManifest.PublicRepo
	case comp.Pulumi != nil:
		return comp.Pulumi.ConnectedRepo, comp.Pulumi.PublicRepo
	default:
		return nil, nil
	}
}

// NeedsTerraformVersion means the caller must resolve the latest release.
func NeedsTerraformVersion(comp *config.Component) bool {
	return comp.TerraformModule != nil && comp.TerraformModule.TerraformVersion == ""
}
