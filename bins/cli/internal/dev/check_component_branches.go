package dev

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/config"
)

type componentRepo struct {
	ComponentName string
	RepoName      string
	RepoBranch    string
	RepoDir       string
}

func (s *Service) checkComponentBranches(ctx context.Context, cfg *config.AppConfig, currentBranch string) ([]componentRepo, error) {
	cmpRepos := []componentRepo{}

	msg := "The following components are not on the same branch:\n\n"
	branchDiff := false

	for _, cmp := range cfg.Components {
		var publicRepo *config.PublicRepoConfig
		var connectedRepo *config.ConnectedRepoConfig

		switch cmp.Type {
		case config.TerraformModuleComponentType:
			publicRepo = cmp.TerraformModule.PublicRepo
			connectedRepo = cmp.TerraformModule.ConnectedRepo
		case config.HelmChartComponentType:
			publicRepo = cmp.HelmChart.PublicRepo
			connectedRepo = cmp.HelmChart.ConnectedRepo
		case config.DockerBuildComponentType:
			publicRepo = cmp.DockerBuild.PublicRepo
			connectedRepo = cmp.DockerBuild.ConnectedRepo
			// these don't have repos
		case config.ContainerImageComponentType:
			fallthrough
		case config.ExternalImageComponentType:
			continue
		}

		repoName := ""
		branchName := ""
		dirName := ""
		switch {
		case publicRepo != nil:
			repoName = publicRepo.Repo
			branchName = publicRepo.Branch
			dirName = publicRepo.Directory
		case connectedRepo != nil:
			repoName = connectedRepo.Repo
			branchName = connectedRepo.Branch
			dirName = connectedRepo.Directory
		}

		cmpRepos = append(cmpRepos, componentRepo{
			ComponentName: cmp.Name,
			RepoName:      repoName,
			RepoBranch:    branchName,
			RepoDir:       dirName,
		})

		if branchName != currentBranch {
			branchDiff = true
			msg += fmt.Sprintf("- %s: %s\n", cmp.Name, branchName)
		}
	}

	// TODO add the ability to override the branch
	if branchDiff {
		msg += "\nDo you want to continue"
		if err := prompt(s.autoApprove, msg); err != nil {
			return cmpRepos, err
		}
	}

	return cmpRepos, nil
}
