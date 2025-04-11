package worker

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	tworker "go.temporal.io/sdk/worker"
)

func (w *worker) getWorker(c client.Client) (tworker.Worker, error) {
	return tworker.New(c, w.Config.TemporalTaskQueue, tworker.Options{
		MaxConcurrentActivityExecutionSize: w.Config.TemporalMaxConcurrentActivities,
		Interceptors:                       []interceptor.WorkerInterceptor{},
		WorkflowPanicPolicy:                tworker.FailWorkflow,
	}), nil
}

func (w *worker) registerWorkflowsOnWorker(wkr tworker.Worker) {
	for _, wkflow := range w.Workflows {
		wkr.RegisterWorkflow(wkflow)
	}
}

func (w *worker) registerActivitiesOnWorker(wkr tworker.Worker) {
	for _, act := range w.Activities {
		wkr.RegisterActivity(act)
	}
}
