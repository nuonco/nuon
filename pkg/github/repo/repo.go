package github

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/nuonco/nuon/pkg/generics"
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

// CloneURLAuth separates transport credentials from the URL stored by go-git.
func CloneURLAuth(raw string) (string, transport.AuthMethod, error) {
	if !strings.Contains(raw, "://") {
		return raw, nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", nil, fmt.Errorf("invalid clone URL")
	}
	if u.User == nil {
		return raw, nil, nil
	}
	user := u.User.Username()
	password, hasPassword := u.User.Password()
	switch u.Scheme {
	case "http", "https":
		u.User = nil
		if user == "" {
			user = "git"
		}
		return u.String(), &http.BasicAuth{Username: user, Password: password}, nil
	case "file":
		u.User = nil
		return u.String(), nil, nil
	default:
		if hasPassword {
			return "", nil, fmt.Errorf("password clone authentication requires HTTP(S)")
		}
		return raw, nil, nil
	}
}
