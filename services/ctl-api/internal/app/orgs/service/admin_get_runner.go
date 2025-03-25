package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//	@ID						AdminGetOrgRunner
//	@Summary				get an org runner
//	@Description.markdown	admin_get_org_runner.md
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					org_id	path	string	true	"org ID for your current org"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/orgs/{org_id}/admin-get-runner [GET]
func (s *service) AdminGetOrgRunner(ctx *gin.Context) {
	nameOrID := ctx.Param("org_id")

	org, err := s.adminGetOrg(ctx, nameOrID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org.RunnerGroup.Runners[0])
}
