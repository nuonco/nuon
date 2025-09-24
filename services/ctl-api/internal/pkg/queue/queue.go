package queue

import (
	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type QueueWorkflowRequest struct {
	QueueID string
	Version string

	State *QueueState
}

type QueueRef struct {
	WorkflowID string
	ID         string
}

// QueueState is the data that is passed between continue-as-news
type QueueState struct {
	QueueRefs []QueueRef
}

// @temporal-gen workflow
// @task-queue "queue"
// @id-template queue-{{.QueueID}}
func (w *Workflows) Queue(ctx workflow.Context, req QueueWorkflowRequest) error {
	q := &queue{
		cfg:     w.cfg,
		v:       w.v,
		queueID: req.QueueID,
		state:   req.State,
	}
	if q.state == nil {
		q.state = &QueueState{
			QueueRefs: make([]QueueRef, 0),
		}
	}

	finished, err := q.run(ctx)
	if err != nil {
		return err
	}
	if !finished {
		req.State = q.state
		return workflow.NewContinueAsNewError(ctx, w.Queue, req)
	}

	return nil
}

type queue struct {
	cfg *internal.Config
	v   *validator.Validate

	queueID string

	ready     bool
	stopped   bool
	restarted bool

	// state is used to store state that will continue between continue-as-news
	state *QueueState
	ch    workflow.Channel
}
