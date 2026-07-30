package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	queuemigration "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/queue_migration"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type MigrateCronsNamespacesRequest struct {
	Enabled bool `json:"enabled"`
}

// @ID						MigrateCronsNamespaces
// @Summary				migrate org cron queues between namespaces
// @Description			Toggle the cron-namespace-isolation feature for the org and migrate its runner-healthcheck and install cron queues into (enabled) or out of (disabled) the dedicated cron Temporal namespaces.
// @Param					org_id	path	string							true	"org ID"
// @Param					payload	body	MigrateCronsNamespacesRequest	true	"enabled"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/migrate-crons-namespaces [POST]
func (s *service) MigrateCronsNamespaces(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req MigrateCronsNamespacesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Flip the feature flag so subsequent EnsureQueues calls target the correct namespace.
	if err := s.features.Enable(ctx, org.ID, map[string]bool{
		string(app.OrgFeatureCronNamespaceIsolation): req.Enabled,
	}); err != nil {
		s.l.Error("unable to set cron-namespace-isolation feature", zap.String("org_id", org.ID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to set feature flag: " + err.Error()})
		return
	}

	// Ensure the org-signals queue exists (needed to enqueue the migration signal).
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

	// Re-ensure all of the org's queues; EnsureInstallQueues/EnsureRunnerQueues
	// now migrate cron queues to the namespace dictated by the feature flag.
	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &queuemigration.Signal{OrgID: org.ID},
	}); err != nil {
		s.l.Error("unable to enqueue migration signal", zap.String("org_id", org.ID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unable to enqueue migration signal: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, true)
}
