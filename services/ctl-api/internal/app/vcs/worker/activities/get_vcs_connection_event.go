package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type GetVCSConnectionEventRequest struct {
	VCSConnectionEventID string `json:"vcs_connection_event_id" validate:"required"`
}

type GetVCSConnectionEventResponse struct {
	VCSConnectionEvent *app.VCSConnectionEvent `json:"vcs_connection_event"`
	GithubEvent        *app.GithubEvent        `json:"github_event"`
	Payload            map[string]any          `json:"payload"`
}

// @temporal-gen-v2 activity
func (a *Activities) GetVCSConnectionEvent(ctx context.Context, req GetVCSConnectionEventRequest) (*GetVCSConnectionEventResponse, error) {
	var connEvent app.VCSConnectionEvent
	if err := a.db.WithContext(ctx).First(&connEvent, "id = ?", req.VCSConnectionEventID).Error; err != nil {
		return nil, fmt.Errorf("unable to get vcs connection event: %w", err)
	}

	dbCtx := blobstore.WithBlobService(ctx, a.blobSvc)
	dbCtx = blobstore.WithBlobAutoLoad(dbCtx, true)

	var event app.GithubEvent
	if err := a.db.WithContext(dbCtx).First(&event, "id = ?", connEvent.GithubEventID).Error; err != nil {
		return nil, fmt.Errorf("unable to get github event: %w", err)
	}

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

	return &GetVCSConnectionEventResponse{
		VCSConnectionEvent: &connEvent,
		GithubEvent:        &event,
		Payload:            payload,
	}, nil
}
