package githubevent

import (
	"fmt"
	"strings"
)

type pushEventInfo struct {
	Repo         string   // "owner/repo" - matches ConnectedGithubVCSConfig.Repo
	Branch       string   // "main" - matches ConnectedGithubVCSConfig.Branch
	PusherEmail  string   // email of the person who pushed
	SenderLogin  string   // GitHub username of the sender (always present)
	PusherEmails []string // all unique emails from the payload (pusher, commit author/committer)
	ChangedFiles []string // deduplicated list of added/modified/removed files across all commits
	HeadSHA      string   // SHA of the head commit
	BeforeSHA    string   // SHA before the push, used to fetch the complete comparison
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

	var pusherEmail string
	if pusher, ok := payload["pusher"].(map[string]any); ok {
		pusherEmail, _ = pusher["email"].(string)
	}

	var senderLogin string
	if sender, ok := payload["sender"].(map[string]any); ok {
		senderLogin, _ = sender["login"].(string)
	}

	emails := collectPushEmails(payload, pusherEmail)
	changedFiles := collectChangedFiles(payload)

	var headSHA string
	if hc, ok := payload["head_commit"].(map[string]any); ok {
		headSHA, _ = hc["id"].(string)
	}
	beforeSHA, _ := payload["before"].(string)

	return &pushEventInfo{
		Repo:         fullName,
		Branch:       branch,
		PusherEmail:  pusherEmail,
		SenderLogin:  senderLogin,
		PusherEmails: emails,
		ChangedFiles: changedFiles,
		HeadSHA:      headSHA,
		BeforeSHA:    beforeSHA,
	}, nil
}

func collectPushEmails(payload map[string]any, pusherEmail string) []string {
	seen := make(map[string]bool)
	var emails []string

	addEmail := func(email string) {
		e := strings.TrimSpace(strings.ToLower(email))
		if e == "" || seen[e] || strings.HasSuffix(e, "@users.noreply.github.com") {
			return
		}
		seen[e] = true
		emails = append(emails, email)
	}

	addEmail(pusherEmail)

	if hc, ok := payload["head_commit"].(map[string]any); ok {
		if author, ok := hc["author"].(map[string]any); ok {
			if e, ok := author["email"].(string); ok {
				addEmail(e)
			}
		}
		if committer, ok := hc["committer"].(map[string]any); ok {
			if e, ok := committer["email"].(string); ok {
				addEmail(e)
			}
		}
	}

	if commits, ok := payload["commits"].([]any); ok {
		for _, c := range commits {
			commit, ok := c.(map[string]any)
			if !ok {
				continue
			}
			if author, ok := commit["author"].(map[string]any); ok {
				if e, ok := author["email"].(string); ok {
					addEmail(e)
				}
			}
			if committer, ok := commit["committer"].(map[string]any); ok {
				if e, ok := committer["email"].(string); ok {
					addEmail(e)
				}
			}
		}
	}

	return emails
}

func collectChangedFiles(payload map[string]any) []string {
	seen := make(map[string]bool)
	var files []string

	addFiles := func(commit map[string]any, key string) {
		if arr, ok := commit[key].([]any); ok {
			for _, f := range arr {
				if s, ok := f.(string); ok && !seen[s] {
					seen[s] = true
					files = append(files, s)
				}
			}
		}
	}

	if commits, ok := payload["commits"].([]any); ok {
		for _, c := range commits {
			commit, ok := c.(map[string]any)
			if !ok {
				continue
			}
			addFiles(commit, "added")
			addFiles(commit, "modified")
			addFiles(commit, "removed")
		}
	}

	return files
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
