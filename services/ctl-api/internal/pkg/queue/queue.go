package queue

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/go-playground/validator/v10"

	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

const (
	AppTriggersQueueName = queuenames.AppTriggersQueueName
	OrgSignalsQueueName  = queuenames.OrgSignalsQueueName
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
	// why: Queues outlive individual signals and must not inherit their log streams,
	// including incomplete streams persisted in the context of a retrying queue.
	ctx = cctx.ClearLogStreamWorkflowContext(ctx)

	q := &queue{
		cfg:             w.cfg,
		v:               w.v,
		mw:              w.mw,
		queueID:         req.QueueID,
		state:           req.State,
		releaseWindow:   req.ReleaseWindow,
		inFlightSignals: make(map[string]bool),
	}
	if q.state == nil {
		q.state = &QueueState{
			QueueRefs: make([]QueueRef, 0),
		}
	}
	q.paused = q.state.Paused
	q.lastActivityTime = q.state.LastActivityTime

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
		ctx = cctx.SetLogStreamWorkflowContext(ctx, nil)
		return workflow.NewContinueAsNewError(ctx, w.Queue, req)
	}

	return nil
}

type queue struct {
	cfg *internal.Config
	v   *validator.Validate
	mw  tmetrics.Writer

	queueID string

	releaseWindow *ReleaseWindow

	ready       bool
	stopped     bool
	restarted   bool
	paused      bool
	maxDepth    int
	maxInFlight int

	sem workflow.Semaphore

	idleTimeout time.Duration

	lastActivityTime time.Time

	activeWorkers int

	inFlightSignals map[string]bool

	state *QueueState
	ch    workflow.Channel
}
