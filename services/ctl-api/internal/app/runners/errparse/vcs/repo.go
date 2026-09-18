package vcs

import "strings"

func normalizeRepo(raw string) string {
	repo := strings.TrimSpace(raw)
	repo = strings.Trim(repo, `"'`)
	repo = strings.TrimRight(repo, ".,;:")
	if repo == "" {
		return ""
	}

	if idx := strings.Index(repo, "://"); idx >= 0 {
		repo = repo[idx+3:]
	}
	if idx := strings.LastIndex(repo, "@"); idx >= 0 {
		repo = repo[idx+1:]
	}
	repo = strings.ReplaceAll(repo, ":", "/")
	repo = strings.TrimSuffix(strings.TrimSuffix(repo, "/"), ".git")

	segments := make([]string, 0, 3)
	for _, segment := range strings.Split(repo, "/") {
		if segment != "" {
			segments = append(segments, segment)
		}
	}
	if len(segments) < 2 {
		return ""
	}
	return strings.Join(segments[len(segments)-2:], "/")
}
