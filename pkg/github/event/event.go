package event

import (
	"fmt"
	"strings"
)

// Commit holds parsed commit data from a GitHub push event payload.
type Commit struct {
	SHA         string
	Message     string
	AuthorName  string
	AuthorEmail string
}

// PushEvent holds parsed data from a GitHub push event payload.
type PushEvent struct {
	Branch    string
	RepoOwner string
	RepoName  string
	Commits   []Commit
}

// ExtractRepoInfo pulls repository.owner.login and repository.name from any
// GitHub webhook payload. These fields are present on all event types.
func ExtractRepoInfo(payload map[string]any) (owner, name string, err error) {
	repo, ok := payload["repository"]
	if !ok {
		return "", "", fmt.Errorf("payload missing 'repository' field")
	}

	repoMap, ok := repo.(map[string]any)
	if !ok {
		return "", "", fmt.Errorf("'repository' is not an object")
	}

	nameRaw, _ := repoMap["name"].(string)
	if nameRaw == "" {
		return "", "", fmt.Errorf("'repository.name' not found")
	}

	ownerObj, ok := repoMap["owner"].(map[string]any)
	if !ok {
		return "", "", fmt.Errorf("'repository.owner' is not an object")
	}

	ownerLogin, _ := ownerObj["login"].(string)
	if ownerLogin == "" {
		return "", "", fmt.Errorf("'repository.owner.login' not found")
	}

	return ownerLogin, nameRaw, nil
}

// ParsePushEvent extracts branch, repo info, and commits from a GitHub push event payload.
func ParsePushEvent(payload map[string]any) PushEvent {
	var ev PushEvent

	if ref, ok := payload["ref"].(string); ok {
		ev.Branch = strings.TrimPrefix(ref, "refs/heads/")
	}

	if repo, ok := payload["repository"].(map[string]interface{}); ok {
		if owner, ok := repo["owner"].(map[string]interface{}); ok {
			ev.RepoOwner, _ = owner["login"].(string)
		}
		ev.RepoName, _ = repo["name"].(string)
	}

	commitsRaw, ok := payload["commits"]
	if !ok {
		return ev
	}
	commits, ok := commitsRaw.([]interface{})
	if !ok {
		return ev
	}

	for _, c := range commits {
		cm, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		sha, _ := cm["id"].(string)
		if sha == "" {
			continue
		}
		message, _ := cm["message"].(string)
		var authorName, authorEmail string
		if author, ok := cm["author"].(map[string]interface{}); ok {
			authorName, _ = author["name"].(string)
			authorEmail, _ = author["email"].(string)
		}
		ev.Commits = append(ev.Commits, Commit{
			SHA:         sha,
			Message:     message,
			AuthorName:  authorName,
			AuthorEmail: authorEmail,
		})
	}

	return ev
}
