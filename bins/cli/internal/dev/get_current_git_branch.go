package dev

import (
	"fmt"
	"os/exec"
	"strings"
)

// getCurrentGitBranch returns the name of the current git branch in the specified directory.
// If not in a git repository or an error occurs, it returns an error.
func getCurrentGitBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting git branch: %w", err)
	}

	// Trim whitespace from the output
	branch := strings.TrimSpace(string(output))
	return branch, nil
}
