package service

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

// @ID						WriteVCSEvent
// @Summary					Write a VCS webhook event
// @Description				Writes incoming webhook events for a VCS connection (legacy endpoint)
// @Param					vcs_connection_id	path	string	true	"VCS Connection ID"
// @Tags					vcs
// @Accept					json
// @Produce					json
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.GithubEvent
// @Router					/v1/vcs/{vcs_connection_id}/events [post]
func (s *service) WriteEvent(ctx *gin.Context) {
	vcsConnectionID := ctx.Param("vcs_connection_id")

	// Verify the VCS connection exists and get its install ID.
	var vcsConn app.VCSConnection
	if err := s.db.WithContext(ctx).First(&vcsConn, "id = ?", vcsConnectionID).Error; err != nil {
		ctx.Error(fmt.Errorf("vcs connection not found: %w", err))
		return
	}

	// Read the raw body.
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to read request body: %w", err))
		return
	}

	// Extract event type from GitHub header if present.
	eventType := ctx.GetHeader("X-GitHub-Event")
	if eventType == "" {
		eventType = "unknown"
	}

	// Extract GitHub installation ID from header, fall back to connection's value.
	githubInstallID := ctx.GetHeader("X-GitHub-Hook-Installation-Target-ID")
	if githubInstallID == "" {
		githubInstallID = vcsConn.GithubInstallID
	}

	// Create blob payload for S3 storage.
	payload := &blobstore.Blob{}
	payload.Set(string(body))
	payload.SetContentType("application/json")
	payload.SetS3Prefix("blobs/github_events")

	// Set blob service on context for the GORM hook.
	dbCtx := blobstore.WithBlobService(ctx.Request.Context(), s.blobSvc)
	dbCtx = blobstore.WithBlobWriteEnabled(dbCtx, true)

	event := app.GithubEvent{
		GithubInstallID: githubInstallID,
		EventType:       eventType,
		Payload:         payload,
		Status: &app.CompositeStatus{
			CreatedAtTS:            time.Now().Unix(),
			Status:                 app.StatusSuccess,
			StatusHumanDescription: fmt.Sprintf("received %s event", eventType),
		},
	}

	if err := s.db.WithContext(dbCtx).Create(&event).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to store github event: %w", err))
		return
	}

	// Find the webhook subscription for this connection to enqueue the signal.
	var sub app.VCSWebhookSubscription
	if err := s.db.WithContext(ctx).
		Where(app.VCSWebhookSubscription{GithubInstallID: githubInstallID}).
		First(&sub).Error; err != nil {
		// No subscription — try to enqueue directly to the connection's queue as fallback.
		s.l.Warn("no webhook subscription found for legacy event, skipping signal enqueue",
			zap.String("vcs_connection_id", vcsConnectionID),
			zap.String("event_id", event.ID),
			zap.Error(err),
		)
		ctx.JSON(http.StatusOK, event)
		return
	}

	// Enqueue signal to process this event (non-blocking).
	if err := s.helpers.EnqueueGithubEvent(ctx, &sub, event.ID); err != nil {
		s.l.Warn("failed to enqueue github event signal",
			zap.String("vcs_connection_id", vcsConnectionID),
			zap.String("event_id", event.ID),
			zap.Error(err),
		)
	}

	ctx.JSON(http.StatusOK, event)
}
