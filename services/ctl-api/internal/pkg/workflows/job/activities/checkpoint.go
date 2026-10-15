package activities

import (
	"context"
	"encoding/json"

	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (a *Activities) LoadRunnerJobCheckpoint(ctx context.Context, signalID, phase string) (result []byte, err error) {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("signal_id", signalID), zap.String("phase", phase))
	l.Debug("loading runner job checkpoint")
	defer func() {
		if err != nil {
			l.Error("unable to load runner job checkpoint", zap.Error(err))
		}
	}()
	var qs app.QueueSignal
	if err := a.db.WithContext(ctx).Select("id", "continuation").Where(app.QueueSignal{ID: signalID}).First(&qs).Error; err != nil {
		return nil, err
	}
	data := make(map[string]json.RawMessage)
	if len(qs.Continuation) > 0 {
		if err := json.Unmarshal(qs.Continuation, &data); err != nil {
			return nil, err
		}
	}
	return data[phase], nil
}

func (a *Activities) SaveRunnerJobCheckpoint(ctx context.Context, signalID, phase string, state []byte) (err error) {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("signal_id", signalID), zap.String("phase", phase))
	l.Debug("saving runner job checkpoint")
	defer func() {
		if err != nil {
			l.Error("unable to save runner job checkpoint", zap.Error(err))
		}
	}()
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var qs app.QueueSignal
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.QueueSignal{ID: signalID}).First(&qs).Error; err != nil {
			return err
		}
		if qs.Status.Status == app.StatusCancelled {
			return temporal.NewNonRetryableApplicationError("signal is cancelled", "CANCELLED", nil)
		}
		data := make(map[string]json.RawMessage)
		if len(qs.Continuation) > 0 {
			if err := json.Unmarshal(qs.Continuation, &data); err != nil {
				return err
			}
		}
		data[phase] = state
		encoded, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return tx.Model(&qs).Update("continuation", encoded).Error
	})
}
