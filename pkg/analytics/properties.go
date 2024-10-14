package analytics

import (
	segment "github.com/segmentio/analytics-go/v3"
)

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
