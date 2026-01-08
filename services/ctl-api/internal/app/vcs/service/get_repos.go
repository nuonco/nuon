package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type Repo struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	Private  bool   `json:"private"`
}

// @ID						GetVCSConnectionRepos
// @Summary				List repositories accessible by VCS connection
// @Description			Returns list of repositories that the VCS connection has access to
// @Param					connection_id	path	string	true	"connection ID"
// @Tags					vcs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		Repo
// @Router					/v1/vcs/connections/{connection_id}/repos [get]
func (s *service) GetVCSConnectionRepos(ctx *gin.Context) {
	vcsID := ctx.Param("connection_id")

	currentOrg, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	vcsConn, err := s.getConnection(ctx, currentOrg.ID, vcsID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get vcs connection: %w", err))
		return
	}

	repos, err := s.listRepos(ctx, vcsConn)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list repos: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, repos)
}

func (s *service) listRepos(ctx context.Context, vcsConn *app.VCSConnection) ([]Repo, error) {
	client, err := s.helpers.GetVCSConnectionClient(ctx, vcsConn)
	if err != nil {
		return nil, fmt.Errorf("unable to get github client: %w", err)
	}

	opt := &github.ListOptions{
		PerPage: 100,
	}

	var allRepos []Repo
	for {
		repos, resp, err := client.Apps.ListRepos(ctx, opt)
		if err != nil {
			return nil, fmt.Errorf("unable to list repos: %w", err)
		}

		for _, repo := range repos.Repositories {
			if repo.FullName != nil && repo.Name != nil && repo.Owner != nil && repo.Owner.Login != nil {
				allRepos = append(allRepos, Repo{
					FullName: *repo.FullName,
					Name:     *repo.Name,
					Owner:    *repo.Owner.Login,
					Private:  repo.Private != nil && *repo.Private,
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}
