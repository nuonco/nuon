package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
)

//	@ID						AdminGetRunnerServiceAccount
//	@Summary				get a runner service account
//	@Description.markdown	get_runner_service_account.md
//	@Tags					runners/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					runner_id	path	string	true	"runner ID to fetch"
//	@Produce				json
//	@Success				200	{object}	app.RunnerGroup
//	@Router					/v1/runners/{runner_id}/service-account [GET]
func (s *service) AdminGetRunnerServiceAccount(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	email := account.ServiceAccountEmail(runnerID)

	svcAcct, err := s.acctClient.FindAccount(ctx, email)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to find account"))
		return
	}

	ctx.JSON(http.StatusOK, svcAcct)
}
