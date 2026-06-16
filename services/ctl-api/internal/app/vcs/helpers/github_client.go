package helpers

//go:generate go run github.com/golang/mock/mockgen -source=github_client.go -destination=mock_github_client.go -package=helpers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v50/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// GithubClient defines the GitHub API operations used by the VCS service.
// The *Helpers struct implements this interface. Tests can provide a fake.
type GithubClient interface {
	GetInstallationAccount(ctx context.Context, installID string) (*github.User, error)
	GetInstallation(ctx context.Context, installID string) (*github.Installation, error)
	DeleteInstallation(ctx context.Context, installID string) error
	ListInstallationRepos(ctx context.Context, vcsConn *app.VCSConnection) ([]*github.Repository, error)
	CreateOrgWebhook(ctx context.Context, vcsConn *app.VCSConnection, webhookURL string, secret string) (int64, error)
}

func (h *Helpers) GetInstallationAccount(ctx context.Context, installID string) (*github.User, error) {
	installation, err := h.GetInstallation(ctx, installID)
	if err != nil {
		return nil, err
	}
	if installation.Account == nil {
		return nil, fmt.Errorf("github installation account is nil")
	}
	return installation.Account, nil
}

func (h *Helpers) GetInstallation(ctx context.Context, installID string) (*github.Installation, error) {
	ghClient, err := h.GetJWTVCSConnectionClient()
	if err != nil {
		return nil, fmt.Errorf("unable to create jwt vcs connection client: %w", err)
	}

	iInstallID, err := strconv.ParseInt(installID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to convert github install ID to int: %w", err)
	}

	installation, _, err := ghClient.Apps.GetInstallation(ctx, iInstallID)
	if err != nil {
		return nil, fmt.Errorf("unable to get github installation: %w", err)
	}

	return installation, nil
}

func (h *Helpers) DeleteInstallation(ctx context.Context, installID string) error {
	ghClient, err := h.GetJWTVCSConnectionClient()
	if err != nil {
		return fmt.Errorf("unable to create jwt vcs connection client: %w", err)
	}

	iInstallID, err := strconv.ParseInt(installID, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to convert github install ID to int: %w", err)
	}

	_, err = ghClient.Apps.DeleteInstallation(ctx, iInstallID)
	if err != nil {
		return fmt.Errorf("unable to delete github app installation id: %d: %w", iInstallID, err)
	}

	return nil
}

func (h *Helpers) CreateOrgWebhook(ctx context.Context, vcsConn *app.VCSConnection, webhookURL string, secret string) (int64, error) {
	ghClient, err := h.GetVCSConnectionClient(ctx, vcsConn)
	if err != nil {
		return 0, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	active := true
	hook := &github.Hook{
		Config: map[string]interface{}{
			"url":          webhookURL,
			"content_type": "json",
			"secret":       secret,
		},
		Events: []string{"push", "pull_request", "create", "delete"},
		Active: &active,
	}

	created, _, err := ghClient.Organizations.CreateHook(ctx, vcsConn.GithubAccountName, hook)
	if err != nil {
		return 0, fmt.Errorf("failed to create org webhook: %w", err)
	}

	return created.GetID(), nil
}

func (h *Helpers) ListInstallationRepos(ctx context.Context, vcsConn *app.VCSConnection) ([]*github.Repository, error) {
	ghClient, err := h.GetVCSConnectionClient(ctx, vcsConn)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	var allRepos []*github.Repository
	opts := &github.ListOptions{PerPage: 100}

	for {
		repos, resp, err := ghClient.Apps.ListRepos(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list GitHub repositories: %w", err)
		}

		allRepos = append(allRepos, repos.Repositories...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}
