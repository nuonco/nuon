package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/joberrors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func newRunnerJobLifecycleFailures(provider metric.MeterProvider) metric.Int64Counter {
	if provider == nil {
		return nil
	}
	counter, _ := provider.Meter("github.com/nuonco/nuon/ctl-api/runner-job-lifecycle").Int64Counter(
		"nuon.runner.job.lifecycle.failures",
		metric.WithUnit("{failure}"),
		metric.WithDescription("Successful lifecycle failure recordings, including repeated activity invocations."),
	)
	return counter
}

func (a *Activities) recordRunnerJobLifecycleFailure(ctx context.Context, jobType app.RunnerJobType, reason joberrors.LifecycleFailureReason) {
	if a.jobFailures == nil {
		return
	}
	kind := string(jobType)
	if jobType.Group() == app.RunnerJobGroupUnknown {
		kind = "other"
	}
	errorType := "other"
	switch reason {
	case joberrors.LifecycleFailureReasonNoActiveRunner,
		joberrors.LifecycleFailureReasonQueueTimeout,
		joberrors.LifecycleFailureReasonRunnerUnhealthy,
		joberrors.LifecycleFailureReasonPickupTimeout,
		joberrors.LifecycleFailureReasonOverallTimeout,
		joberrors.LifecycleFailureReasonExecutionTimeout,
		joberrors.LifecycleFailureReasonAttemptsExhausted,
		joberrors.LifecycleFailureReasonResultMissing,
		joberrors.LifecycleFailureReasonRunnerDisabled:
		errorType = string(reason)
	}
	a.jobFailures.Add(ctx, 1, metric.WithAttributes(
		attribute.String("nuon.runner.job.type", kind),
		attribute.String("error.type", errorType),
	))
}
