package github

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/servers/converters"
	githubv1 "github.com/powertoolsdev/protos/api/generated/types/github/v1"
)

func (s *server) GetRepos(
	ctx context.Context,
	req *connect.Request[githubv1.GetReposRequest],
) (*connect.Response[githubv1.GetReposResponse], error) {
	repos, _, err := s.Svc.Repos(ctx, req.Msg.GithubInstallId)
	if err != nil {
		return nil, fmt.Errorf("unable to list repos: %w", err)
	}

	return connect.NewResponse(&githubv1.GetReposResponse{
		Repos: converters.GithubRepoModelsToProto(repos),
	}), nil
}
