package analytics

import (
	"context"

	"github.com/pkg/errors"
	segment "github.com/segmentio/analytics-go/v3"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/analytics/events"
)

func (w *writer) Identify(ctx context.Context) {
	ident, err := w.IdentifyFn(ctx)
	if err != nil {
		w.handleErr("identify", errors.Wrap(err, "unable to get identity"))
		return
	}

	if w.Disable {
		w.Logger.Debug("skipping identify")
		return
	}

	if err := w.client.Enqueue(ident); err != nil {
		w.handleErr("send identify", errors.Wrap(err, "unable to send identify"))
	}
}

func (w *writer) Group(ctx context.Context) {
	grp, err := w.GroupFn(ctx)
	if err != nil {
		w.handleErr("group", errors.Wrap(err, "unable to get group using fn"))
		return
	}

	if w.Disable {
		w.Logger.Debug("skipping group")
		return
	}

	if err := w.client.Enqueue(grp); err != nil {
		w.handleErr("send group", errors.Wrap(err, "unable send group"))
	}
}

func (w *writer) Track(ctx context.Context, ev events.Event, props map[string]interface{}) {
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
