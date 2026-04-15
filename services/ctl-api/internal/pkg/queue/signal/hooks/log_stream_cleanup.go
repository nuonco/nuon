package hooks

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// logStreamMetadataKey is the metadata key used by the handler to pass
// log stream IDs that should be closed after signal execution.
const logStreamMetadataKey = "log_stream_id"

type LogStreamCleanupHook struct {
	l  *zap.Logger
	db *gorm.DB
}

var _ signal.SignalLifecycleHook = (*LogStreamCleanupHook)(nil)

func NewLogStreamCleanupHook(params Params) *LogStreamCleanupHook {
	logger := params.L
	if logger == nil {
		logger = zap.NewNop()
	}

	return &LogStreamCleanupHook{
		l:  logger,
		db: params.DB,
	}
}

func (h *LogStreamCleanupHook) Name() string {
	return "log_stream_cleanup"
}

func (h *LogStreamCleanupHook) Supports(event signal.SignalPhaseEvent) bool {
	if h.db == nil {
		return false
	}

	return event.Phase == signal.SignalPhaseExecute || event.Phase == signal.SignalPhaseCancel
}

func (h *LogStreamCleanupHook) PreExecute(_ context.Context, _ signal.SignalPhaseEvent) (signal.PreExecuteDecision, error) {
	return signal.AllowDecision(), nil
}

func (h *LogStreamCleanupHook) PostExecute(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	logStreamID, ok := outcome.Metadata[logStreamMetadataKey].(string)
	if !ok || logStreamID == "" {
		return nil
	}

	l := h.l.With(
		zap.String("hook", h.Name()),
		zap.String("queue_signal_id", event.QueueSignalID),
		zap.String("log_stream_id", logStreamID),
		zap.String("status", string(outcome.Status)),
	)

	ls := &app.LogStream{ID: logStreamID}
	res := h.db.WithContext(ctx).
		Model(ls).
		Updates(map[string]interface{}{
			"open": false,
		})
	if res.Error != nil {
		l.Error("failed to close log stream", zap.Error(res.Error))
		return nil
	}

	l.Info("closed log stream via lifecycle hook")
	return nil
}
