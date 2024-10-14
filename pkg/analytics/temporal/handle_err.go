package temporalanalytics

import (
	segment "github.com/segmentio/analytics-go/v3"
	"go.uber.org/zap"
)

func (w *writer) handleErr(typ string, err error) {
	w.Logger.Error("error recording event", zap.String("type", typ), zap.Error(err))
}

func (w *writer) toProperties(props map[string]interface{}) segment.Properties {
	segmentProperties := segment.NewProperties()
	for k, v := range w.Properties {
		segmentProperties.Set(k, v)
	}

	for k, v := range props {
		segmentProperties.Set(k, v)
	}

	return segmentProperties
}
