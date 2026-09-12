package flow

import (
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
)

// StepConfig holds the queue/owner configuration needed to dispatch and manage
// workflow steps.
type StepConfig struct {
	GroupQueueName         string
	QueueName              string
	TargetQueueName        string
	GenerateStepsQueueName string
	OwnerID                string
	OwnerType              string
	MW                     tmetrics.Writer
}
