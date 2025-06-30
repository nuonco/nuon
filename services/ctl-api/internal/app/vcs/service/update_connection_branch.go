package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type UpdateConnectionCommitRequest struct {
	SHA         string `json:"sha" validate:"required"`
	AuthorName  string `json:"author_name" validate:"required"`
	AuthorEmail string `json:"author_email" validate:"required"`
	Message     string `json:"message" validate:"required"`
}

type UpdateConnectionRepoRequest struct {
	Name   string                      `json:"name" validate:"required"`
	Status app.VCSConnectionRepoStatus `json:"status" validate:"required"`
}

type UpdateConnectionBranchRequest struct {
	Name string                        `json:"name" validate:"required"`
	Repo UpdateConnectionRepoRequest   `json:"repo" validate:"required"`
	Head UpdateConnectionCommitRequest `json:"head" validate:"required"`
}

func (c *UpdateConnectionBranchRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						UpdateVCSConnectionBranch
// @Summary				update a vcs connection branch for Github
// @Description.markdown	update_connection_branch.md
// @Param					connection_id						path	string	true	"connection ID"
// @Param					connection_branch_id				path	string	true	"connection branch ID"
// @Param					req	body	UpdateConnectionBranchRequest	true	"Input"
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
// @Success				201	{object}	app.VCSConnectionBranch
// @Router					/v1/vcs/connections/{connection_id}/branches/{connection_branch_id} [patch]
func (s *service) UpdateConnectionBranch(ctx *gin.Context) {
	currentOrg, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	connectionID := ctx.Param("connection_id")
	connectionBranchID := ctx.Param("connection_branch_id")

	var req UpdateConnectionBranchRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	vcsConn, err := s.updateOrgConnectionBranch(ctx, currentOrg.ID, connectionID, connectionBranchID, req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, vcsConn)
}

func (s *service) updateOrgConnectionBranch(ctx context.Context, orgID, connectionID, connectionBranchID string, req UpdateConnectionBranchRequest) (*app.VCSConnectionBranch, error) {
	// upsert the repo
	repo := app.VCSConnectionRepo{
		OrgID:           orgID,
		VCSConnectionID: connectionID,
		Name:            req.Repo.Name,
		Status:          req.Repo.Status,
	}
	res := s.db.WithContext(ctx).Save(&repo)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to save vcs connection repo: %w", res.Error)
	}

	// update the branch
	branch := app.VCSConnectionBranch{
		ID:                  connectionBranchID,
		OrgID:               orgID,
		VCSConnectionRepoID: repo.ID,
		Name:                req.Name,
		Status:              app.VCSConnectionBranchStatusActive,
	}
	res = s.db.WithContext(ctx).Updates(&branch)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update vcs connection branch: %w", res.Error)
	}

	// upsert the head commit
	commits := []app.VCSConnectionCommit{
		{
			OrgID:                 orgID,
			VCSConnectionID:       connectionID,
			VCSConnectionRepoID:   generics.NewNullString(repo.ID),
			VCSConnectionBranchID: generics.NewNullString(branch.ID),
			SHA:                   req.Head.SHA,
			AuthorName:            req.Head.AuthorName,
			AuthorEmail:           req.Head.AuthorEmail,
			Message:               req.Head.Message,
		},
	}
	res = s.db.WithContext(ctx).Save(&commits)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to save vcs connection commit: %w", res.Error)
	}

	return &branch, nil
}

