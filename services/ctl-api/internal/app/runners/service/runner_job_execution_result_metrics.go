package service

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func newRunnerJobExecutionResults(provider metric.MeterProvider) metric.Int64Counter {
	if provider == nil {
		return nil
	}
	counter, _ := provider.Meter("github.com/nuonco/nuon/ctl-api/runner-job-execution").Int64Counter(
		"nuon.runner.job.execution.results",
		metric.WithUnit("{result}"),
		metric.WithDescription("Newly persisted runner-reported execution results by reported outcome."),
	)
	return counter
}

func (s *service) recordRunnerJobExecutionResult(ctx context.Context, job *app.RunnerJob, result *app.RunnerJobExecutionResult) {
	if s.executionResults == nil {
		return
	}
	jobType := string(job.Type)
	if job.Type.Group() == app.RunnerJobGroupUnknown {
		jobType = "other"
	}
	operation := "other"
	switch job.Operation {
	case app.RunnerJobOperationTypeExec,
		app.RunnerJobOperationTypeBuild,
		app.RunnerJobOperationTypeCreateApplyPlan,
		app.RunnerJobOperationTypeCreateTeardownPlan,
		app.RunnerJobOperationTypeApplyPlan:
		operation = string(job.Operation)
	}
	outcome := "failure"
	if result.Success {
		outcome = "success"
	}
	attrs := []attribute.KeyValue{
		attribute.String("nuon.runner.job.type", jobType),
		attribute.String("nuon.runner.job.operation", operation),
		attribute.String("outcome", outcome),
	}
	if installID := job.FlowInstallID(); installID != "" {
		attrs = append(attrs, attribute.String("nuon.install.id", installID))
	}
	s.executionResults.Add(ctx, 1, metric.WithAttributes(attrs...))
}
