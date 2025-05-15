package dev

import "os/exec"

// isGitRepository checks if the current directory is within a git repository.
// Returns true if it is, false otherwise.
func isGitRepository() bool {
	// Run "git rev-parse --is-inside-work-tree" which returns true if in a git repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

	// We don't care about the output, just whether the command succeeds
	err := cmd.Run()

	// If the command executed successfully (exit code 0), we're in a git repo
	return err == nil
}
