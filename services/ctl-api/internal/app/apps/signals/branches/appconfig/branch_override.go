package appconfig

import (
	"strings"

	"github.com/nuonco/nuon/pkg/config"
)

func overrideBranches(cfg *config.AppConfig, branchRepo, branchName string) {
	for _, comp := range cfg.Components {
		overrideComponentBranch(comp, branchRepo, branchName)
	}

	if cfg.Sandbox != nil {
		overrideRepoConfigs(cfg.Sandbox.ConnectedRepo, cfg.Sandbox.PublicRepo, branchRepo, branchName)
	}

	for _, action := range cfg.Actions {
		for _, step := range action.Steps {
			overrideRepoConfigs(step.ConnectedRepo, step.PublicRepo, branchRepo, branchName)
		}
	}
}

func overrideComponentBranch(comp *config.Component, branchRepo, branchName string) {
	if comp.TerraformModule != nil {
		overrideRepoConfigs(comp.TerraformModule.ConnectedRepo, comp.TerraformModule.PublicRepo, branchRepo, branchName)
	}
	if comp.HelmChart != nil {
		overrideRepoConfigs(comp.HelmChart.ConnectedRepo, comp.HelmChart.PublicRepo, branchRepo, branchName)
	}
	if comp.DockerBuild != nil {
		overrideRepoConfigs(comp.DockerBuild.ConnectedRepo, comp.DockerBuild.PublicRepo, branchRepo, branchName)
	}
	if comp.KubernetesManifest != nil {
		overrideRepoConfigs(comp.KubernetesManifest.ConnectedRepo, comp.KubernetesManifest.PublicRepo, branchRepo, branchName)
	}
	if comp.Pulumi != nil {
		overrideRepoConfigs(comp.Pulumi.ConnectedRepo, comp.Pulumi.PublicRepo, branchRepo, branchName)
	}
}

func overrideRepoConfigs(connected *config.ConnectedRepoConfig, public *config.PublicRepoConfig, branchRepo, branchName string) {
	if connected != nil && repoURLsMatch(connected.Repo, branchRepo) {
		connected.Branch = branchName
	}
	if public != nil && repoURLsMatch(public.Repo, branchRepo) {
		public.Branch = branchName
	}
}

// repoURLsMatch normalizes and compares two repo identifiers.
// Handles both formats:
//   - ConnectedGithubVCSConfig: "owner/repo"
//   - PublicGitVCSConfig: "https://github.com/owner/repo.git"
func repoURLsMatch(a, b string) bool {
	return normalizeRepo(a) == normalizeRepo(b)
}

func normalizeRepo(repo string) string {
	repo = strings.TrimSuffix(repo, ".git")
	repo = strings.TrimPrefix(repo, "https://github.com/")
	repo = strings.TrimPrefix(repo, "http://github.com/")
	repo = strings.TrimPrefix(repo, "git@github.com:")
	repo = strings.TrimSuffix(repo, "/")
	return strings.ToLower(repo)
}
