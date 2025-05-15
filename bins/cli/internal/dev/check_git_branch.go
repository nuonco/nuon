package dev

import (
	"context"
	"fmt"
)

func (s *Service) checkGitBranch(ctx context.Context) (string, error) {
	if !isGitRepository() {
		return "", fmt.Errorf("Error getting the current git branch: You are not in a git repository")
	}

	gitBranch, err := getCurrentGitBranch()
	if err != nil {
		return gitBranch, fmt.Errorf("Error getting the current git branch: %s", err)
	}

	switch gitBranch {
	case "":
		return gitBranch, fmt.Errorf("Error getting the current git branch: You are not on a branch, please check out a dev branch")
	case "main":
		fallthrough
	case "master":
		fallthrough
	case "integration":
		if err := prompt(s.autoApprove, "You are currently on the branch \"%s\". Are you sure you want to continue?", gitBranch); err != nil {
			return gitBranch, err
		}
	}
	return gitBranch, nil
}
