package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type GetGithubEventRequest struct {
	GithubEventID string `json:"github_event_id" validate:"required"`
}

type GetGithubEventResponse struct {
	GithubEvent *app.GithubEvent `json:"github_event"`
	// Payload is the parsed webhook payload loaded from blob storage.
	Payload map[string]any `json:"payload"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetGithubEvent(ctx context.Context, req GetGithubEventRequest) (*GetGithubEventResponse, error) {
	dbCtx := blobstore.WithBlobService(ctx, a.blobSvc)
	dbCtx = blobstore.WithBlobAutoLoad(dbCtx, true)

	var event app.GithubEvent
	if err := a.db.WithContext(dbCtx).First(&event, "id = ?", req.GithubEventID).Error; err != nil {
		return nil, fmt.Errorf("unable to get github event: %w", err)
	}

	// Parse the blob payload into a map for use in workflows.
	var payload map[string]any
	if event.Payload != nil && event.Payload.IsSet() {
		payloadStr, err := event.Payload.Get(dbCtx)
		if err != nil {
			return nil, fmt.Errorf("unable to load github event payload from blob: %w", err)
		}
		if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
			return nil, fmt.Errorf("unable to parse github event payload: %w", err)
		}
	}

	return &GetGithubEventResponse{
		GithubEvent: &event,
		Payload:     payload,
	}, nil
}
