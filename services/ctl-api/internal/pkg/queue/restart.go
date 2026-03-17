package queue

import "go.temporal.io/sdk/workflow"

const RestartUpdateName string = "restart"

type RestartRequest struct{}

type RestartResponse struct{}

func (q *queue) restartUpdateHandler(ctx workflow.Context, req *RestartRequest) (*RestartResponse, error) {
	// Only trigger a continue-as-new if the queue was already running (ready).
	// When UpdateWithStart creates a fresh queue because the previous one was stopped
	// or terminated, the restart update arrives before q.ready is set. In that case
	// the queue is already starting fresh, so there is no need for continue-as-new.
	if q.ready {
		q.restarted = true
	}
	return &RestartResponse{}, nil
}
