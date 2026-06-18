package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func readBody(ctx *gin.Context) ([]byte, error) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read request body: %w", err)
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

	var sub app.VCSWebhookSubscription
	if err := s.db.WithContext(ctx).First(&sub, "id = ?", subscriptionID).Error; err != nil {
		ctx.Error(fmt.Errorf("webhook subscription not found: %w", err))
		return
	}

	body, err := readBody(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	signature := ctx.GetHeader("X-Hub-Signature-256")
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
