package services

import (
	"context"
	"net/http"

	gh "github.com/bradleyfalzon/ghinstallation/v2"

	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/utils"
	"go.uber.org/zap"
)

type GithubService struct {
	repoGetter RepoGetter
}

type RepoGetter interface {
	Repos(context.Context, int64) ([]*models.Repo, error)
}

func NewGithubService(tsprt *gh.AppsTransport, l *zap.Logger) *GithubService {
	githubRepo := repos.NewGithubRepo(tsprt, l, &http.Client{})
	return &GithubService{
		repoGetter: githubRepo,
	}
}

func (ghs *GithubService) Repos(ctx context.Context, githubInstallationID int64, options *models.ConnectionOptions) ([]*models.Repo, *utils.Page, error) {
	repos, err := ghs.repoGetter.Repos(ctx, githubInstallationID)
	if err != nil {
		return nil, nil, err
	}

	return repos, &utils.Page{}, nil
}
