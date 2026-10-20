package actions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestActionRunFailureStatus(t *testing.T) {
	tests := map[string]struct {
		jobStatus           app.RunnerJobStatus
		run                 *app.InstallActionWorkflowRun
		expectedStatus      app.InstallActionWorkflowRunStatus
		expectedDescription string
	}{
		"failed job": {
			jobStatus:           app.RunnerJobStatusFailed,
			run:                 &app.InstallActionWorkflowRun{},
			expectedStatus:      app.InstallActionRunStatusError,
			expectedDescription: "job failed",
		},
		"cancelled job": {
			jobStatus:           app.RunnerJobStatusCancelled,
			run:                 &app.InstallActionWorkflowRun{},
			expectedStatus:      app.InstallActionRunStatusError,
			expectedDescription: "job failed",
		},
		"timed out with config timeout": {
			jobStatus: app.RunnerJobStatusTimedOut,
			run: &app.InstallActionWorkflowRun{
				Timeout:              time.Minute,
				ActionWorkflowConfig: app.ActionWorkflowConfig{Timeout: 10 * time.Minute},
			},
			expectedStatus:      app.InstallActionRunStatusTimedOut,
			expectedDescription: "action timed out after 10m0s",
		},
		"timed out adhoc run": {
			jobStatus:           app.RunnerJobStatusTimedOut,
			run:                 &app.InstallActionWorkflowRun{Timeout: 10 * time.Second},
			expectedStatus:      app.InstallActionRunStatusTimedOut,
			expectedDescription: "action timed out after 10s",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			status, description := actionRunFailureStatus(tc.jobStatus, tc.run)
			assert.Equal(t, tc.expectedStatus, status)
			assert.Equal(t, tc.expectedDescription, description)
		})
	}
}

func TestActionRunStatusV2(t *testing.T) {
	tests := map[string]struct {
		status   app.InstallActionWorkflowRunStatus
		expected app.InstallActionWorkflowRunStatus
	}{
		"timed out maps to error": {
			status:   app.InstallActionRunStatusTimedOut,
			expected: app.InstallActionRunStatusError,
		},
		"error is unchanged": {
			status:   app.InstallActionRunStatusError,
			expected: app.InstallActionRunStatusError,
		},
		"finished is unchanged": {
			status:   app.InstallActionRunStatusFinished,
			expected: app.InstallActionRunStatusFinished,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, actionRunStatusV2(tc.status))
		})
	}
}
