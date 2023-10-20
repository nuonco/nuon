package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminForgetOrgInstallsRequest struct{}

//	@BasePath	/v1/orgs
//
// Forget all installs for an org. This should only be used in _dire_ cases where an org has been lost, and we need to
// deprovision it but do not know / care about the end installs. The primary use case for this is to be able to run
// aws-nuke on the canary account when things go wrong, but still allow us to cleanup the apps/orgs after.
//
//	@Summary	forget all installs for an org
//	@Schemes
//	@Description	forget all installs for an org
//	@Param			org_id	path	string							true	"org ID"
//	@Param			req		body	AdminForgetOrgInstallsRequest	true	"Input"
//	@Tags			installs/admin
//	@Accept			json
//	@Produce		json
//	@Failure		400	{object}	stderr.ErrResponse
//	@Failure		404	{object}	stderr.ErrResponse
//	@Failure		500	{object}	stderr.ErrResponse
//	@Success		200	{boolean}	true
//	@Router			/v1/orgs/{org_id}/admin-forget-installs [POST]
func (s *service) ForgetOrgInstalls(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	installs, err := s.getOrgInstalls(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org installs: %w", err))
		return
	}

	for _, install := range installs {
		err := s.forgetInstall(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		s.hooks.Forgotten(ctx, install.ID)
	}

	ctx.JSON(http.StatusOK, true)
}
