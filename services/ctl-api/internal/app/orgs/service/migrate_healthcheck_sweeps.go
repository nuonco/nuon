package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	healthchecksweepsmigration "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/healthcheck_sweeps_migration"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type MigrateHealthcheckSweepsRequest struct {
	Enabled bool `json:"enabled"`
}

// @ID						MigrateHealthcheckSweeps
// @Summary				migrate org between per-entity and per-org healthcheck emitters
// @Description			Toggle the org-healthcheck-sweeps feature and migrate the org's healthcheck emitters: enabled creates the two per-org sweep emitters and removes the per-runner/per-process cron emitters; disabled recreates the per-entity emitters and removes the sweeps.
// @Param					org_id	path	string							true	"org ID"
// @Param					payload	body	MigrateHealthcheckSweepsRequest	true	"enabled"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/migrate-healthcheck-sweeps [POST]
func (s *service) MigrateHealthcheckSweeps(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req MigrateHealthcheckSweepsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.features.Enable(ctx, org.ID, map[string]bool{
		string(app.OrgFeatureOrgHealthcheckSweeps): req.Enabled,
	}); err != nil {
		s.l.Error("unable to set org-healthcheck-sweeps feature", zap.String("org_id", org.ID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to set feature flag: " + err.Error()})
		return
	}

	if err := s.helpers.EnsureOrgQueue(ctx, org.ID); err != nil {
		s.l.Error("unable to ensure org queue", zap.String("org_id", org.ID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to ensure org queue: " + err.Error()})
		return
	}

	var queue app.Queue
	if res := s.db.WithContext(ctx).
		Where(app.Queue{OwnerID: org.ID, Name: orgshelpers.OrgSignalsQueueName}).
		First(&queue); res.Error != nil {
		s.l.Error("unable to find org queue", zap.String("org_id", org.ID), zap.Error(res.Error))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to find org queue"})
		return
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &healthchecksweepsmigration.Signal{OrgID: org.ID, Enabled: req.Enabled},
	}); err != nil {
		s.l.Error("unable to enqueue healthcheck sweeps migration signal", zap.String("org_id", org.ID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to enqueue migration signal: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, true)
}
