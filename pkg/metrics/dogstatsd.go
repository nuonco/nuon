package metrics

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
		w.Log.Debug("flush")
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}
	w.handleErr(client.Flush())
}

func (w *writer) tagsToZapFields(tags []string) []zapcore.Field {
	fields := make([]zapcore.Field, 0)
	for _, t := range tags {
		k, v := SplitTag(t)
		fields = append(fields, zap.String(k, v))
	}

	return fields
}

func (w *writer) Incr(name string, tags []string) {
	if w.Disable {
		w.Log.Debug(fmt.Sprintf("incr.%s", name), w.tagsToZapFields(tags)...)
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Incr(name, append(w.Tags, tags...), defaultRate))
}

func (w *writer) Decr(name string, tags []string) {
	if w.Disable {
		w.Log.Debug(fmt.Sprintf("decr.%s", name), w.tagsToZapFields(tags)...)
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Decr(name, append(w.Tags, tags...), defaultRate))
}

func (w *writer) Count(name string, value int64, tags []string) {
	if w.Disable {
		allTags := w.tagsToZapFields(tags)
		allTags = append(allTags, zap.Int64("value", value))
		w.Log.Debug(fmt.Sprintf("count.%s", name), allTags...)
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Count(name, value, append(w.Tags, tags...), defaultRate))
}

func (w *writer) Gauge(name string, value float64, tags []string) {
	if w.Disable {
		allTags := w.tagsToZapFields(tags)
		allTags = append(allTags, zap.Float64("value", value))
		w.Log.Debug(fmt.Sprintf("gauge.%s", name), allTags...)
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Gauge(name, float64(value), append(w.Tags, tags...), defaultRate))
}

func (w *writer) Timing(name string, value time.Duration, tags []string) {
	if w.Disable {
		allTags := w.tagsToZapFields(tags)
		allTags = append(allTags, zap.String("duration", value.String()))
		w.Log.Debug(fmt.Sprintf("timing.%s", name), allTags...)
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
		allTags := w.tagsToZapFields(ev.Tags)
		w.Log.Debug(fmt.Sprintf("event.%s (agg key: %s): %s", ev.Title, ev.AggregationKey, ev.Text, ), allTags...)
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Event(ev))
}
