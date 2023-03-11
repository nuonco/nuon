package github

import (
	"fmt"
	"net/http"
	"strconv"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/nuonctl/internal"
)

type githubOption func(*github) error

type github struct {
	AppID  int64  `validate:"required"`
	AppKey []byte `validate:"required"`
}

func New(opts ...githubOption) (*ghinstallation.AppsTransport, error) {
	gh := &github{}

	for idx, opt := range opts {
		if err := opt(gh); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	validate := validator.New()
	if err := validate.Struct(gh); err != nil {
		return nil, fmt.Errorf("unable to validate temporal: %w", err)
	}

	tr := http.DefaultTransport
	appstp, err := ghinstallation.NewAppsTransport(tr, gh.AppID, gh.AppKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create github apps transport: %w", err)
	}

	return appstp, nil
}

func WithConfig(cfg *internal.Config) githubOption {
	return func(d *github) error {
		githubAppID, err := strconv.ParseInt(cfg.GithubAppID, 10, 64)
		if err != nil {
			return fmt.Errorf("unable to parse github app id: %w", err)
		}
		d.AppID = githubAppID
		d.AppKey = []byte(cfg.GithubAppKey)
		return nil
	}
}
