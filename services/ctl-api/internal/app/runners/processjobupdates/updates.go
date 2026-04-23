// Package processjobupdates exports the update names and payload types
// used by the push-based notification protocol between external state
// writers and ProcessJob workflows.
//
// The corresponding update handler registration + processJobState struct
// live in the runners/worker package with the ProcessJob workflow itself.
// Only the wire-format names and payload shapes are here so senders
// (service handlers, activities) can import without cycling back into
// worker code.
package processjobupdates

const (
	UpdateNameJobStatusChanged    = "job_status_changed"
	UpdateNameJobExecutionCreated = "job_execution_created"
	UpdateNameJobExecutionStatus  = "job_execution_status_changed"
	UpdateNameRunnerStatusChanged = "runner_status_changed"
	UpdateNameRunnerRestarted     = "runner_restarted"
)

type (
	JobStatusChangedPayload struct {
		JobID             string `json:"job_id,omitempty"`
		Status            string `json:"status,omitempty"`
		StatusDescription string `json:"status_description,omitempty"`
	}

	JobExecutionCreatedPayload struct {
		JobID          string `json:"job_id,omitempty"`
		JobExecutionID string `json:"job_execution_id,omitempty"`
	}

	JobExecutionStatusPayload struct {
		JobExecutionID string `json:"job_execution_id,omitempty"`
		Status         string `json:"status,omitempty"`
	}

	RunnerStatusChangedPayload struct {
		RunnerID string `json:"runner_id,omitempty"`
		Status   string `json:"status,omitempty"`
	}

	RunnerRestartedPayload struct {
		RunnerID  string `json:"runner_id,omitempty"`
		StartedAt string `json:"started_at,omitempty"`
	}
)
