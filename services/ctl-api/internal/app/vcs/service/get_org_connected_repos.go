package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"golang.org/x/oauth2"
)

const (
	defaultReposPerPage int = 100
)

func (s *service) GetOrgConnectedRepos(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	vcsConns, err := s.getOrgConnections(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org vcs connections: %w", err))
		return
	}

	repos := make([]*github.Repository, 0)
	for _, vcsConn := range vcsConns {
		vcsConnRepos, err := s.getConnectionRepos(ctx, vcsConn)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get repos: %w", err))
			return
		}
		repos = append(repos, vcsConnRepos...)
	}

	ctx.JSON(http.StatusOK, repos)
}

func (s *service) getConnectionRepos(ctx context.Context, conn *app.VCSConnection) ([]*github.Repository, error) {
	// get a static token
	installID, err := strconv.ParseInt(conn.GithubInstallID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to get install ID: %w", err)
	}
	resp, _, err := s.ghClient.Apps.CreateInstallationToken(ctx, installID, &github.InstallationTokenOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get installation token: %w", err)
	}

	// get a client with the github install token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *resp.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// fetch all repos
	allRepos := make([]*github.Repository, 0)
	page := 1
	for {
		repos, resp, err := client.Apps.ListRepos(ctx, &github.ListOptions{
			Page:    page,
			PerPage: defaultReposPerPage,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to get repos: %w", err)
		}

		allRepos = append(allRepos, repos.Repositories...)
		if resp.NextPage < 1 {
			break
		}
		page += 1
	}

	return allRepos, nil
}
