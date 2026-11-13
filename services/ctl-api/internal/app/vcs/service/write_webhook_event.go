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

const maxWebhookBody = 25 << 20

func parseGitHubSignature(header string) ([]byte, bool) {
	sig, ok := strings.CutPrefix(header, "sha256=")
	if !ok || len(sig) != hex.EncodedLen(sha256.Size) {
		return nil, false
	}

	decoded, err := hex.DecodeString(sig)
	if err != nil {
		return nil, false
	}
	return decoded, true
}

func readBody(ctx *gin.Context) ([]byte, bool, error) {
	body, err := io.ReadAll(io.LimitReader(ctx.Request.Body, maxWebhookBody+1))
	if err != nil {
		return nil, false, fmt.Errorf("unable to read request body: %w", err)
	}
	if len(body) > maxWebhookBody {
		return nil, true, nil
	}
	return body, false, nil
}

func verifyGitHubSignature(secret string, signature []byte, body []byte) bool {
	if secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hmac.Equal(signature, mac.Sum(nil))
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
// @Failure					413	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.GithubEvent
// @Router					/v1/vcs/webhooks/{subscription_id}/events [post]
func (s *service) WriteWebhookEvent(ctx *gin.Context) {
	subscriptionID := ctx.Param("subscription_id")

	signature, ok := parseGitHubSignature(ctx.GetHeader("X-Hub-Signature-256"))
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	if ctx.Request.ContentLength > maxWebhookBody {
		ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
		return
	}

	var sub app.VCSWebhookSubscription
	if err := s.db.WithContext(ctx).First(&sub, "id = ?", subscriptionID).Error; err != nil {
		ctx.Error(fmt.Errorf("webhook subscription not found: %w", err))
		return
	}

	body, tooLarge, err := readBody(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if tooLarge {
		ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
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
