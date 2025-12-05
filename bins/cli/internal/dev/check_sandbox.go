package dev

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/config"
)

func (s *Service) checkSandbox(ctx context.Context, cfg *config.AppConfig, currentBranch string) error {
	var publicRepo = cfg.Sandbox.PublicRepo
	var connectedRepo = cfg.Sandbox.ConnectedRepo

	branchName := ""
	switch {
	case publicRepo != nil:
		branchName = publicRepo.Branch
	case connectedRepo != nil:
		branchName = connectedRepo.Branch
	}
	if branchName != currentBranch {
		return fmt.Errorf("The sanbox is configured to use the git branch \"%s\". Please configure it to use your dev branch.", branchName)
	}

	return nil
}
