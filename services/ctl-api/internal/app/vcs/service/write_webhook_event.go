package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Matches the public API cap; GitHub allows up to 25 MB but the largest payload seen is ~5 MB.
const maxWebhookBody = 16 << 20

var errWebhookBodyTooLarge = errors.New("webhook body too large")

// readBody bounds the read because the HMAC can only be checked after the whole body is in memory.
func readBody(ctx *gin.Context) ([]byte, error) {
	if ctx.Request.ContentLength > maxWebhookBody {
		return nil, errWebhookBodyTooLarge
	}
	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, maxWebhookBody+1))
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return nil, errWebhookBodyTooLarge
		}
		return nil, fmt.Errorf("unable to read request body: %w", err)
	}
	if len(body) > maxWebhookBody {
		return nil, errWebhookBodyTooLarge
	}
	return body, nil
}

func verifyGitHubSignature(secret string, signature string, body []byte) bool {
	if signature == "" || secret == "" {
		return false
	}

	sig := strings.TrimPrefix(signature, "sha256=")
	if sig == signature {
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

	signature := ctx.GetHeader("X-Hub-Signature-256")
	if !strings.HasPrefix(signature, "sha256=") {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	var sub app.VCSWebhookSubscription
	if err := s.db.WithContext(ctx).First(&sub, "id = ?", subscriptionID).Error; err != nil {
		ctx.Error(fmt.Errorf("webhook subscription not found: %w", err))
		return
	}

	body, err := readBody(ctx)
	if errors.Is(err, errWebhookBodyTooLarge) {
		ctx.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		ctx.Error(err)
		return
	}

	if !verifyGitHubSignature(sub.WebhookSecret, signature, body) {
		s.l.Warn("webhook signature verification failed",
			zap.String("subscription_id", subscriptionID),
		)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	eventType := ctx.GetHeader("X-GitHub-Event")
	if eventType == "" {
		eventType = "unknown"
	}

	event, err := s.createGithubEvent(ctx.Request.Context(), sub.GithubInstallID, eventType, body)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.l.Info("stored github event",
		zap.String("event_id", event.ID),
		zap.String("event_type", eventType),
		zap.String("github_install_id", sub.GithubInstallID),
	)

	s.fanOutToVCSConnections(ctx.Request.Context(), event)

	ctx.JSON(http.StatusOK, event)
}
