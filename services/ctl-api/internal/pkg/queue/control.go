package queue

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

const (
	PauseHandlerName  string = "pause"
	ResumeHandlerName string = "resume"
)

type PauseRequest struct{}
type PauseResponse struct{}

// @temporal-gen-v2 update
// @id pause
func (q *queue) pauseHandler(ctx workflow.Context, req *PauseRequest) (*PauseResponse, error) {
	q.paused = true
	q.state.Paused = true

	if err := activities.AwaitUpdateQueuePaused(ctx, activities.UpdateQueuePausedRequest{
		QueueID: q.queueID,
		Paused:  true,
	}, &workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}); err != nil {
		return nil, err
	}

	return &PauseResponse{}, nil
}

type ResumeRequest struct{}
type ResumeResponse struct{}

// @temporal-gen-v2 update
// @id resume
func (q *queue) resumeHandler(ctx workflow.Context, req *ResumeRequest) (*ResumeResponse, error) {
	q.paused = false
	q.state.Paused = false

	if err := activities.AwaitUpdateQueuePaused(ctx, activities.UpdateQueuePausedRequest{
		QueueID: q.queueID,
		Paused:  false,
	}, &workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}); err != nil {
		return nil, err
	}

	return &ResumeResponse{}, nil
}
