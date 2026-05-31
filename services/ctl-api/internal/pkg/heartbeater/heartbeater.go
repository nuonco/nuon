package heartbeater

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	defaultFlushInterval = 5 * time.Second
	defaultBatchSize     = 1000
	channelSize          = 10000
)

type Heartbeater struct {
	chDB *gorm.DB
	l    *zap.Logger
	mw   metrics.Writer

	ch     chan app.RunnerHeartBeat
	stopCh chan struct{}
	doneCh chan struct{}

	flushInterval time.Duration
	batchSize     int
}

type Params struct {
	fx.In

	CHDB *gorm.DB `name:"ch"`
	L    *zap.Logger
	MW   metrics.Writer
	LC   fx.Lifecycle
	Cfg  *internal.Config
}

func New(params Params) *Heartbeater {
	flushInterval := defaultFlushInterval
	if params.Cfg.HeartbeaterFlushInterval > 0 {
		flushInterval = params.Cfg.HeartbeaterFlushInterval
	}

	batchSize := defaultBatchSize
	if params.Cfg.HeartbeaterBatchSize > 0 {
		batchSize = params.Cfg.HeartbeaterBatchSize
	}

	h := &Heartbeater{
		chDB:          params.CHDB,
		l:             params.L.Named("heartbeater"),
		mw:            params.MW,
		ch:            make(chan app.RunnerHeartBeat, channelSize),
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
		flushInterval: flushInterval,
		batchSize:     batchSize,
	}

	params.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go h.run()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			close(h.stopCh)
			<-h.doneCh
			return nil
		},
	})

	return h
}

// Send enqueues a heartbeat for batched writing to ClickHouse.
// The heartbeat must have ID, CreatedByID, and CreatedAt pre-populated.
func (h *Heartbeater) Send(hb app.RunnerHeartBeat) {
	select {
	case h.ch <- hb:
	default:
		h.l.Warn("heartbeater channel full, dropping heartbeat",
			zap.String("runner_id", hb.RunnerID),
			zap.String("process_id", hb.ProcessID),
		)
	}
}

func (h *Heartbeater) run() {
	defer close(h.doneCh)

	ticker := time.NewTicker(h.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			h.drainAndFlush()
		case <-h.stopCh:
			h.drainAndFlush()
			return
		}
	}
}

func (h *Heartbeater) drainAndFlush() {
	batch := make([]app.RunnerHeartBeat, 0, h.batchSize)
	for {
		select {
		case hb := <-h.ch:
			batch = append(batch, hb)
			if len(batch) >= h.batchSize {
				h.flush(batch)
				batch = make([]app.RunnerHeartBeat, 0, h.batchSize)
			}
		default:
			if len(batch) > 0 {
				h.flush(batch)
			}
			return
		}
	}
}

func (h *Heartbeater) flush(batch []app.RunnerHeartBeat) {
	if len(batch) == 0 {
		return
	}

	if res := h.chDB.CreateInBatches(&batch, len(batch)); res.Error != nil {
		h.l.Error("failed to flush heartbeats to clickhouse",
			zap.Int("batch_size", len(batch)),
			zap.Error(res.Error),
		)
		h.mw.Incr("heartbeater.flush.error", nil)
		return
	}

	h.mw.Gauge("heartbeater.flush.count", float64(len(batch)), nil)
}
