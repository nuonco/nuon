package github

import (
	"fmt"
	"net/http"
	"strconv"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-playground/validator/v10"
)

type githubOption func(*github) error

type github struct {
	v *validator.Validate

	AppID  int64  `validate:"required"`
	AppKey []byte `validate:"required"`
}

func New(v *validator.Validate, opts ...githubOption) (*ghinstallation.AppsTransport, error) {
	gh := &github{
		v: v,
	}

	for idx, opt := range opts {
		if err := opt(gh); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := v.Struct(gh); err != nil {
		return nil, fmt.Errorf("unable to validate temporal: %w", err)
	}

	tr := http.DefaultTransport
	appstp, err := ghinstallation.NewAppsTransport(tr, gh.AppID, gh.AppKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create github apps transport: %w", err)
	}

	return appstp, nil
}

func WithGithubAppID(ghAppID string) githubOption {
	return func(g *github) error {
		githubAppID, err := strconv.ParseInt(ghAppID, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse github app id: %w", err)
		}
		g.AppID = githubAppID
		return nil
	}
}

func WithGithubAppKey(ghAppKey string) githubOption {
	return func(g *github) error {
		g.AppKey = []byte(ghAppKey)
		return nil
	}
}
