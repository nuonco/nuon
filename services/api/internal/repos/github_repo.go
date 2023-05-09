package repos

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	gh "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v41/github"

	"github.com/powertoolsdev/mono/services/api/internal/models"
	"go.uber.org/zap"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_github_repo.go -source=github_repo.go -package=repos
type GithubRepo interface {
	Repos(context.Context, string) ([]*models.Repo, error)
	GetCommit(context.Context, string, string, string, string) (*github.RepositoryCommit, error)
	GetRepo(context.Context, string, string, string) (*github.Repository, error)
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

func (gr *githubRepo) Repos(ctx context.Context, githubInstallID string) ([]*models.Repo, error) {
	giid, err := parseGithubInstallID(githubInstallID)
	if err != nil {
		return nil, fmt.Errorf("invalid github install ID: %w", err)
	}
	installtp := gh.NewFromAppsTransport(gr.Transport, giid)

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

func (gr *githubRepo) GetCommit(ctx context.Context, githubInstallID string, ghRepoOwner, ghRepo, ghBranch string) (*github.RepositoryCommit, error) {
	client := github.NewClient(gr.client)
	if githubInstallID != "" {
		giid, err := parseGithubInstallID(githubInstallID)
		if err != nil {
			return nil, fmt.Errorf("invalid github install ID: %w", err)
		}
		installtp := gh.NewFromAppsTransport(gr.Transport, giid)
		gr.client.Transport = installtp
		client = github.NewClient(gr.client)
	}

	commit, _, err := client.Repositories.GetCommit(ctx, ghRepoOwner, ghRepo, ghBranch, &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	return commit, err
}

func (gr *githubRepo) GetRepo(ctx context.Context, githubInstallID string, ghRepoOwner string, ghRepoName string) (*github.Repository, error) {
	// public repo using our generic github application credentials to call the API
	client := github.NewClient(gr.client)

	if githubInstallID != "" {
		// private repo using our installed github app permission to access it
		giid, parsingErr := parseGithubInstallID(githubInstallID)
		if parsingErr != nil {
			gr.logger.Error("failed to parse GithubInstallID",
				zap.String("GithubInstallID", githubInstallID),
				zap.String("error", parsingErr.Error()))
			return nil, fmt.Errorf("error parsing GithubInstallID during GetRepo: %s. %w", githubInstallID, parsingErr)
		}

		installtp := gh.NewFromAppsTransport(gr.Transport, giid)
		gr.client.Transport = installtp
	}

	repo, _, err := client.Repositories.Get(ctx, ghRepoOwner, ghRepoName)
	return repo, err
}

func parseGithubInstallID(githubInstallID string) (int64, error) {
	giid, parsingErr := strconv.ParseInt(githubInstallID, 10, 64)
	if parsingErr != nil {
		return 0, fmt.Errorf("invalid GithubInstallID: %s. %w", githubInstallID, parsingErr)
	}
	return giid, nil
}
