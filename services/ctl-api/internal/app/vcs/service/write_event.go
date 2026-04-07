package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						WriteVCSEvent
// @Summary					Write a VCS webhook event
// @Description				Writes incoming webhook events for a VCS connection
// @Param					vcs_connection_id	path	string	true	"VCS Connection ID"
// @Tags					vcs
// @Accept					json
// @Produce					json
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.VCSEvent
// @Router					/v1/vcs/{vcs_connection_id}/events [post]
func (s *service) WriteEvent(ctx *gin.Context) {
	vcsConnectionID := ctx.Param("vcs_connection_id")

	// Verify the VCS connection exists.
	var vcsConn app.VCSConnection
	if err := s.db.WithContext(ctx).First(&vcsConn, "id = ?", vcsConnectionID).Error; err != nil {
		ctx.Error(fmt.Errorf("vcs connection not found: %w", err))
		return
	}

	// Read the raw JSON payload.
	var payload app.VCSEventPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.Error(fmt.Errorf("unable to parse event payload: %w", err))
		return
	}

	// Extract event type from GitHub header if present.
	eventType := ctx.GetHeader("X-GitHub-Event")
	if eventType == "" {
		eventType = "unknown"
	}

	event := app.VCSEvent{
		OrgID:           vcsConn.OrgID,
		VCSConnectionID: vcsConnectionID,
		EventType:       eventType,
		Payload:         payload,
		Status: &app.CompositeStatus{
			CreatedAtTS:            time.Now().Unix(),
			Status:                 app.StatusSuccess,
			StatusHumanDescription: fmt.Sprintf("received %s event", eventType),
		},
	}

	if err := s.db.WithContext(ctx).Create(&event).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to store vcs event: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, event)
}
