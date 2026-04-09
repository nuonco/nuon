package queue

import (
	"time"

	"github.com/pkg/errors"
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
	if err := workflow.Await(ctx, func() bool {
		return q.ready
	}); err != nil {
		return nil, errors.Wrap(err, "unable to await for ready")
	}

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
	if err := workflow.Await(ctx, func() bool {
		return q.ready
	}); err != nil {
		return nil, errors.Wrap(err, "unable to await for ready")
	}

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
