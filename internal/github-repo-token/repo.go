package github

import (
	"fmt"
	"strings"
)

func parseRepo(repo string) (string, string, error) {
	if strings.HasPrefix(repo, "https://") || strings.HasPrefix(repo, "git@github") {
		return "", "", fmt.Errorf("invalid github repo: %s", repo)
	}

	if strings.Count(repo, "/") != 1 {
		return "", "", fmt.Errorf("invalid github repo: %s", repo)
	}

	pieces := strings.SplitN(repo, "/", 2)
	return pieces[0], pieces[1], nil
}
