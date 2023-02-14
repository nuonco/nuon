package repos

import (
	"context"
	"errors"
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
	GetInstallToken(context.Context, int64) (string, error)
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

// Get an install token from the github API
func (gr *githubRepo) GetInstallToken(ctx context.Context, githubInstallationID int64) (string, error) {
	// create github client
	gr.client.Transport = gr.Transport
	client := github.NewClient(gr.client)

	// get a new install token
	token, resp, _ := client.Apps.CreateInstallationToken(
		ctx,
		githubInstallationID,
		&github.InstallationTokenOptions{})
	// The go-github-mock library has a bug that causes the test to panic
	// when the Error() method is called, se we cannot have the
	// if err != nil check here if we want to have tests.
	// The library always returns a response object so we are throwing
	// the status returned by github as an error if the token is nil
	// bug reference: https://github.com/migueleliasweb/go-github-mock/issues/6
	if token != nil {
		return token.GetToken(), nil
	}
	return "", errors.New(resp.Response.Status)
}
