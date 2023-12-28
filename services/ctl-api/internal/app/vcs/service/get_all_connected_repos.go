package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v50/github"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"golang.org/x/oauth2"
)

const (
	defaultReposPerPage int = 100
)

type Repository struct {
	Name            string `json:"name,omitempty" validate:"required"`
	FullName        string `json:"full_name,omitempty" validate:"required"`
	UserName        string `json:"user_name" validate:"required"`
	GitURL          string `json:"git_url,omitempty" validate:"required"`
	DefaultBranch   string `json:"default_branch,omitempty" validate:"required"`
	CloneURL        string `json:"clone_url,omitempty" validate:"required"`
	GithubInstallID string `json:"github_install_id,omitempty" validate:"required"`
}

// @ID GetAllVCSConnectedRepos
// @Summary	get all vcs connected repos for an org
// @Description.markdown get_all_vcs_connected_repos.md
// @Tags			vcs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		Repository
// @Router			/v1/vcs/connected-repos [get]
func (s *service) GetAllConnectedRepos(ctx *gin.Context) {
	currentOrg, err := org.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	vcsConns, err := s.getOrgConnections(ctx, currentOrg.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org vcs connections: %w", err))
		return
	}

	repos := make([]*Repository, 0)
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

func (s *service) getConnectionRepos(ctx context.Context, conn *app.VCSConnection) ([]*Repository, error) {
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
	allRepos := make([]*Repository, 0)
	page := 1
	for {
		repos, resp, err := client.Apps.ListRepos(ctx, &github.ListOptions{
			Page:    page,
			PerPage: defaultReposPerPage,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to get repos: %w", err)
		}

		for _, repo := range repos.Repositories {
			allRepos = append(allRepos, &Repository{
				Name:            generics.FromPtrStr(repo.Name),
				FullName:        generics.FromPtrStr(repo.FullName),
				UserName:        generics.FromPtrStr(repo.Owner.Login),
				GitURL:          generics.FromPtrStr(repo.GitURL),
				CloneURL:        generics.FromPtrStr(repo.CloneURL),
				DefaultBranch:   generics.FromPtrStr(repo.DefaultBranch),
				GithubInstallID: conn.GithubInstallID,
			})
		}
		if resp.NextPage < 1 {
			break
		}
		page += 1
	}

	return allRepos, nil
}
