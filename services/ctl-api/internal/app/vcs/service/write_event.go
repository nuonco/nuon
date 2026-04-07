package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	ghevent "github.com/nuonco/nuon/pkg/github/event"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type writeEventResponse struct {
	WebhookEventID string `json:"webhook_event_id"`
	SignalsSent    int    `json:"signals_sent"`
	Error          string `json:"error,omitempty"`
}

// @ID						WriteVCSEvent
// @Summary					Write a VCS webhook event
// @Description				Receives incoming GitHub webhook events and fans out to VCS connections with configs matching the repo
// @Tags					vcs
// @Accept					json
// @Produce					json
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	writeEventResponse
// @Router					/v1/vcs/events [post]
func (s *service) WriteEvent(ctx *gin.Context) {
	rawBody, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to read request body: %w", err))
		return
	}

	// Parse the JSON payload for field extraction.
	var payload map[string]any
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		ctx.Error(fmt.Errorf("unable to parse event payload: %w", err))
		return
	}

	eventType := ctx.GetHeader("X-GitHub-Event")
	if eventType == "" {
		eventType = "unknown"
	}

	body := &blobstore.Blob{}
	body.SetPrefix("webhooks")
	body.SetContentType("application/json")
	body.Set(string(rawBody))

	webhookEvent := app.VCSWebhookEvent{
		EventType: eventType,
		Body:      body,
		Status:    string(app.VCSWebhookEventStatusPending),
	}
	if err := s.db.WithContext(ctx).Create(&webhookEvent).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to store webhook event: %w", err))
		return
	}

	s.l.Debug("received vcs webhook event",
		zap.String("event_type", eventType),
		zap.String("webhook_event_id", webhookEvent.ID),
	)

	resp := writeEventResponse{
		WebhookEventID: webhookEvent.ID,
	}

	repoOwner, repoName, err := ghevent.ExtractRepoInfo(payload)
	if err != nil {
		s.l.Warn("unable to extract repo info from webhook payload",
			zap.String("webhook_event_id", webhookEvent.ID),
			zap.Error(err))
		resp.Error = fmt.Sprintf("unable to extract repo info: %s", err)
		ctx.JSON(http.StatusOK, resp)
		return
	}

	pushEvent := ghevent.ParsePushEvent(payload)
	s.l.Info("extracted repo info from vcs webhook",
		zap.String("webhook_event_id", webhookEvent.ID),
		zap.String("repo_owner", repoOwner),
		zap.String("repo_name", repoName),
		zap.String("branch", pushEvent.Branch),
	)

	signalsSent, fanOutErr := s.fanOutEvent(ctx, webhookEvent.ID, repoOwner, repoName)
	resp.SignalsSent = signalsSent

	if fanOutErr != nil {
		s.l.Error("fan out failed",
			zap.String("webhook_event_id", webhookEvent.ID),
			zap.Error(fanOutErr))
		resp.Error = fanOutErr.Error()
		ctx.JSON(http.StatusOK, resp)
		return
	}

	s.db.WithContext(ctx).Model(&webhookEvent).Update("status", string(app.VCSWebhookEventStatusEmitted))

	ctx.JSON(http.StatusOK, resp)
}

// fanOutEvent finds ConnectedGithubVCSConfigs and PublicGitVCSConfigs matching
// the repo owner and name, then signals each distinct VCSConnection queue.
func (s *service) fanOutEvent(ctx *gin.Context, webhookEventID, repoOwner, repoName string) (int, error) {
	// Collect VCSConnection IDs from both config types. Deduplicate by ID.
	connIDs := make(map[string]bool)

	// 1. Connected GitHub VCS configs — have a direct VCSConnectionID.
	var ghConfigs []app.ConnectedGithubVCSConfig
	if err := s.db.WithContext(ctx).
		Where("repo_owner = ? AND repo_name = ?", repoOwner, repoName).
		Find(&ghConfigs).Error; err != nil {
		return 0, fmt.Errorf("unable to find connected vcs configs: %w", err)
	}
	for _, cfg := range ghConfigs {
		connIDs[cfg.VCSConnectionID] = true
	}

	// 2. Public Git VCS configs — no VCSConnectionID, so look up a
	//    VCSConnection for the matching org.
	repoSlug := fmt.Sprintf("%s/%s", repoOwner, repoName)
	repoHTTPS := fmt.Sprintf("https://github.com/%s/%s", repoOwner, repoName)
	var publicConfigs []app.PublicGitVCSConfig
	if err := s.db.WithContext(ctx).
		Where("repo IN ?", []string{repoSlug, repoHTTPS}).
		Find(&publicConfigs).Error; err != nil {
		s.l.Warn("unable to find public git vcs configs",
			zap.String("webhook_event_id", webhookEventID),
			zap.Error(err))
	}

	// Collect unique org IDs that need a VCSConnection lookup,
	// excluding orgs already covered by connected github configs.
	orgIDs := make(map[string]bool)
	for _, cfg := range publicConfigs {
		orgIDs[cfg.OrgID] = true
	}
	if len(connIDs) > 0 {
		// Look up orgs for already-found connections to exclude them.
		existingIDs := make([]string, 0, len(connIDs))
		for id := range connIDs {
			existingIDs = append(existingIDs, id)
		}
		var existingConns []app.VCSConnection
		if err := s.db.WithContext(ctx).
			Select("id", "org_id").
			Where("id IN ?", existingIDs).
			Find(&existingConns).Error; err == nil {
			for _, c := range existingConns {
				delete(orgIDs, c.OrgID)
			}
		}
	}
	if len(orgIDs) > 0 {
		ids := make([]string, 0, len(orgIDs))
		for id := range orgIDs {
			ids = append(ids, id)
		}
		var connections []app.VCSConnection
		if err := s.db.WithContext(ctx).
			Where("org_id IN ?", ids).
			Find(&connections).Error; err != nil {
			s.l.Warn("unable to find vcs connections for public git orgs",
				zap.String("webhook_event_id", webhookEventID),
				zap.Error(err))
		}
		for _, conn := range connections {
			connIDs[conn.ID] = true
		}
	}

	if len(connIDs) == 0 {
		s.l.Warn("no vcs configs found for repo",
			zap.String("repo_owner", repoOwner),
			zap.String("repo_name", repoName),
			zap.String("webhook_event_id", webhookEventID))
		return 0, nil
	}

	signalsSent := 0
	for connID := range connIDs {
		if err := s.helpers.EnqueueGithubEventSignal(ctx, connID, webhookEventID); err != nil {
			s.l.Warn("unable to enqueue github event signal",
				zap.String("vcs_connection_id", connID),
				zap.String("webhook_event_id", webhookEventID),
				zap.Error(err),
			)
			continue
		}
		signalsSent++
	}

	return signalsSent, nil
}
