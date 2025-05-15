package dev

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

// GetUnstagedFiles returns a list of all filenames in a git repository that have
// unstaged changes. The function will work regardless of what subdirectory of the
// repo the user is currently in.
func getUnstagedFiles() ([]string, error) {
	// Find the git repository root
	repoRoot, err := getGitRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to find git repository: %w", err)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get the status of files in the worktree
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree status: %w", err)
	}

	// Collect files with unstaged changes
	var unstagedFiles []string
	for file, fileStatus := range status {
		// Check if the file has unstaged changes
		if fileStatus.Worktree != git.Unmodified && fileStatus.Staging != fileStatus.Worktree {
			unstagedFiles = append(unstagedFiles, file)
		}
	}

	return unstagedFiles, nil
}
