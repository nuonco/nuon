package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	orgdeprovision "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/deprovision"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminDeprovisionOrgRequest struct {
	Force bool `json:"force"`
}

// @ID						AdminDeprovisionOrg
// @Summary				deprovision an org, but keep it in the database
// @Description.markdown	deprovision_org.md
// @Param					org_id	path	string	true	"org ID for your current org"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminDeprovisionOrgRequest	true	"Input"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-deprovision [POST]
func (s *service) AdminDeprovisionOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")
	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req AdminDeprovisionOrgRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	// App deletion is handled by the deprovision signal automatically.
	// The signal will fail if any apps still have installs that need to be forgotten first.

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
		if err := s.enqueueOrgSignal(ctx, queueID, &orgdeprovision.Signal{OrgID: org.ID, Force: req.Force}); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		sigTyp := sigs.OperationDeprovision
		if req.Force {
			sigTyp = sigs.OperationForceDeprovision
		}
		s.evClient.Send(ctx, org.ID, &sigs.Signal{
			Type: sigTyp,
		})
	}

	ctx.JSON(http.StatusOK, true)
}
