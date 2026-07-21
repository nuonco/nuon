package taskqueue

import "github.com/nuonco/nuon/pkg/workflows"

// For returns the Temporal task queue a Nuon queue's workflows run on, given the
// queue's namespace. Cron work lives in dedicated namespaces, each polled on its
// own task queue; everything else falls back to workflows.APITaskQueue.
func For(namespace, queueName string) string {
	switch namespace {
	case workflows.RunnerHealthcheckCronsNamespace:
		return workflows.RunnerHealthcheckCronsTaskQueue
	case workflows.InstallCronsNamespace:
		return workflows.InstallCronsTaskQueue
	}

	return workflows.APITaskQueue
}
