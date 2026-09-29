package metrics

import (
	"io"
	"time"

	"github.com/uber-go/tally/v4"
)

const (
	defaultSamplingInterval time.Duration = time.Second * 10
)

func NewTallyScope(mw Writer) (tally.Scope, io.Closer) {
	return tally.NewRootScope(tally.ScopeOptions{
		Reporter: &TallyReporter{mw: mw},
	}, defaultSamplingInterval)
}

type TallyReporter struct {
	mw Writer
}

var _ tally.StatsReporter = (*TallyReporter)(nil)

type cap bool

func (c cap) Reporting() bool { return bool(c) }
func (c cap) Tagging() bool   { return bool(c) }

func (w *TallyReporter) Capabilities() tally.Capabilities {
	return cap(true)
}

func (w *TallyReporter) Flush() {
	w.mw.Flush()
}

func (w *TallyReporter) ReportGauge(name string, tags map[string]string, value float64) {
	stags := make([]string, 0, len(tags))
	for k, v := range tags {
		stags = append(stags, k+":"+v)
	}
	w.mw.Gauge(name, value, stags)
}

func (w *TallyReporter) ReportHistogramDurationSamples(name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound time.Duration, bucketUpperBound time.Duration, samples int64) {
}

func (w *TallyReporter) ReportHistogramValueSamples(name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound float64, bucketUpperBound float64, samples int64) {
}

func (w *TallyReporter) ReportTimer(name string, tags map[string]string, interval time.Duration) {
	stags := make([]string, 0, len(tags))
	for k, v := range tags {
		stags = append(stags, k+":"+v)
	}
	w.mw.Timing(name, interval, stags)
}

func (w *TallyReporter) ReportCounter(name string, tags map[string]string, value int64) {
	stags := make([]string, 0, len(tags))
	for k, v := range tags {
		stags = append(stags, k+":"+v)
	}
	w.mw.Count(name, value, stags)
}
