package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgphonehomebackfill "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/phone_home_backfill"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type BackfillPhoneHomeRequest struct{}

// @ID						AdminBackfillOrgPhoneHome
// @Summary				backfill phone home secrets for an org
// @Description			Step one of onboarding an org to phone-home auth; step two is enabling the flag via PATCH /v1/orgs/{org_id}/admin-features. Enqueues a signal that, for every install in the org, backfills the cloud platform metadata from the identifier the install's stack already reported, then provisions the phone-home secret: minting a token per live stack version, publishing the phone_home_id-to-token map to Secrets Manager in the Nuon AWS account, and granting the install's phone-home role cross-account read. The metadata step comes first because an install created before target_account_id existed has no target, and the secret step skips any install without one. Runs whether or not the org has phone-home-auth enabled, so provisioning always precedes the flag rather than racing it. Idempotent — the reconciler is convergent and a target identifier supplied by a user is never overwritten. Note this is pre-provisioning only: an already-deployed phone-home Lambda does not know the secret's location and will not send a token until its stack version is regenerated and re-applied.
// @Param					org_id	path	string							true	"org ID"
// @Param					req		body	BackfillPhoneHomeRequest	true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	map[string]string
// @Router					/v1/orgs/{org_id}/admin-backfill-phone-home [POST]
func (s *service) AdminBackfillOrgPhoneHome(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req BackfillPhoneHomeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	queueID, err := s.getOrgSignalsQueueID(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org signals queue: %w", err))
		return
	}

	resp, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queueID,
		Signal:    &orgphonehomebackfill.Signal{OrgID: org.ID},
		OwnerID:   org.ID,
		OwnerType: plugins.TableName(s.db, app.Org{}),
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue phone home secret backfill signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"queue_signal_id": resp.ID,
		"queue_id":        queueID,
	})
}
