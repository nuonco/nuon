package dev

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// checkUnpushedChanges checks for any changes in the repository that haven't been pushed,
// regardless of the current working directory.
// Assumes that the caller has already verified we're in a git repository.
// Returns:
// - changedFiles: slice of all changed files (with absolute paths). Empty if no changes.
// - error: if the git commands fail
func checkUnpushedChanges() (changedFiles []string, err error) {
	// Get the git repository root directory
	repoRoot, err := getGitRepoRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get git repository root: %w", err)
	}

	// Initialize results list
	changedFiles = []string{}

	// Save current directory to return to it later
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	defer os.Chdir(currentDir) // Ensure we return to the original directory

	// Change to the repository root for consistent git operations
	if err := os.Chdir(repoRoot); err != nil {
		return nil, fmt.Errorf("failed to change to repository root: %w", err)
	}

	// Check for unpushed commits
	hasUnpushedCommits, unpushedFiles, err := getUnpushedCommitFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to check for unpushed commits: %w", err)
	}

	// Add unpushed commit files to results
	if hasUnpushedCommits {
		// Convert to absolute paths and add to results
		for _, f := range unpushedFiles {
			absPath := filepath.Join(repoRoot, f)
			if !contains(changedFiles, absPath) {
				changedFiles = append(changedFiles, absPath)
			}
		}
	}

	// Check for staged changes
	stagedFiles, err := getStagedChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to check for staged changes: %w", err)
	}

	if len(stagedFiles) > 0 {
		// Convert to absolute paths and add to results
		for _, f := range stagedFiles {
			absPath := filepath.Join(repoRoot, f)
			if !contains(changedFiles, absPath) {
				changedFiles = append(changedFiles, absPath)
			}
		}
	}

	// Check for unstaged changes
	unstagedFiles, err := getUnstagedChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to check for unstaged changes: %w", err)
	}

	if len(unstagedFiles) > 0 {
		// Convert to absolute paths and add to results
		for _, f := range unstagedFiles {
			absPath := filepath.Join(repoRoot, f)
			if !contains(changedFiles, absPath) {
				changedFiles = append(changedFiles, absPath)
			}
		}
	}

	return changedFiles, nil
}

// contains checks if a string slice contains a specific string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

// getGitRepoRoot returns the absolute path to the root of the git repository
func getGitRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getUnpushedCommitFiles gets files changed in unpushed commits
func getUnpushedCommitFiles() (bool, []string, error) {
	// Get the current branch name
	branch, err := getCurrentBranch()
	if err != nil {
		return false, nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Check if the branch has a remote tracking branch
	hasRemote, err := hasRemoteTrackingBranch(branch)
	if err != nil {
		return false, nil, fmt.Errorf("failed to check for remote tracking branch: %w", err)
	}

	if !hasRemote {
		// If there's no tracking branch, we consider all commits as "unpushed"
		files, err := getAllFiles()
		return true, files, err
	}

	// Compare local branch with its remote tracking branch
	cmd := exec.Command("git", "rev-list", "--count", "--left-right",
		fmt.Sprintf("%s...@{upstream}", branch))

	output, err := cmd.Output()
	if err != nil {
		return false, nil, fmt.Errorf("failed to compare with remote branch: %w", err)
	}

	// Output format is "X Y" where:
	// X = number of commits ahead (unpushed)
	// Y = number of commits behind (unpulled)
	counts := strings.Fields(strings.TrimSpace(string(output)))
	if len(counts) != 2 {
		return false, nil, fmt.Errorf("unexpected output format from git rev-list: %s", output)
	}

	// Check if we have unpushed commits
	if counts[0] == "0" {
		// No unpushed commits
		return false, nil, nil
	}

	// We have unpushed commits, so get the list of changed files
	files, err := getFilesInUnpushedCommits(branch)
	return true, files, err
}

// getFilesInUnpushedCommits gets the list of files that have been changed in unpushed commits
func getFilesInUnpushedCommits(branch string) ([]string, error) {
	// Get list of files changed between local branch and remote tracking branch
	cmd := exec.Command("git", "diff", "--name-only", fmt.Sprintf("%s..@{upstream}", branch))
	output, err := cmd.Output()
	if err != nil {
		// Try another approach if the previous command failed
		// This might happen if the branch exists remotely but has diverged completely
		cmd = exec.Command("git", "diff", "--name-only", "origin/main...")
		output, err = cmd.Output()
		if err != nil {
			return nil, err
		}
	}

	// Split output into file list
	if len(strings.TrimSpace(string(output))) == 0 {
		return nil, nil
	}
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

// getAllFiles gets the list of all files in the repository
func getAllFiles() ([]string, error) {
	// Get list of all files in the repository
	cmd := exec.Command("git", "ls-files")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Split output into file list
	if len(strings.TrimSpace(string(output))) == 0 {
		return nil, nil
	}
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

// getStagedChanges returns a list of files with staged changes
func getStagedChanges() ([]string, error) {
	// Get list of staged files
	cmd := exec.Command("git", "diff", "--name-only", "--cached")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Split output into file list
	if len(strings.TrimSpace(string(output))) == 0 {
		return nil, nil
	}
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

// getUnstagedChanges returns a list of files with unstaged changes
func getUnstagedChanges() ([]string, error) {
	// Get list of unstaged modified, added, and deleted files
	cmd := exec.Command("git", "diff", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Split output into file list
	if len(strings.TrimSpace(string(output))) == 0 {
		return nil, nil
	}
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

// getCurrentBranch gets the name of the current git branch
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// hasRemoteTrackingBranch checks if the given branch has a remote tracking branch
func hasRemoteTrackingBranch(branch string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref",
		fmt.Sprintf("%s@{upstream}", branch))

	// If the command succeeds, the branch has a remote tracking branch
	err := cmd.Run()
	if err != nil {
		// This specific error means no upstream branch, not a command failure
		return false, nil
	}
	return true, nil
}
