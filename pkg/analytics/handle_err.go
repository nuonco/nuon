package analytics

import "go.uber.org/zap"

func (w *writer) handleErr(typ string, err error) {
	w.Logger.Error("error recording event", zap.String("type", typ), zap.Error(err))
}
