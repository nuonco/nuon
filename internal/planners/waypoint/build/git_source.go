package build

import (
	"context"
	"fmt"

	vcsv1 "github.com/powertoolsdev/protos/components/generated/types/vcs/v1"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	githubtoken "github.com/powertoolsdev/workers-executors/internal/github-repo-token"
)

const (
	//nolint:gosec
	githubAppKeySecretName      string = "graphql-api-github-app-key"
	githubAppKeySecretNamespace string = "default"
)

//nolint:unparam
func (p *planner) getPublicGitSource(_ context.Context, cfg *vcsv1.PublicGithubConfig) (*planv1.GitSource, error) {
	return &planv1.GitSource{
		Url:  cfg.Repo,
		Ref:  cfg.GitRef,
		Path: cfg.Directory,
	}, nil
}

func (p *planner) getPrivateGitSource(ctx context.Context, cfg *vcsv1.PrivateGithubConfig) (*planv1.GitSource, error) {
	tokenGetter, err := githubtoken.New(p.V,
		githubtoken.WithRepo(cfg.Repo),
		githubtoken.WithInstallID(cfg.GithubInstallId),
		githubtoken.WithAppKeyID(cfg.GithubAppKeyId),
		githubtoken.WithAppKeySecretName(githubAppKeySecretName),
		githubtoken.WithAppKeySecretNamespace(githubAppKeySecretNamespace),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get github token: %w", err)
	}

	clonePath, err := tokenGetter.ClonePath(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get clone path: %w", err)
	}

	return &planv1.GitSource{
		Url:               clonePath,
		Ref:               cfg.GitRef,
		Path:              cfg.Directory,
		RecurseSubmodules: 2,
	}, nil
}
