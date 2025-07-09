package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetVCSBranches
// @Summary				get all vcs branches
// @Description.markdown	get_vcs_branches.md
// @Param					offset						query	int		false	"offset of branches to return"	Default(0)
// @Param					limit						query	int		false	"limit of branches to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Param					x-nuon-pagination-enabled	header	bool	false	"Enable pagination"
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
// @Success				200	{object}	[]app.VCSConnectionBranch
// @Router					/v1/vcs/branches [get]
func (s *service) GetVCSBranches(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	cfgs, err := s.getAllVCSBranches(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, cfgs)
}

func (s *service) getAllVCSBranches(ctx *gin.Context, orgID string) ([]app.VCSConnectionBranch, error) {
	branches := make([]app.VCSConnectionBranch, 0)

	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where(app.VCSConnectionBranch{
			OrgID: orgID,
		}).
		Preload("VCSConnectionRepo").
		Preload("VCSConnectionCommits").
		Preload("VCSConnectionCommits", func(db *gorm.DB) *gorm.DB {
			return db.Order("vcs_connection_commits.created_at DESC")
		}).
		Order("created_at desc").
		Find(&branches)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all vcs branches: %w", res.Error)
	}

	branches, err := db.HandlePaginatedResponse(ctx, branches)
	if err != nil {
		return nil, fmt.Errorf("unable to get all vcs branches: %w", err)
	}

	return branches, nil
}
