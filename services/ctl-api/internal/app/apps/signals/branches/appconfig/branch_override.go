package appconfig

import "github.com/nuonco/nuon/pkg/config"

// overrideComponentBranch checks if a component's connected repo matches the
// branch config's repo. If so, it overrides the component's branch to match
// the deploying branch. This ensures that when a component lives in the same
// repo as the app config, it uses the correct branch.
func overrideComponentBranch(comp *config.Component, branchRepo, branchName string) {
	if comp.TerraformModule != nil && comp.TerraformModule.ConnectedRepo != nil {
		if comp.TerraformModule.ConnectedRepo.Repo == branchRepo {
			comp.TerraformModule.ConnectedRepo.Branch = branchName
		}
	}
	if comp.HelmChart != nil && comp.HelmChart.ConnectedRepo != nil {
		if comp.HelmChart.ConnectedRepo.Repo == branchRepo {
			comp.HelmChart.ConnectedRepo.Branch = branchName
		}
	}
	if comp.DockerBuild != nil && comp.DockerBuild.ConnectedRepo != nil {
		if comp.DockerBuild.ConnectedRepo.Repo == branchRepo {
			comp.DockerBuild.ConnectedRepo.Branch = branchName
		}
	}
	if comp.KubernetesManifest != nil && comp.KubernetesManifest.ConnectedRepo != nil {
		if comp.KubernetesManifest.ConnectedRepo.Repo == branchRepo {
			comp.KubernetesManifest.ConnectedRepo.Branch = branchName
		}
	}
	if comp.Pulumi != nil && comp.Pulumi.ConnectedRepo != nil {
		if comp.Pulumi.ConnectedRepo.Repo == branchRepo {
			comp.Pulumi.ConnectedRepo.Branch = branchName
		}
	}
}
