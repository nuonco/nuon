package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	gh "github.com/bradleyfalzon/ghinstallation/v2"

	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/utils"
	"go.uber.org/zap"
)

type githubService struct {
	repoGetter RepoGetter
}

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_github_service.go -source=github_service.go -package=services
type GithubService interface {
	Repos(context.Context, string) ([]*models.Repo, *utils.Page, error)
}

type RepoGetter interface {
	Repos(context.Context, int64) ([]*models.Repo, error)
}

func NewGithubService(tsprt *gh.AppsTransport, l *zap.Logger) *githubService {
	githubRepo := repos.NewGithubRepo(tsprt, l, &http.Client{})
	return &githubService{
		repoGetter: githubRepo,
	}
}

func (ghs *githubService) Repos(ctx context.Context, githubInstallationID string) ([]*models.Repo, *utils.Page, error) {
	ghInstallID, err := strconv.ParseInt(githubInstallationID, 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to convert gh install ID to int: %w", err)
	}
	repos, err := ghs.repoGetter.Repos(ctx, ghInstallID)
	if err != nil {
		return nil, nil, err
	}

	return repos, &utils.Page{}, nil
}
