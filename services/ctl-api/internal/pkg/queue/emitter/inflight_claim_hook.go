package emitter

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

type InFlightClaimHookParams struct {
	fx.In

	DB *gorm.DB
	L  *zap.Logger
}

type InFlightClaimHook struct {
	db *gorm.DB
	l  *zap.Logger
}

var _ signal.SignalLifecycleHook = (*InFlightClaimHook)(nil)

func NewInFlightClaimHook(params InFlightClaimHookParams) *InFlightClaimHook {
	return &InFlightClaimHook{db: params.DB, l: params.L}
}

func (h *InFlightClaimHook) Name() string {
	return "emitter-inflight-claim"
}

func (h *InFlightClaimHook) Supports(_ signal.SignalPhaseEvent) bool {
	return true
}

func (h *InFlightClaimHook) BeforePhase(_ context.Context, _ signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	return signal.AllowPhaseDecision(), nil
}

func (h *InFlightClaimHook) AfterPhase(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	switch outcome.Status {
	case signal.SignalStatusSuccess, signal.SignalStatusError, signal.SignalStatusCancelled:
	default:
		return nil
	}

	var qs app.QueueSignal
	if err := h.db.WithContext(ctx).
		Select("emitter_id").
		First(&qs, "id = ?", event.QueueSignalID).Error; err != nil {
		h.l.Warn("unable to load queue signal to clear emitter claim",
			zap.String("queue_signal_id", event.QueueSignalID),
			zap.Error(err))
		return nil
	}
	if qs.EmitterID == nil {
		return nil
	}

	res := h.db.WithContext(ctx).
		Model(&app.QueueEmitter{}).
		Where("id = ?", *qs.EmitterID).
		Update("in_flight_claimed_at", nil)
	if res.Error != nil {
		h.l.Warn("unable to clear emitter in-flight claim",
			zap.String("emitter_id", *qs.EmitterID),
			zap.Error(res.Error))
	}
	return nil
}
