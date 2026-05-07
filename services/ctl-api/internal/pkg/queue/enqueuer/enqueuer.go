package enqueuer

import (
	"context"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

const (
	enqueueChannelSize = 1000
	processOneTimeout  = 30 * time.Second
)

// Enqueuer processes signal enqueue requests in the background. It receives
// queue signal IDs via a channel and performs the UpdateWithStart call to
// enqueue them into their respective queue workflows.
type Enqueuer struct {
	db      *gorm.DB
	cfg     *internal.Config
	tClient temporalclient.Client
	l       *zap.Logger

	ctx    context.Context
	cancel context.CancelFunc

	ch     chan string
	stopCh chan struct{}
	doneCh chan struct{}
}

type Params struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Cfg     *internal.Config
	TClient temporalclient.Client
	L       *zap.Logger
	LC      fx.Lifecycle
}

func New(params Params) *Enqueuer {
	ctx, cancel := context.WithCancel(context.Background())

	e := &Enqueuer{
		db:      params.DB,
		cfg:     params.Cfg,
		tClient: params.TClient,
		l:       params.L.Named("queue-enqueuer"),
		ctx:     ctx,
		cancel:  cancel,
		ch:      make(chan string, enqueueChannelSize),
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}

	params.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go e.run()
			e.startSweepWorkflow(ctx)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			e.cancel()
			close(e.stopCh)
			select {
			case <-e.doneCh:
			case <-ctx.Done():
			}
			return nil
		},
	})

	return e
}

// Send enqueues a queue signal ID for background processing. If the channel
// is full the ID is dropped — the AwaitSignal inline path will pick it up.
func (e *Enqueuer) Send(queueSignalID string) {
	select {
	case e.ch <- queueSignalID:
	default:
		e.l.Warn("enqueue channel full, signal will be enqueued inline by AwaitSignal",
			zap.String("queue-signal-id", queueSignalID))
	}
}

func (e *Enqueuer) run() {
	defer close(e.doneCh)

	for {
		select {
		case <-e.stopCh:
			e.drain()
			return
		case id := <-e.ch:
			e.processOne(id)
		}
	}
}

// startSweepWorkflow starts the EnqueuerSweep workflow if not already running.
func (e *Enqueuer) startSweepWorkflow(ctx context.Context) {
	opts := tclient.StartWorkflowOptions{
		ID:                       "enqueuer-sweep",
		TaskQueue:                workflows.APITaskQueue,
		WorkflowIDConflictPolicy: enumsv1.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}

	type sweepReq struct{}
	_, err := e.tClient.ExecuteWorkflowInNamespace(ctx, "general", opts, "EnqueuerSweep", sweepReq{})
	if err != nil {
		e.l.Warn("unable to start enqueuer sweep workflow", zap.Error(err))
		return
	}
	e.l.Info("enqueuer sweep workflow started")
}

// drain processes any remaining channel items during shutdown.
func (e *Enqueuer) drain() {
	for {
		select {
		case id := <-e.ch:
			e.processOne(id)
		default:
			return
		}
	}
}
