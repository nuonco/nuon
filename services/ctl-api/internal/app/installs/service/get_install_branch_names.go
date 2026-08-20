package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetInstallBranchNames
// @Summary				get distinct branch names across all installs for an org
// @Description			Returns all distinct app branch names assigned to installs in the current org.
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	string
// @Router					/v1/installs/branch-names [GET]
func (s *service) GetInstallBranchNames(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var names []string
	query := s.db.WithContext(ctx).
		Raw(`SELECT DISTINCT app_branches.name
FROM installs
JOIN app_branches ON app_branches.id = installs.app_branch_id AND app_branches.deleted_at = 0
WHERE installs.org_id = ? AND installs.deleted_at = 0
ORDER BY app_branches.name`, org.ID)

	if err := query.Scan(&names).Error; err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, names)
}
