package queue

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type QueueWorkflowRequest struct {
	QueueID string
	Version string

	ReleaseWindow *ReleaseWindow

	State *QueueState
}

type QueueRef struct {
	WorkflowID string
	ID         string
}

// QueueState is the data that is passed between continue-as-news
type QueueState struct {
	QueueRefs        []QueueRef
	Paused           bool
	LastActivityTime time.Time
}

// @temporal-gen-v2 workflow
// @task-queue "queue"
// @id-template queue-{{.QueueID}}
// @memo type queue
func (w *Workflows) Queue(ctx workflow.Context, req QueueWorkflowRequest) error {
	q := &queue{
		cfg:           w.cfg,
		v:             w.v,
		queueID:       req.QueueID,
		state:         req.State,
		releaseWindow: req.ReleaseWindow,
	}
	if q.state == nil {
		q.state = &QueueState{
			QueueRefs: make([]QueueRef, 0),
		}
	}
	q.paused = q.state.Paused

	for _, hook := range w.StartupHooks {
		if err := hook(ctx, req); err != nil {
			return err
		}
	}

	finished, err := q.run(ctx)
	if err != nil {
		return err
	}
	if !finished {
		req.State = q.state
		req.State.LastActivityTime = q.lastActivityTime
		// Clear the log stream from context before continue-as-new so the next
		// run doesn't inherit a stale log stream from a previously executed signal.
		ctx = cctx.SetLogStreamWorkflowContext(ctx, nil)
		return workflow.NewContinueAsNewError(ctx, w.Queue, req)
	}

	return nil
}

type queue struct {
	cfg *internal.Config
	v   *validator.Validate

	queueID string

	releaseWindow *ReleaseWindow

	ready          bool
	stopped        bool
	restarted      bool
	forceRestarted bool
	paused         bool
	maxDepth       int
	maxInFlight    int

	// sem limits the number of concurrently processing signals to maxInFlight.
	sem workflow.Semaphore

	// idleTimeout is how long the queue can be idle before terminating.
	// Loaded from the queue's DB record, falling back to config default.
	idleTimeout time.Duration

	// lastActivityTime tracks when any worker last received a signal or when the queue started.
	// Used to detect idle queues that should terminate to free resources.
	lastActivityTime time.Time

	// activeWorkers tracks the number of workers currently processing a signal.
	// Used to prevent continue-as-new while workers are mid-processing.
	activeWorkers int

	// state is used to store state that will continue between continue-as-news
	state *QueueState
	ch    workflow.Channel
}
