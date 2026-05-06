package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	orgreprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/reprovision"
)

type ReprovisionOrgRequest struct{}

// @ID						AdminReprovisionOrg
// @Summary				reprovision an org
// @Description.markdown	reprovision_org.md
// @Param					org_id	path	string	true	"org ID for your current org"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	ReprovisionOrgRequest	true	"Input"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-reprovision [POST]
func (s *service) AdminReprovisionOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	useQueues, err := s.useOrgQueues(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}
	if useQueues {
		queueID, err := s.getOrgSignalsQueueID(ctx, org.ID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get org signals queue: %w", err))
			return
		}
		if err := s.enqueueOrgSignal(ctx, queueID, &orgreprovision.Signal{OrgID: org.ID}, org.ID); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, org.ID, &sigs.Signal{
			Type: sigs.OperationReprovision,
		})
	}
	ctx.JSON(http.StatusOK, true)
}
