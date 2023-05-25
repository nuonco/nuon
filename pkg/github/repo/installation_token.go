package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/pkg/github/client"
	"github.com/powertoolsdev/mono/pkg/kube/secret"
)

func (g *gh) ClonePath(ctx context.Context) (string, error) {
	token, err := g.InstallationToken(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get installation token: %w", err)
	}

	return fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", g.RepoOwner, token, g.RepoOwner, g.RepoName), nil
}

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=installation_token_mock_test.go -source=installation_token.go -package=github
func (g *gh) InstallationToken(ctx context.Context) (string, error) {
	sg, err := secret.New(g.v,
		secret.WithNamespace(g.AppKeySecretNamespace),
		secret.WithName(g.AppKeySecretName),
		secret.WithKey(appKeySecretKeyKey),
		secret.WithCluster(g.AppKeyClusterInfo),
	)
	if err != nil {
		return "", fmt.Errorf("unable to get secret getter: %w", err)
	}

	appKey, err := sg.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get app key: %w", err)
	}

	ghClient, err := client.New(g.v,
		client.WithAppID(g.AppKeyID),
		client.WithAppKey(appKey),
	)
	if err != nil {
		return "", fmt.Errorf("unable to get github client: %w", err)
	}

	token, err := g.createInstallationToken(ctx, ghClient.Apps)
	if err != nil {
		return "", fmt.Errorf("unable to get github client: %w", err)
	}

	return token, nil
}

type installationTokenCreatorClient interface {
	CreateInstallationToken(ctx context.Context,
		id int64,
		opts *github.InstallationTokenOptions,
	) (*github.InstallationToken, *github.Response, error)
}

func (g *gh) createInstallationToken(ctx context.Context, ghClient installationTokenCreatorClient) (string, error) {
	resp, _, err := ghClient.CreateInstallationToken(ctx, g.InstallID, &github.InstallationTokenOptions{
		Repositories: []string{g.RepoName},
	})
	if err != nil {
		return "", fmt.Errorf("error creating installation token: %w", err)
	}

	if len(resp.Repositories) != 1 || *resp.Repositories[0].Name != g.RepoName {
		return "", fmt.Errorf("installation does not allow accessing repo: %s", g.RepoName)
	}

	return *resp.Token, nil
}
