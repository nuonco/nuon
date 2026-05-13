package emitter

import (
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/metrics"
)

const lifecycleSourceTypeName = "nuon-queue-emitter"

// Stable reason tags emitted with queue-emitter-stopped events. Keep the
// set small so dashboards can pivot on them without growing cardinality.
const (
	stopReasonEmitterNotFound      = "emitter_not_found"
	stopReasonQueueTerminated      = "queue_terminated"
	stopReasonStopSignal           = "stop_signal_received"
	stopReasonScheduledComplete    = "scheduled_complete"
	stopReasonScheduledAlreadyDone = "scheduled_already_fired"
	stopReasonWorkflowError        = "workflow_error"
	stopReasonUnknown              = "unknown"
)

func (e *emitterWorkflow) lifecycleTags(extras map[string]string) []string {
	tags := map[string]string{
		"emitter_id": e.emitterID,
		"queue_id":   e.queueID,
	}
	for k, v := range extras {
		tags[k] = v
	}
	return metrics.ToTags(tags)
}

// emitStartedEvent fires on every workflow invocation. Not gated by
// continue-as-new: each CaN restart re-fires the event with
// continued_as_new=true so we can separate fresh starts from restarts.
func (e *emitterWorkflow) emitStartedEvent(ctx workflow.Context, continuedAsNew bool) {
	tags := e.lifecycleTags(map[string]string{
		"continued_as_new": fmt.Sprintf("%t", continuedAsNew),
	})
	e.mw.Event(ctx, &statsd.Event{
		Title:          "queue emitter started",
		Text:           "Emitter workflow entered run loop",
		Tags:           tags,
		SourceTypeName: lifecycleSourceTypeName,
		Priority:       statsd.Low,
		AlertType:      statsd.Info,
		AggregationKey: "queue-emitter-started",
	})
}

// emitStoppedEvent fires only on terminal workflow exit. continue-as-new
// is not a stop and must not emit. The reason tag answers "why did it
// stop entirely" — unexpected reasons are surfaced as warnings.
func (e *emitterWorkflow) emitStoppedEvent(ctx workflow.Context, reason string) {
	alertType := statsd.Info
	switch reason {
	case stopReasonEmitterNotFound, stopReasonQueueTerminated, stopReasonWorkflowError:
		alertType = statsd.Warning
	}
	tags := e.lifecycleTags(map[string]string{
		"reason": reason,
	})
	e.mw.Event(ctx, &statsd.Event{
		Title:          fmt.Sprintf("queue emitter stopped: %s", reason),
		Text:           fmt.Sprintf("Emitter workflow exited. reason=%s", reason),
		Tags:           tags,
		SourceTypeName: lifecycleSourceTypeName,
		Priority:       statsd.Normal,
		AlertType:      alertType,
		AggregationKey: "queue-emitter-stopped",
	})
}

func (e *emitterWorkflow) emitPausedEvent(ctx workflow.Context) {
	e.mw.Event(ctx, &statsd.Event{
		Title:          "queue emitter paused",
		Text:           "Pause signal received; emitter status set to cancelled",
		Tags:           e.lifecycleTags(nil),
		SourceTypeName: lifecycleSourceTypeName,
		Priority:       statsd.Normal,
		AlertType:      statsd.Info,
		AggregationKey: "queue-emitter-paused",
	})
}

func (e *emitterWorkflow) emitResumedEvent(ctx workflow.Context) {
	e.mw.Event(ctx, &statsd.Event{
		Title:          "queue emitter resumed",
		Text:           "Resume signal received; emitter status set to in_progress",
		Tags:           e.lifecycleTags(nil),
		SourceTypeName: lifecycleSourceTypeName,
		Priority:       statsd.Normal,
		AlertType:      statsd.Info,
		AggregationKey: "queue-emitter-resumed",
	})
}
