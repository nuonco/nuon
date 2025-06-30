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

type CreateConnectionCommitRequest struct {
	SHA         string `json:"sha" validate:"required"`
	AuthorName  string `json:"author_name" validate:"required"`
	AuthorEmail string `json:"author_email" validate:"required"`
	Message     string `json:"message" validate:"required"`
}

type CreateConnectionRepoRequest struct {
	Name   string                      `json:"name" validate:"required"`
	Status app.VCSConnectionRepoStatus `json:"status" validate:"required"`
}

type CreateConnectionBranchRequest struct {
	Name string                        `json:"name" validate:"required"`
	Repo CreateConnectionRepoRequest   `json:"repo" validate:"required"`
	Head CreateConnectionCommitRequest `json:"head" validate:"required"`
}

func (c *CreateConnectionBranchRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						CreateVCSConnectionBranch
// @Summary				create a vcs connection branch for Github
// @Description.markdown	create_vcs_connection_branch.md
// @Param					connection_id						path	string	true	"connection ID"
// @Param					req	body	CreateConnectionBranchRequest	true	"Input"
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
// @Router					/v1/vcs/connections/{connection_id}/branches [post]
func (s *service) CreateConnectionBranch(ctx *gin.Context) {
	currentOrg, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	connectionID := ctx.Param("connection_id")

	var req CreateConnectionBranchRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	vcsConn, err := s.createOrgConnectionBranch(ctx, currentOrg.ID, connectionID, req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, vcsConn)
}

func (s *service) createOrgConnectionBranch(ctx context.Context, orgID, connectionID string, req CreateConnectionBranchRequest) (*app.VCSConnectionBranch, error) {
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

	branch := app.VCSConnectionBranch{
		OrgID:               orgID,
		VCSConnectionRepoID: repo.ID,
		Name:                req.Name,
		Status:              app.VCSConnectionBranchStatusActive,
	}
	res = s.db.WithContext(ctx).Create(&branch)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create vcs connection branch: %w", res.Error)
	}

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

