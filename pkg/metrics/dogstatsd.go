package metrics

import (
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"go.uber.org/zap"
)

const (
	// defaultRate is used to control the sampling rate, by default we send everything.
	defaultRate float64 = 1.0
)

func (w *writer) handleErr(err error) {
	if err == nil {
		return
	}

	w.Log.Error("unable to write", zap.String("addr", w.Address))
}

func (w *writer) Incr(name string, value int) {
	if w.Disable {
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Incr(name, w.Tags, float64(value)))
}

func (w *writer) Decr(name string, value int) {
	if w.Disable {
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Decr(name, w.Tags, float64(value)))
}

func (w *writer) Timing(name string, value time.Duration) {
	if w.Disable {
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Timing(name, value, w.Tags, defaultRate))
}

func (w *writer) Event(ev *statsd.Event) {
	if w.Disable {
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Event(ev))
}
