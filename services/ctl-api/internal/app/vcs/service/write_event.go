package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	githubevent "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/signals/github_event"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (s *service) createGithubEvent(ctx context.Context, githubInstallID, eventType string, body []byte) (*app.GithubEvent, error) {
	payload := &blobstore.Blob{}
	payload.Set(string(body))
	payload.SetContentType("application/json")
	payload.SetS3Prefix("blobs/github_events")

	dbCtx := blobstore.WithBlobService(ctx, s.blobSvc)

	event := &app.GithubEvent{
		GithubInstallID: githubInstallID,
		EventType:       eventType,
		Payload:         payload,
		Status: &app.CompositeStatus{
			CreatedAtTS:            time.Now().Unix(),
			Status:                 app.StatusSuccess,
			StatusHumanDescription: fmt.Sprintf("received %s event", eventType),
		},
	}

	if err := s.db.WithContext(dbCtx).Create(event).Error; err != nil {
		return nil, fmt.Errorf("unable to store github event: %w", err)
	}

	return event, nil
}

func (s *service) fanOutToVCSConnections(ctx context.Context, event *app.GithubEvent) {
	var conns []app.VCSConnection
	if err := s.db.WithContext(ctx).
		Where(app.VCSConnection{GithubInstallID: event.GithubInstallID}).
		Find(&conns).Error; err != nil {
		s.l.Error("failed to find vcs connections for github install id",
			zap.String("github_install_id", event.GithubInstallID),
			zap.Error(err),
		)
		return
	}

	if len(conns) == 0 {
		s.l.Warn("no vcs connections found for github install id",
			zap.String("github_install_id", event.GithubInstallID),
		)
		return
	}

	for _, conn := range conns {
		connCtx := cctx.SetOrgIDContext(ctx, conn.OrgID)
		connCtx = cctx.SetAccountIDContext(connCtx, conn.CreatedByID)

		connEvent := app.VCSConnectionEvent{
			OrgID:           conn.OrgID,
			VCSConnectionID: conn.ID,
			GithubEventID:   event.ID,
			Status: &app.CompositeStatus{
				CreatedAtTS:            time.Now().Unix(),
				Status:                 app.StatusQueued,
				StatusHumanDescription: fmt.Sprintf("received %s event", event.EventType),
			},
		}

		if err := s.db.WithContext(connCtx).Create(&connEvent).Error; err != nil {
			s.l.Error("failed to create vcs connection event",
				zap.String("vcs_connection_id", conn.ID),
				zap.String("github_event_id", event.ID),
				zap.Error(err),
			)
			continue
		}

		queue, err := s.queueClient.GetQueueByOwner(connCtx, conn.ID, "vcs_connections")
		if err != nil {
			s.l.Warn("failed to get queue for vcs connection",
				zap.String("vcs_connection_id", conn.ID),
				zap.Error(err),
			)
			continue
		}

		_, err = s.queueClient.EnqueueSignal(connCtx, &queueclient.EnqueueSignalRequest{
			QueueID: queue.ID,
			Signal: &githubevent.Signal{
				VCSConnectionEventID: connEvent.ID,
			},
		})
		if err != nil {
			s.l.Warn("failed to enqueue github event signal",
				zap.String("vcs_connection_event_id", connEvent.ID),
				zap.Error(err),
			)
			continue
		}

		s.l.Info("enqueued github_event signal",
			zap.String("vcs_connection_event_id", connEvent.ID),
			zap.String("vcs_connection_id", conn.ID),
		)
	}
}

// @ID						WriteVCSEvent
// @Summary					Write a VCS webhook event
// @Description				Writes incoming webhook events for a VCS connection (legacy endpoint)
// @Param					vcs_connection_id	path	string	true	"VCS Connection ID"
// @Tags					vcs
// @Accept					json
// @Produce					json
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.GithubEvent
// @Router					/v1/vcs/{vcs_connection_id}/events [post]
func (s *service) WriteEvent(ctx *gin.Context) {
	vcsConnectionID := ctx.Param("vcs_connection_id")

	var vcsConn app.VCSConnection
	if err := s.db.WithContext(ctx).First(&vcsConn, "id = ?", vcsConnectionID).Error; err != nil {
		ctx.Error(fmt.Errorf("vcs connection not found: %w", err))
		return
	}

	body, err := readBody(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	eventType := ctx.GetHeader("X-GitHub-Event")
	if eventType == "" {
		eventType = "unknown"
	}

	event, err := s.createGithubEvent(ctx.Request.Context(), vcsConn.GithubInstallID, eventType, body)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.l.Info("stored github event (legacy)",
		zap.String("event_id", event.ID),
		zap.String("event_type", eventType),
		zap.String("github_install_id", vcsConn.GithubInstallID),
	)

	s.fanOutToVCSConnections(ctx.Request.Context(), event)

	ctx.JSON(http.StatusOK, event)
}
