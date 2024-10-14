package temporalanalytics

import (
	"github.com/pkg/errors"
	segment "github.com/segmentio/analytics-go/v3"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/analytics/events"
)

func (w *writer) Track(ctx workflow.Context, ev events.Event, props map[string]interface{}) {
	userID, err := w.UserIDFn(ctx)
	if err != nil {
		w.handleErr("track", errors.Wrap(err, "unable to get user id"))
		return
	}

	if w.Disable {
		w.Logger.Debug("tracking event", zap.String("event", string(ev)))
		return
	}

	segProps := w.toProperties(props)
	if err := w.client.Enqueue(segment.Track{
		UserId:     userID,
		Event:      string(ev),
		Properties: segProps,
	}); err != nil {
		w.handleErr("track", errors.Wrap(err, "unable to emit event"))
	}
}
