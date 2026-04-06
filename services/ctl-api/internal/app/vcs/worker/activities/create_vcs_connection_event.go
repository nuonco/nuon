package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateVCSConnectionEventRequest struct {
	VCSConnectionID string `json:"vcs_connection_id" validate:"required"`
	GithubEventID   string `json:"github_event_id" validate:"required"`
	OrgID           string `json:"org_id" validate:"required"`
}

type CreateVCSConnectionEventResponse struct {
	VCSConnectionEventID string `json:"vcs_connection_event_id"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateVCSConnectionEvent(ctx context.Context, req CreateVCSConnectionEventRequest) (*CreateVCSConnectionEventResponse, error) {
	event := app.VCSConnectionEvent{
		OrgID:           req.OrgID,
		VCSConnectionID: req.VCSConnectionID,
		GithubEventID:   req.GithubEventID,
		Status: &app.CompositeStatus{
			CreatedAtTS:            time.Now().Unix(),
			Status:                 app.StatusSuccess,
			StatusHumanDescription: "vcs connection event created",
		},
	}

	if err := a.db.WithContext(ctx).Create(&event).Error; err != nil {
		return nil, fmt.Errorf("unable to create vcs connection event: %w", err)
	}

	return &CreateVCSConnectionEventResponse{
		VCSConnectionEventID: event.ID,
	}, nil
}
