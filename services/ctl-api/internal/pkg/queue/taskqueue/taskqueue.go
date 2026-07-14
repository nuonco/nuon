package taskqueue

import "github.com/nuonco/nuon/pkg/workflows"

// Queue names are duplicated as literals rather than imported from
// installs/helpers and runners/helpers: those packages import the queue client,
// which imports this package, so importing them back would create a cycle.
const (
	runnersNamespace  = "runners"
	installsNamespace = "installs"

	runnerHealthcheckCronsQueueName = "runner-healthcheck-crons"

	installActionCronSignalsQueueName = "install-action-cron-signals"
	installDriftCronSignalsQueueName  = "install-drift-cron-signals"
	installActionWorkflowsQueueName   = "install-action-workflows"
	installDriftWorkflowsQueueName    = "install-drift-workflows"
)

// For returns the Temporal task queue a Nuon queue's workflows run on, given the
// queue's namespace and name. Unrouted queues fall back to workflows.APITaskQueue.
func For(namespace, queueName string) string {
	switch namespace {
	case runnersNamespace:
		if queueName == runnerHealthcheckCronsQueueName {
			return workflows.RunnerHealthcheckCronsTaskQueue
		}
	case installsNamespace:
		switch queueName {
		case installActionCronSignalsQueueName,
			installDriftCronSignalsQueueName,
			installActionWorkflowsQueueName,
			installDriftWorkflowsQueueName:
			return workflows.InstallCronsTaskQueue
		}
	}

	return workflows.APITaskQueue
}
