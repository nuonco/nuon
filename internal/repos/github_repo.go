package repos

import (
	"context"
	"net/http"

	gh "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v41/github"

	"github.com/powertoolsdev/api/internal/models"
	"go.uber.org/zap"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_github_repo.go -source=github_repo.go -package=repos
type GithubRepo interface {
	Repos(context.Context, int64) ([]*models.Repo, error)
	GetCommit(context.Context, int64, string, string, string) (*github.RepositoryCommit, error)
	GetRepo(context.Context, int64, string, string) (*github.Repository, error)
}

var _ GithubRepo = (*githubRepo)(nil)

type githubRepo struct {
	Transport *gh.AppsTransport
	logger    *zap.Logger
	client    *http.Client
}

func NewGithubRepo(transport *gh.AppsTransport, logger *zap.Logger, client *http.Client) *githubRepo {
	return &githubRepo{
		Transport: transport,
		logger:    logger,
		client:    client,
	}
}

func (gr *githubRepo) Repos(ctx context.Context, githubInstallationID int64) ([]*models.Repo, error) {
	installtp := gh.NewFromAppsTransport(gr.Transport, githubInstallationID)

	gr.client.Transport = installtp

	client := github.NewClient(gr.client)

	repos, _, err := client.Apps.ListRepos(ctx, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	r := make([]*models.Repo, 0)

	for _, repo := range repos.Repositories {
		r = append(r, &models.Repo{
			DefaultBranch: repo.DefaultBranch,
			FullName:      repo.FullName,
			Name:          repo.Name,
			Owner:         repo.Owner.Login,
			Private:       repo.Private,
			URL:           repo.HTMLURL,
		})
	}

	return r, nil
}

func (gr *githubRepo) GetCommit(ctx context.Context, githubInstallationID int64, ghRepoOwner, ghRepo, ghBranch string) (*github.RepositoryCommit, error) {
	installtp := gh.NewFromAppsTransport(gr.Transport, githubInstallationID)
	gr.client.Transport = installtp
	client := github.NewClient(gr.client)

	commit, _, err := client.Repositories.GetCommit(ctx, ghRepoOwner, ghRepo, ghBranch, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	return commit, err
}

func (gr *githubRepo) GetRepo(ctx context.Context, githubInstallationID int64, ghRepoOwner string, ghRepoName string) (*github.Repository, error) {
	installtp := gh.NewFromAppsTransport(gr.Transport, githubInstallationID)
	gr.client.Transport = installtp
	client := github.NewClient(gr.client)

	repo, _, err := client.Repositories.Get(ctx, ghRepoOwner, ghRepoName)
	if err != nil {
		return nil, err
	}

	return repo, err
}
