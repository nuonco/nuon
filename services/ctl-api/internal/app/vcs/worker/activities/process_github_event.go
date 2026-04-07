package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	ghevent "github.com/nuonco/nuon/pkg/github/event"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ProcessGithubEventRequest struct {
	VCSConnectionID string `validate:"required"`
	WebhookEventID  string `validate:"required"`
}

type ProcessGithubEventResponse struct {
	EventType      string `json:"event_type"`
	CommitsCreated int    `json:"commits_created"`
}

// @temporal-gen-v2 activity
func (a *Activities) ProcessGithubEvent(ctx context.Context, req ProcessGithubEventRequest) (*ProcessGithubEventResponse, error) {
	var webhookEvent app.VCSWebhookEvent
	if err := a.db.WithContext(ctx).First(&webhookEvent, "id = ?", req.WebhookEventID).Error; err != nil {
		return nil, fmt.Errorf("unable to get webhook event: %w", err)
	}

	var conn app.VCSConnection
	if err := a.db.WithContext(ctx).First(&conn, "id = ?", req.VCSConnectionID).Error; err != nil {
		return nil, fmt.Errorf("unable to get vcs connection: %w", err)
	}

	// Parse the webhook body into a payload map.
	payload, err := webhookEvent.ParseBody(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to parse webhook body: %w", err)
	}

	resp := &ProcessGithubEventResponse{
		EventType: webhookEvent.EventType,
	}

	switch webhookEvent.EventType {
	case "push":
		count, err := a.processPushEvent(ctx, payload, &conn)
		if err != nil {
			return nil, fmt.Errorf("unable to process push event: %w", err)
		}
		resp.CommitsCreated = count
	}

	return resp, nil
}

func (a *Activities) processPushEvent(ctx context.Context, payload map[string]any, conn *app.VCSConnection) (int, error) {
	push := ghevent.ParsePushEvent(payload)
	if len(push.Commits) == 0 {
		return 0, nil
	}

	created := 0
	for _, c := range push.Commits {
		var existing app.VCSConnectionCommit
		err := a.db.WithContext(ctx).
			Where("sha = ? AND owner_id = ? AND owner_type = ? AND branch = ?", c.SHA, conn.ID, "vcs_connections", push.Branch).
			First(&existing).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return created, fmt.Errorf("unable to check existing commit: %w", err)
		}

		vcsCommit := app.VCSConnectionCommit{
			OrgID:           conn.OrgID,
			VCSConnectionID: &conn.ID,
			OwnerID:         conn.ID,
			OwnerType:       "vcs_connections",
			SHA:             c.SHA,
			Branch:          push.Branch,
			RepoOwner:       push.RepoOwner,
			RepoName:        push.RepoName,
			Source:          app.VCSCommitSourceWebhook,
			AuthorName:      c.AuthorName,
			AuthorEmail:     c.AuthorEmail,
			Message:         c.Message,
		}

		if err := a.db.WithContext(ctx).Create(&vcsCommit).Error; err != nil {
			return created, fmt.Errorf("unable to create vcs commit: %w", err)
		}
		created++
	}

	return created, nil
}
