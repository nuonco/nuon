package examples

import "go.temporal.io/sdk/workflow"

type Workflows struct {
	cfg string
	v   int
}

type QueueWorkflowRequest struct {
	QueueID string
}

// @temporal-gen-v2 workflow
// @task-queue "queue"
// @id-template queue-{{.QueueID}}
func (w *Workflows) Queue(ctx workflow.Context, req QueueWorkflowRequest) error {
	return nil
}
