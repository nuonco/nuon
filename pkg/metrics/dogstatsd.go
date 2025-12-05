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

	msg := fmt.Sprintf("unable to write metric. error: %s", err.Error())
	w.Log.Error(msg, zap.String("addr", w.Address))
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
	allTags := append(w.Tags, tags...)

	if w.Disable {
		allTags = append(allTags, "metric_type:incr")
		w.Log.Debug(
			fmt.Sprintf("metric.incr.%s", name),
			w.tagsToZapFields(allTags)...)
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Incr(name, allTags, defaultRate))
}

func (w *writer) Decr(name string, tags []string) {
	allTags := append(w.Tags, tags...)
	if w.Disable {
		l := w.Log.With(
			zap.String("metric_type", "decr"),
			zap.String("is_metric", "true"))
		l.Debug(
			fmt.Sprintf("metric.decr.%s", name),
			w.tagsToZapFields(allTags)...)
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Decr(name, allTags, defaultRate))
}

func (w *writer) Count(name string, value int64, tags []string) {
	allTags := append(w.Tags, tags...)
	if w.Disable {
		l := w.Log.With(
			zap.String("metric_type", "count"),
			zap.String("is_metric", "true"))

		allAttrs := w.tagsToZapFields(allTags)
		allAttrs = append(allAttrs, zap.Int64("value", value))
		l.Debug(fmt.Sprintf("metric.count.%s", name), allAttrs...)
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
	allTags := append(w.Tags, tags...)
	if w.Disable {
		l := w.Log.With(
			zap.String("metric_type", "gauge"),
			zap.String("is_metric", "true"))
		allAttrs := w.tagsToZapFields(allTags)
		allAttrs = append(allAttrs, zap.Float64("value", value))
		l.Debug(fmt.Sprintf("metric.gauge.%s", name), allAttrs...)
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Gauge(name, float64(value), allTags, defaultRate))
}

func (w *writer) Timing(name string, value time.Duration, tags []string) {
	allTags := append(w.Tags, tags...)
	if w.Disable {
		l := w.Log.With(zap.String("metric_type", "timing"), zap.String("is_metric", "true"))
		allAttrs := w.tagsToZapFields(allTags)
		allAttrs = append(allAttrs, zap.String("duration", value.String()))
		l.Debug(fmt.Sprintf("metric.timing.%s", name), allAttrs...)
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Timing(name, value, allTags, defaultRate))
}

func (w *writer) Event(ev *statsd.Event) {
	if w.Disable {
		allTags := w.tagsToZapFields(ev.Tags)
		w.Log.Debug(fmt.Sprintf("event.%s (agg key: %s): %s", ev.Title, ev.AggregationKey, ev.Text), allTags...)
		return
	}

	client, err := w.getClient()
	if err != nil {
		w.handleErr(err)
		return
	}

	w.handleErr(client.Event(ev))
}
