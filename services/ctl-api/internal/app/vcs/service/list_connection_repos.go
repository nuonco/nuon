package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v50/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type VCSConnectionRepo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Description   string `json:"description,omitempty"`
	Private       bool   `json:"private"`
	Fork          bool   `json:"fork"`
	HTMLURL       string `json:"html_url"`
	DefaultBranch string `json:"default_branch"`
	UpdatedAt     string `json:"updated_at"`
}

type VCSConnectionReposResponse struct {
	Repositories []VCSConnectionRepo `json:"repositories"`
	TotalCount   int                 `json:"total_count"`
}

// ListConnectionRepos godoc
//
// @Summary      List VCS connection repositories
// @Description  Lists all repositories accessible by a GitHub App installation (VCS connection)
// @Tags         vcs
// @Produce      json
// @Param        connection_id  path      string  true  "VCS Connection ID"
// @Success      200            {object}  VCSConnectionReposResponse
// @Failure      400            {object}  stderr.ErrResponse
// @Failure      401            {object}  stderr.ErrResponse
// @Failure      403            {object}  stderr.ErrResponse
// @Failure      404            {object}  stderr.ErrResponse
// @Failure      500            {object}  stderr.ErrResponse
// @Security     APIKey
// @Security     OrgID
// @Router       /v1/vcs/connections/{connection_id}/repos [get]
func (s *service) ListConnectionRepos(ctx *gin.Context) {
	connectionID := ctx.Param("connection_id")

	currentOrg, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	vcsConn, err := s.getConnection(ctx, currentOrg.ID, connectionID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org vcs connection: %w", err))
		return
	}

	repos, err := s.ghClient.ListInstallationRepos(ctx, vcsConn)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list repositories: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, buildReposResponse(repos))
}

func buildReposResponse(ghRepos []*github.Repository) *VCSConnectionReposResponse {
	repos := make([]VCSConnectionRepo, len(ghRepos))
	for i, r := range ghRepos {
		repos[i] = VCSConnectionRepo{
			ID:            r.GetID(),
			Name:          r.GetName(),
			FullName:      r.GetFullName(),
			Description:   r.GetDescription(),
			Private:       r.GetPrivate(),
			Fork:          r.GetFork(),
			HTMLURL:       r.GetHTMLURL(),
			DefaultBranch: r.GetDefaultBranch(),
			UpdatedAt:     r.GetUpdatedAt().Time.Format(time.RFC3339),
		}
	}

	return &VCSConnectionReposResponse{
		Repositories: repos,
		TotalCount:   len(repos),
	}
}
