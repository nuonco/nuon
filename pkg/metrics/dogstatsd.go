package metrics

import (
	"fmt"
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

func (w *writer) Flush() {
	if w.Disable {
		w.Log.Info("flush")
		return
	}
}

func (w *writer) Incr(name string, value int, tags []string) {
	if w.Disable {
		w.Log.Info(fmt.Sprintf("incr.%s", name))
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Incr(name, append(w.Tags, tags...), float64(value)))
}

func (w *writer) Decr(name string, value int, tags []string) {
	if w.Disable {
		w.Log.Info(fmt.Sprintf("decr.%s", name))
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Decr(name, append(w.Tags, tags...), float64(value)))
}

func (w *writer) Timing(name string, value time.Duration, tags []string) {
	if w.Disable {
		w.Log.Info(fmt.Sprintf("timing.%s", name))
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Timing(name, value, append(w.Tags, tags...), defaultRate))
}

func (w *writer) Event(ev *statsd.Event) {
	if w.Disable {
		w.Log.Info(fmt.Sprintf("event.%s", ev.Title))
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Event(ev))
}
