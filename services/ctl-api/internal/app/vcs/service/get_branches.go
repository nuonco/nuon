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

type Branch struct {
	Name string `json:"name"`
}

// @ID						GetVCSConnectionRepoBranches
// @Summary				List branches for a repository
// @Description			Returns list of branches for the specified repository
// @Param					connection_id	path	string	true	"connection ID"
// @Param					owner			query	string	true	"repository owner"
// @Param					repo			query	string	true	"repository name"
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
// @Success				200	{array}		Branch
// @Router					/v1/vcs/connections/{connection_id}/branches [get]
func (s *service) GetVCSConnectionRepoBranches(ctx *gin.Context) {
	vcsID := ctx.Param("connection_id")
	owner := ctx.Query("owner")
	repo := ctx.Query("repo")

	if owner == "" || repo == "" {
		ctx.Error(fmt.Errorf("owner and repo query parameters are required"))
		return
	}

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

	branches, err := s.listBranches(ctx, vcsConn, owner, repo)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list branches: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, branches)
}

func (s *service) listBranches(ctx context.Context, vcsConn *app.VCSConnection, owner, repo string) ([]Branch, error) {
	client, err := s.helpers.GetVCSConnectionClient(ctx, vcsConn)
	if err != nil {
		return nil, fmt.Errorf("unable to get github client: %w", err)
	}

	opt := &github.BranchListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allBranches []Branch
	for {
		branches, resp, err := client.Repositories.ListBranches(ctx, owner, repo, opt)
		if err != nil {
			return nil, fmt.Errorf("unable to list branches: %w", err)
		}

		for _, branch := range branches {
			if branch.Name != nil {
				allBranches = append(allBranches, Branch{
					Name: *branch.Name,
				})
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allBranches, nil
}
