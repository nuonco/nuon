package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// @ID						AdminGetOrgRunner
// @Summary				get an org runner
// @Description.markdown	admin_get_org_runner.md
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					org_id	path	string	true	"org ID for your current org"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-get-runner [GET]
func (s *service) AdminGetOrgRunner(ctx *gin.Context) {
	nameOrID := ctx.Param("org_id")

	org, err := s.adminGetOrg(ctx, nameOrID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if len(org.RunnerGroup.Runners) == 0 {
		ctx.Error(stderr.ErrNotFound{
			Err:         fmt.Errorf("org has no runner"),
			Description: "no runner was found for this org",
		})
		return
	}

	ctx.JSON(http.StatusOK, org.RunnerGroup.Runners[0])
}
