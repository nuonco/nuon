package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

//	@ID			DeleteOrg
//	@Summary	Delete an org
//	@Schemes
//	@Description.markdown	delete_org.md
//	@Tags					orgs
//	@Accept					json
//	@Security				APIKey
//	@Security				OrgID
//	@Produce				json
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				200	{boolean}	ok
//	@Router					/v1/orgs/current [DELETE]
func (s *service) DeleteOrg(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if org.OrgType == app.OrgTypeIntegration {
		err := s.helpers.HardDelete(ctx, org.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusOK, true)
		return
	}

	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type:        sigs.OperationDelete,
		ForceDelete: false,
	})

	ctx.JSON(http.StatusOK, true)
}
