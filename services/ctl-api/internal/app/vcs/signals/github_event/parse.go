package githubevent

import (
	"fmt"
	"strings"
)

type pushEventInfo struct {
	Repo   string // "owner/repo" - matches ConnectedGithubVCSConfig.Repo
	Branch string // "main" - matches ConnectedGithubVCSConfig.Branch
}

type pullRequestEventInfo struct {
	Repo       string // "owner/repo"
	BaseBranch string // target branch (e.g., "main")
	HeadSHA    string // head commit SHA
	PRNumber   int    // pull request number
	Action     string // "opened", "synchronize", "closed", etc.
}

func parsePushEvent(payload map[string]any) (*pushEventInfo, error) {
	// Extract ref (e.g. "refs/heads/main")
	ref, ok := payload["ref"].(string)
	if !ok || ref == "" {
		return nil, fmt.Errorf("missing or invalid ref in push payload")
	}

	branch := strings.TrimPrefix(ref, "refs/heads/")
	if branch == ref {
		// ref didn't have the expected prefix (e.g. tag push)
		return nil, fmt.Errorf("ref %q is not a branch push", ref)
	}

	// Extract repository.full_name (e.g. "owner/repo")
	repository, ok := payload["repository"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing or invalid repository in push payload")
	}

	fullName, ok := repository["full_name"].(string)
	if !ok || fullName == "" {
		return nil, fmt.Errorf("missing or invalid repository.full_name in push payload")
	}

	return &pushEventInfo{
		Repo:   fullName,
		Branch: branch,
	}, nil
}

func parsePullRequestEvent(payload map[string]any) (*pullRequestEventInfo, error) {
	action, _ := payload["action"].(string)
	if action == "" {
		return nil, fmt.Errorf("missing action in pull_request payload")
	}

	prData, ok := payload["pull_request"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing pull_request in payload")
	}

	number, ok := prData["number"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing pull_request.number")
	}

	base, ok := prData["base"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing pull_request.base")
	}
	baseBranch, _ := base["ref"].(string)
	if baseBranch == "" {
		return nil, fmt.Errorf("missing pull_request.base.ref")
	}

	head, ok := prData["head"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing pull_request.head")
	}
	headSHA, _ := head["sha"].(string)

	repository, ok := payload["repository"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing repository in pull_request payload")
	}
	fullName, _ := repository["full_name"].(string)
	if fullName == "" {
		return nil, fmt.Errorf("missing repository.full_name")
	}

	return &pullRequestEventInfo{
		Repo:       fullName,
		BaseBranch: baseBranch,
		HeadSHA:    headSHA,
		PRNumber:   int(number),
		Action:     action,
	}, nil
}
