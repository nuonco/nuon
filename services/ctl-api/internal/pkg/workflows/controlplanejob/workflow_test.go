package controlplanejob

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestExecuteControlPlaneJobDoesNotRetryJobFailure(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	env.OnActivity((*Activities).ControlPlaneJobEnsureExecution, mock.Anything, mock.Anything, mock.Anything).
		Return(&EnsureExecutionResponse{ExecutionID: "execution-id", JobExecutionTimeout: time.Minute}, nil).
		Once()
	env.OnActivity((*Activities).ControlPlaneJobRunJob, mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("image not found")).
		Once()
	env.OnActivity((*Activities).ControlPlaneJobFinalize, mock.Anything, mock.Anything, mock.Anything).
		Return(&FinalizeResponse{Status: app.RunnerJobExecutionStatusFailed}, nil).
		Once()

	env.ExecuteWorkflow((&Workflows{}).ExecuteControlPlaneJob, &ExecuteRequest{JobID: "job-id"})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
	env.AssertExpectations(t)
}

func TestExecuteControlPlaneJobFailureDiagnostics(t *testing.T) {
	for name, tc := range map[string]struct {
		runError      error
		finalizeError error
		status        app.RunnerJobExecutionStatus
		message       string
		want          []string
		cancelled     bool
		cancelContext bool
	}{
		"new cancelled context and failed finalization": {
			cancelContext: true, cancelled: true,
			finalizeError: temporal.NewNonRetryableApplicationError("database unavailable", "database", nil),
		},
		"ECR failure and failed finalization": {
			runError:      errors.New("InvalidIdentityToken: No OpenIDConnect provider found"),
			finalizeError: temporal.NewNonRetryableApplicationError("database unavailable", "database", nil),
			want:          []string{"InvalidIdentityToken", "database unavailable"},
		},
		"GAR failure and failed finalization": {
			runError:      errors.New("PERMISSION_DENIED: artifactregistry.repositories.downloadArtifacts denied"),
			finalizeError: temporal.NewNonRetryableApplicationError("database unavailable", "database", nil),
			want:          []string{"artifactregistry.repositories.downloadArtifacts", "database unavailable"},
		},
		"ACR failure and failed finalization": {
			runError:      errors.New("UNAUTHORIZED: pull access to acme/api denied"),
			finalizeError: temporal.NewNonRetryableApplicationError("database unavailable", "database", nil),
			want:          []string{"pull access to acme/api denied", "database unavailable"},
		},
		"successful run and failed finalization": {
			finalizeError: temporal.NewNonRetryableApplicationError("database unavailable", "database", nil),
			want:          []string{"unable to finalize", "database unavailable"},
		},
		"missing tag": {
			runError: errors.New("MANIFEST_UNKNOWN: tag v9.7.3 does not exist"),
			status:   app.RunnerJobExecutionStatusFailed,
			want:     []string{"MANIFEST_UNKNOWN", "v9.7.3"},
		},
		"structured result takes precedence": {
			runError: errors.New("unable to copy image"),
			status:   app.RunnerJobExecutionStatusFailed,
			message:  "Image registry rejected access",
			want:     []string{"Image registry rejected access"},
		},
		"cancellation survives failed finalization": {
			runError:      temporal.NewCanceledError(),
			finalizeError: temporal.NewNonRetryableApplicationError("database unavailable", "database", nil),
			cancelled:     true,
		},
		"success": {status: app.RunnerJobExecutionStatusFinished},
	} {
		t.Run(name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			if tc.cancelContext {
				env.RegisterDelayedCallback(env.CancelWorkflow, 0)
			}
			env.OnActivity((&Activities{}).ControlPlaneJobEnsureExecution, mock.Anything, mock.Anything).
				Return(&EnsureExecutionResponse{ExecutionID: "execution-id", JobExecutionTimeout: time.Minute}, nil).Once()
			run := env.OnActivity((&Activities{}).ControlPlaneJobRunJob, mock.Anything, mock.Anything).
				Return(tc.runError)
			if tc.cancelContext {
				run.Maybe()
			} else {
				run.Once()
			}
			env.OnActivity((&Activities{}).ControlPlaneJobFinalize, mock.Anything, mock.Anything).
				Return(&FinalizeResponse{Status: tc.status, Message: tc.message}, tc.finalizeError).Once()

			env.ExecuteWorkflow((&Workflows{}).ExecuteControlPlaneJob, &ExecuteRequest{JobID: "job-id"})
			err := env.GetWorkflowError()
			if tc.cancelled {
				require.True(t, temporal.IsCanceledError(err), "%v", err)
			} else if len(tc.want) == 0 {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.False(t, temporal.IsCanceledError(err), "%v", err)
				for _, want := range tc.want {
					require.ErrorContains(t, err, want)
				}
				if tc.runError != nil && tc.finalizeError != nil {
					var applicationErr *temporal.ApplicationError
					require.True(t, errors.As(err, &applicationErr))
					require.Equal(t, "control-plane-build", applicationErr.Type())
					require.True(t, applicationErr.NonRetryable())
				}
			}
			env.AssertExpectations(t)
		})
	}
}
