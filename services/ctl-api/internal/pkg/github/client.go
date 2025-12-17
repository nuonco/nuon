package github

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"github.com/nuonco/nuon/pkg/github/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

func New(v *validator.Validate, cfg *internal.Config) (*github.Client, error) {
	ghClient, err := client.New(v,
		client.WithAppID(cfg.GithubAppID),
		client.WithAppKey([]byte(cfg.GithubAppKey)),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get github client: %w", err)
	}

	return ghClient, nil
}
