package worker

import (
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	tworker "go.temporal.io/sdk/worker"
)

func (w *worker) getWorker(c client.Client) (tworker.Worker, error) {
	t, err := w.getTracer()
	if err != nil {
		return nil, fmt.Errorf("unable to get tracer: %w", err)
	}

	traceIntercepter, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
		Tracer: t,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get tracing interceptor: %w", err)
	}

	return tworker.New(c, w.Config.TemporalTaskQueue, tworker.Options{
		MaxConcurrentActivityExecutionSize: w.Config.TemporalMaxConcurrentActivities,
		Interceptors:                       []interceptor.WorkerInterceptor{traceIntercepter},
	}), nil
}

func (w *worker) registerWorker(wkr tworker.Worker) {
	for _, wkflow := range w.Workflows {
		wkr.RegisterWorkflow(wkflow)
	}

	for _, act := range w.Activities {
		wkr.RegisterActivity(act)
	}
}
