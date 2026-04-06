package githubevent

import (
	"fmt"
	"strings"
)

type pushEventInfo struct {
	Repo   string // "owner/repo" - matches ConnectedGithubVCSConfig.Repo
	Branch string // "main" - matches ConnectedGithubVCSConfig.Branch
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
