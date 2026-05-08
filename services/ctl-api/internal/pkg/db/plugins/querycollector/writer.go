package querycollector

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WriterConfig configures the buffered ClickHouse writer.
type WriterConfig struct {
	DB             *gorm.DB
	Logger         *zap.Logger
	DisabledTables map[string]struct{}
	BufferSize     int
	FlushSize      int
	FlushInterval  time.Duration
}

// Writer buffers QueryRecords and flushes them to ClickHouse in batches.
type Writer struct {
	cfg       WriterConfig
	ch        chan CHQueryRecord
	processID string
	cancel    context.CancelFunc
	done      chan struct{}
}

// NewWriter creates a Writer. Call Start() to begin the background flush loop.
func NewWriter(cfg WriterConfig) *Writer {
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 4096
	}
	if cfg.FlushSize == 0 {
		cfg.FlushSize = 1000
	}
	if cfg.FlushInterval == 0 {
		cfg.FlushInterval = 5 * time.Second
	}
	if cfg.DisabledTables == nil {
		cfg.DisabledTables = make(map[string]struct{})
	}
	// Always disable our own table to prevent recursion.
	cfg.DisabledTables["queries"] = struct{}{}

	hostname, _ := os.Hostname()

	return &Writer{
		cfg:       cfg,
		ch:        make(chan CHQueryRecord, cfg.BufferSize),
		processID: hostname,
		done:      make(chan struct{}),
	}
}

// Write enqueues a record for persistence. Non-blocking; drops if the channel is full.
func (w *Writer) Write(r QueryRecord) {
	if _, disabled := w.cfg.DisabledTables[r.Table]; disabled {
		return
	}

	rec := CHQueryRecord{
		Table:        r.Table,
		Operation:    r.Operation,
		SQL:          r.SQL,
		DurationMS:   r.DurationMS,
		RowsAffected: r.RowsAffected,
		ResponseSize: r.ResponseSize,
		PreloadCount: r.PreloadCount,
		Timestamp:    r.Timestamp,
		Error:        r.Error,
		Caller:       r.Caller,
		CallerURL:    r.CallerURL,
		DBType:       r.DBType,
		Source:       r.Source,
		Endpoint:     r.Endpoint,
		ProcessID:    w.processID,
	}

	select {
	case w.ch <- rec:
	default:
	}
}

// Start begins the background flush goroutine. Call Stop() to shut down.
func (w *Writer) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	go w.loop(ctx)
}

// Stop signals the flush goroutine to drain and exit, then blocks until done.
func (w *Writer) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	<-w.done
}

func (w *Writer) loop(ctx context.Context) {
	defer close(w.done)

	ticker := time.NewTicker(w.cfg.FlushInterval)
	defer ticker.Stop()

	buf := make([]CHQueryRecord, 0, w.cfg.FlushSize)

	for {
		select {
		case rec := <-w.ch:
			buf = append(buf, rec)
			if len(buf) >= w.cfg.FlushSize {
				w.flush(buf)
				buf = buf[:0]
			}
		case <-ticker.C:
			if len(buf) > 0 {
				w.flush(buf)
				buf = buf[:0]
			}
		case <-ctx.Done():
			// Drain remaining records from the channel.
			for {
				select {
				case rec := <-w.ch:
					buf = append(buf, rec)
					if len(buf) >= w.cfg.FlushSize {
						w.flush(buf)
						buf = buf[:0]
					}
				default:
					if len(buf) > 0 {
						w.flush(buf)
					}
					return
				}
			}
		}
	}
}

func (w *Writer) flush(batch []CHQueryRecord) {
	if len(batch) == 0 {
		return
	}
	// Copy so the caller can reuse the slice.
	rows := make([]CHQueryRecord, len(batch))
	copy(rows, batch)

	if err := w.cfg.DB.Create(&rows).Error; err != nil {
		w.cfg.Logger.Warn("query collector: failed to flush to clickhouse",
			zap.Error(err),
			zap.Int("batch_size", len(rows)),
		)
	}
}
