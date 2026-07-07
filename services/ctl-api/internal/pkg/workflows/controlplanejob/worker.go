package controlplanejob

import (
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
)

// WorkerConfig configures a control-plane build worker.
type WorkerConfig struct {
	MaxConcurrentActivityExecutionSize int
	MaxConcurrentActivityTaskPollers   int
	Interceptors                       []interceptor.WorkerInterceptor
	WorkflowPanicPolicy                worker.WorkflowPanicPolicy
}

// NewWorker creates a worker that polls the control-plane build task queue on
// the given namespace client and registers the provided activities.
//
// The ExecuteControlPlaneJob workflow runs as a child of its caller, so it
// inherits the caller's namespace and its RunJob activity is pinned to the
// control-plane build task queue in that same namespace (see workflow.go).
// Every namespace that starts control-plane build child workflows (components
// for component builds, apps for sandbox builds) must therefore run one of
// these workers, otherwise RunJob is scheduled on a task queue with no poller
// and the build hangs silently until its execution timeout.
func NewWorker(c tclient.Client, cfg WorkerConfig, activities ...any) worker.Worker {
	wkr := worker.New(c, TaskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize: cfg.MaxConcurrentActivityExecutionSize,
		MaxConcurrentActivityTaskPollers:   cfg.MaxConcurrentActivityTaskPollers,
		Interceptors:                       cfg.Interceptors,
		WorkflowPanicPolicy:                cfg.WorkflowPanicPolicy,
	})
	for _, act := range activities {
		wkr.RegisterActivity(act)
	}
	return wkr
}
