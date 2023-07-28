package gqlclient

import (
	"context"
	"fmt"
)

type ConnectedRepo = getConnectedReposReposRepoConnectionEdgesRepoEdgeNodeRepo

func (c *client) GetConnectedRepo(ctx context.Context, orgID, repoName string) (*ConnectedRepo, error) {
	resp, err := getOrgGithubInstallID(ctx, c.graphqlClient, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org github install ID: %w", err)
	}

	repoResp, err := getConnectedRepos(ctx, c.graphqlClient, resp.Org.GithubInstallId)
	if err != nil {
		return nil, fmt.Errorf("unable to get  connected repos: %w", err)
	}
	for _, repo := range repoResp.Repos.Edges {
		if repo.Node.Name == repoName || repo.Node.Url == repoName || repo.Node.FullName == repoName {
			return repo.Node, nil
		}
	}

	return nil, fmt.Errorf("repo not found: %s", repoName)
}

func (c *client) GetConnectedRepos(ctx context.Context, orgID string) ([]*ConnectedRepo, error) {
	resp, err := getOrgGithubInstallID(ctx, c.graphqlClient, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get org github install ID: %w", err)
	}

	repoResp, err := getConnectedRepos(ctx, c.graphqlClient, resp.Org.GithubInstallId)
	if err != nil {
		return nil, fmt.Errorf("unable to get  connected repos: %w", err)
	}

	repos := make([]*ConnectedRepo, 0)
	for _, repo := range repoResp.Repos.Edges {
		r := repo
		repos = append(repos, r.Node)
	}

	return repos, nil
}
