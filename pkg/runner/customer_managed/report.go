package customermanaged

import (
	"encoding/json"
	"time"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/statestore"
)

// Report is the post-trip summary of an offline run: everything a vendor or
// customer needs to pull out of the state store after the runner exits,
// collated from status.json and the per-step result/outputs/executions files.
type Report struct {
	InstallID  string       `json:"install_id"`
	RunID      string       `json:"run_id"`
	Status     string       `json:"status"`
	FailedStep string       `json:"failed_step,omitempty"`
	StartedAt  time.Time    `json:"started_at"`
	FinishedAt *time.Time   `json:"finished_at,omitempty"`
	Steps      []StepReport `json:"steps"`
}

type StepReport struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	JobType      string          `json:"job_type"`
	JobOperation string          `json:"job_operation"`
	JobGroup     string          `json:"job_group"`
	Status       string          `json:"status"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	FinishedAt   *time.Time      `json:"finished_at,omitempty"`
	Error        string          `json:"error,omitempty"`
	Executions   int             `json:"executions"`
	Success      *bool           `json:"success,omitempty"`
	ErrorCode    string          `json:"error_code,omitempty"`
	Outputs      json.RawMessage `json:"outputs,omitempty"`
}

// BuildReport collates the run status with per-step results and outputs.
// Steps that never produced a result file (apply-plan steps report outputs
// only) still get a report row; success is then inferred from step status.
func BuildReport(envelope *Envelope, status *statestore.Status, store statestore.Store) (*Report, error) {
	report := &Report{
		InstallID:  status.InstallID,
		RunID:      status.RunID,
		Status:     status.Status,
		FailedStep: status.FailedStep,
		StartedAt:  status.StartedAt,
		FinishedAt: status.FinishedAt,
	}
	stepStatus := map[string]statestore.StepStatus{}
	for _, s := range status.Steps {
		stepStatus[s.ID] = s
	}
	for _, step := range envelope.Steps {
		row := StepReport{ID: step.ID, Name: step.Name, JobType: step.JobType, JobOperation: step.JobOperation, JobGroup: step.JobGroup}
		if s, ok := stepStatus[step.ID]; ok {
			row.Status = s.Status
			row.StartedAt = s.StartedAt
			row.FinishedAt = s.FinishedAt
			row.Error = s.Error
		}
		if raw, ok, err := store.ReadExecutions(step.ID); err != nil {
			return nil, err
		} else if ok {
			var executions []json.RawMessage
			if err := json.Unmarshal(raw, &executions); err == nil {
				row.Executions = len(executions)
			}
		}
		if raw, ok, err := store.ReadResult(step.ID); err != nil {
			return nil, err
		} else if ok {
			var result struct {
				Success   bool   `json:"success"`
				ErrorCode string `json:"error_code"`
			}
			if err := json.Unmarshal(raw, &result); err == nil {
				row.Success = &result.Success
				row.ErrorCode = result.ErrorCode
			}
		}
		if row.Success == nil && row.Status != "" {
			finished := stepSucceeded(row.Status)
			row.Success = &finished
		}
		if outputs, ok := status.Outputs[step.ID]; ok && string(outputs) != "null" {
			row.Outputs = outputs
		}
		report.Steps = append(report.Steps, row)
	}
	return report, nil
}
