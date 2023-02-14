package converters

import (
	"github.com/powertoolsdev/api/internal/models"
	githubv1 "github.com/powertoolsdev/protos/api/generated/types/github/v1"
)

// GithubRepoModelToProto converts github repo domain model into github repo proto message
func GithubRepoModelToProto(repo *models.Repo) *githubv1.Repo {
	return &githubv1.Repo{
		DefaultBranch: *repo.DefaultBranch,
		FullName:      *repo.FullName,
		Name:          *repo.Name,
		Owner:         *repo.Owner,
		Private:       *repo.Private,
		Url:           *repo.URL,
	}
}

// GithubRepoModelsToProto converts a slice of github repo models to protos
func GithubRepoModelsToProto(repos []*models.Repo) []*githubv1.Repo {
	protos := make([]*githubv1.Repo, len(repos))
	for idx, repo := range repos {
		protos[idx] = GithubRepoModelToProto(repo)
	}

	return protos
}
