package helpers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/joberrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestResolveJobCompositeError(t *testing.T) {
	executionError := &compositeerrors.CompositeErrorData{Type: "terraform.error"}
	lifecycleError := &compositeerrors.CompositeErrorData{Type: "runner.job_lifecycle_failure"}
	cancellationError := &compositeerrors.CompositeErrorData{Type: joberrors.CancellationErrorType}
	policyWarning := &compositeerrors.CompositeErrorData{Type: "policy.evaluation_failed", Severity: compositeerrors.SeverityWarning}

	tests := map[string]struct {
		job      app.RunnerJob
		expected *compositeerrors.CompositeErrorData
	}{
		"latest execution result takes precedence": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusFailed,
				CompositeError: lifecycleError,
				Executions: []app.RunnerJobExecution{{
					Result: &app.RunnerJobExecutionResult{CompositeError: executionError},
				}},
			},
			expected: executionError,
		},
		"failed job uses lifecycle error": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusFailed,
				CompositeError: lifecycleError,
			},
			expected: lifecycleError,
		},
		"cancelled job uses cancellation error instead of execution error": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusCancelled,
				CompositeError: cancellationError,
				Executions: []app.RunnerJobExecution{{
					Result: &app.RunnerJobExecutionResult{CompositeError: executionError},
				}},
			},
			expected: cancellationError,
		},
		"cancelled job hides stale lifecycle error": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusCancelled,
				CompositeError: lifecycleError,
			},
		},
		"not attempted job uses lifecycle error": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusNotAttempted,
				CompositeError: lifecycleError,
			},
			expected: lifecycleError,
		},
		"finished job hides stale lifecycle error": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusFinished,
				CompositeError: lifecycleError,
			},
		},
		"finished job exposes policy warning": {
			job: app.RunnerJob{
				Status:         app.RunnerJobStatusFinished,
				CompositeError: policyWarning,
			},
			expected: policyWarning,
		},
		"finished job hides stale execution error": {
			job: app.RunnerJob{
				Status: app.RunnerJobStatusFinished,
				Executions: []app.RunnerJobExecution{{
					Result: &app.RunnerJobExecutionResult{CompositeError: executionError},
				}},
			},
		},
		"retrying job hides previous execution error": {
			job: app.RunnerJob{
				Status: app.RunnerJobStatusAvailable,
				Executions: []app.RunnerJobExecution{{
					Result: &app.RunnerJobExecutionResult{CompositeError: executionError},
				}},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Same(t, tt.expected, ResolveJobCompositeError(&tt.job))
		})
	}
}
