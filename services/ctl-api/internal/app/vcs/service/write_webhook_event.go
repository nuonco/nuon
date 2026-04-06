package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

// verifyGitHubSignature validates the X-Hub-Signature-256 header against the raw request body.
func verifyGitHubSignature(secret string, signature string, body []byte) bool {
	if signature == "" || secret == "" {
		return false
	}

	// GitHub sends "sha256=<hex>"
	sig := strings.TrimPrefix(signature, "sha256=")
	if sig == signature {
		// No "sha256=" prefix — invalid format
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expected))
}

// @ID						WriteWebhookEvent
// @Summary					Write a VCS webhook event (shared per subscription)
// @Description				Receives webhook events for a webhook subscription and creates a GithubEvent for processing
// @Param					subscription_id	path	string	true	"Webhook Subscription ID"
// @Tags					vcs
// @Accept					json
// @Produce					json
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.GithubEvent
// @Router					/v1/vcs/webhooks/{subscription_id}/events [post]
func (s *service) WriteWebhookEvent(ctx *gin.Context) {
	subscriptionID := ctx.Param("subscription_id")

	// Look up the webhook subscription.
	var sub app.VCSWebhookSubscription
	if err := s.db.WithContext(ctx).First(&sub, "id = ?", subscriptionID).Error; err != nil {
		ctx.Error(fmt.Errorf("webhook subscription not found: %w", err))
		return
	}

	// Read the raw body for signature verification.
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to read request body: %w", err))
		return
	}

	// Verify the GitHub webhook signature.
	signature := ctx.GetHeader("X-Hub-Signature-256")
	if !verifyGitHubSignature(sub.WebhookSecret, signature, body) {
		s.l.Warn("webhook signature verification failed",
			zap.String("subscription_id", subscriptionID),
		)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Extract event type from GitHub header.
	eventType := ctx.GetHeader("X-GitHub-Event")
	if eventType == "" {
		eventType = "unknown"
	}

	// Extract GitHub installation ID from header.
	githubInstallID := ctx.GetHeader("X-GitHub-Hook-Installation-Target-ID")
	if githubInstallID == "" {
		// Fall back to the subscription's known install ID.
		githubInstallID = sub.GithubInstallID
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

	// Enqueue signal to process this event (non-blocking).
	if err := s.helpers.EnqueueGithubEvent(ctx, &sub, event.ID); err != nil {
		s.l.Warn("failed to enqueue github event signal",
			zap.String("subscription_id", subscriptionID),
			zap.String("event_id", event.ID),
			zap.Error(err),
		)
	}

	ctx.JSON(http.StatusOK, event)
}
