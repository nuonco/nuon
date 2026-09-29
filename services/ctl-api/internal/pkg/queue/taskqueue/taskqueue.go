package taskqueue

import "github.com/nuonco/nuon/pkg/workflows"

func For(namespace, queueName string) string {
	switch namespace {
	case workflows.RunnerHealthcheckCronsNamespace:
		return workflows.RunnerHealthcheckCronsTaskQueue
	case workflows.InstallCronsNamespace:
		return workflows.InstallCronsTaskQueue
	}

	return workflows.APITaskQueue
}
