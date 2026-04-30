package enqueuer

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

const (
	enqueueChannelSize = 1000
	sweepInterval      = 1 * time.Minute
	sweepBatchSize     = 100
)

const (
	sweepTimeout      = 30 * time.Second
	processOneTimeout = 30 * time.Second
)

// Enqueuer processes signal enqueue requests in the background. It receives
// queue signal IDs via a channel and also periodically sweeps for orphaned
// signals that were never enqueued.
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
		OnStart: func(context.Context) error {
			go e.run()
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
// is full the ID is dropped with a warning — the periodic sweep will recover it.
func (e *Enqueuer) Send(queueSignalID string) {
	select {
	case e.ch <- queueSignalID:
	default:
		e.l.Warn("enqueue channel full, signal will be retried by sweep",
			zap.String("queue-signal-id", queueSignalID))
	}
}

func (e *Enqueuer) run() {
	defer close(e.doneCh)

	// ticker := time.NewTicker(sweepInterval)
	// defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			e.drain()
			return
		case id := <-e.ch:
			e.processOne(id)
			// NOTE(jm): we no longer run the sweep in process
			// case <-ticker.C:
			// e.sweep()
		}
	}
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
