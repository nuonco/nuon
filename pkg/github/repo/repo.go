package github

import (
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/pkg/generics"
)

const (
	defaultGitPrefix         string = "git@"
	defaultHttpsPrefix       string = "https://"
	defaultGithubURLTemplate string = "https://github.com/%s/%s"
)

func EnsureURL(url string) (string, error) {
	if generics.HasAnyPrefix(url, defaultHttpsPrefix, defaultGitPrefix) {
		return url, nil
	}

	owner, name, err := ParseRepo(url)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(defaultGithubURLTemplate, owner, name), nil
}

func ParseRepo(repo string) (string, string, error) {
	if strings.HasPrefix(repo, "https://") || strings.HasPrefix(repo, "git@github") {
		return "", "", fmt.Errorf("invalid github repo: %s", repo)
	}

	if strings.Count(repo, "/") != 1 {
		return "", "", fmt.Errorf("invalid github repo: %s", repo)
	}

	pieces := strings.SplitN(repo, "/", 2)
	return pieces[0], pieces[1], nil
}

func RepoPath(owner, repo, token string) string {
	return fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", owner, token, owner, repo)
}
