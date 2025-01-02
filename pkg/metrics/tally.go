package metrics

import (
	"io"
	"time"

	"github.com/uber-go/tally/v4"
)

const (
	defaultSamplingInterval time.Duration = time.Second * 10
)

// NewTallyScope creates a new tally scope with the given metrics.Writer as the backend,
// by way of the TallyReporter wrapper.
//
// Tags to be persisted on all metrics emitted via the returned [tally.Scope] should be attached to the provided [Writer].
//
// The returned io.Closer should be closed when no more telemetry is to be written to the scope.
func NewTallyScope(mw Writer) (tally.Scope, io.Closer) {
	return tally.NewRootScope(tally.ScopeOptions{
		Reporter: &TallyReporter{mw: mw},
	}, defaultSamplingInterval)
}

// TallyReporter is a wrapper around the metrics.Writer interface that conforms to [tally.StatsReporter],
// allowing us to use the same metrics.Writer implementation for temporal's client, which expects a [tally.Scope]
// type TallyReporter metrics.Writer
type TallyReporter struct {
	mw Writer
}

var _ tally.StatsReporter = (*TallyReporter)(nil)

type cap bool

func (c cap) Reporting() bool { return bool(c) }
func (c cap) Tagging() bool   { return bool(c) }

// Capabilities implements tally.StatsReporter.
func (w *TallyReporter) Capabilities() tally.Capabilities {
	return cap(true)
}

// Flush implements tally.StatsReporter.
func (w *TallyReporter) Flush() {
	w.mw.Flush()
}

// ReportGauge implements tally.StatsReporter.
func (w *TallyReporter) ReportGauge(name string, tags map[string]string, value float64) {
	stags := make([]string, 0, len(tags))
	for k, v := range tags {
		stags = append(stags, k+":"+v)
	}
	w.mw.Gauge(name, value, stags)
}

// ReportHistogramDurationSamples implements tally.StatsReporter as a no-op.
//
// Tally's interface assumes histograms that are calculated/aggregated client-side. Datadog wants to aggregate histograms in their agent.
// We could sorta-approximate this by breaking down the input histogram as a series of gauges, but the agent would then smash the data points
// in the time range together, significantly undermining the fidelity of histogram data.
//
// So, rather than produce potentially bad data that increases noise in our obs platform, we err on the side of just dropping this data.
// We can always revisit if there's a compelling need for some metrics expressed as histograms that outweights the risk of noise.
func (w *TallyReporter) ReportHistogramDurationSamples(name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound time.Duration, bucketUpperBound time.Duration, samples int64) {
}

// ReportHistogramValueSamples implements tally.StatsReporter as a no-op.
//
// See ReportHistogramDurationSamples for no-op rationale.
func (w *TallyReporter) ReportHistogramValueSamples(name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound float64, bucketUpperBound float64, samples int64) {
}

// ReportTimer implements tally.StatsReporter.
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
