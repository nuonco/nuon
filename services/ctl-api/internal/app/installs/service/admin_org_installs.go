package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminGetOrgRequest struct {
	Name string `validate:"required"`
}

// @ID						AdminGetOrgInstalls
// @Summary				get installs in an org
// @Description.markdown	admin_get_org_installs.md
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					org_id	path	string	true	"install ID"
// @Produce				json
// @Success				200	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-get-installs [GET]
func (s *service) AdminGetOrgInstalls(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	installs, err := s.getOrgInstalls(ctx, orgID, "")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org installs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installs)
}
