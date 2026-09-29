package service

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/sandboxmode/templates"
)

type SandboxConfigResponse struct {
	ID        string    `json:"id"`
	JobType   string    `json:"job_type"`
	Operation string    `json:"operation,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Enabled         bool          `json:"enabled"`
	Duration        time.Duration `json:"duration" swaggertype:"primitive,integer"`
	SleepDuration   time.Duration `json:"sleep_duration,omitempty" swaggertype:"primitive,integer"`
	TriggerShutdown bool          `json:"trigger_shutdown,omitempty"`

	LogLines            json.RawMessage `json:"log_lines,omitempty"`
	PlanContents        string          `json:"plan_contents,omitempty"`
	PlanDisplayContents string          `json:"plan_display_contents,omitempty"`
	StateJSON           string          `json:"state_json,omitempty"`
	Outputs             json.RawMessage `json:"outputs,omitempty"`

	ErrorMessage string `json:"error_message,omitempty"`
}

func convertToSandboxConfigResponse(cfg app.SandboxModeJobConfig) SandboxConfigResponse {
	resp := SandboxConfigResponse{
		ID:              cfg.ID,
		JobType:         cfg.JobType,
		Operation:       cfg.Operation,
		CreatedAt:       cfg.CreatedAt,
		UpdatedAt:       cfg.UpdatedAt,
		Enabled:         cfg.Enabled,
		Duration:        cfg.Duration,
		SleepDuration:   cfg.SleepDuration,
		TriggerShutdown: cfg.TriggerShutdown,
	}

	if cfg.LogTemplate != "" {
		if t := templates.FindTemplate(cfg.LogTemplate); t != nil {
			lines := strings.Split(t.Contents, "\n")
			if data, err := json.Marshal(lines); err == nil {
				resp.LogLines = data
			}
		}
	}

	if cfg.PlanTemplate != "" {
		if t := templates.FindTemplate(cfg.PlanTemplate); t != nil {
			resp.PlanContents = t.Contents
		}
	}

	if cfg.PlanDisplayTemplate != "" {
		if t := templates.FindTemplate(cfg.PlanDisplayTemplate); t != nil {
			resp.PlanDisplayContents = t.Contents
		}
	}

	if cfg.StateTemplate != "" {
		if t := templates.FindTemplate(cfg.StateTemplate); t != nil {
			resp.StateJSON = t.Contents
		}
	}

	if cfg.OutputTemplate != "" {
		if t := templates.FindTemplate(cfg.OutputTemplate); t != nil {
			resp.Outputs = json.RawMessage(t.Contents)
		}
	}

	if cfg.ShouldError {
		resp.ErrorMessage = "sandbox error"
		if cfg.ErrorMessage != "" {
			resp.ErrorMessage = cfg.ErrorMessage
		}
	}
	if cfg.Panic {
		resp.ErrorMessage = "sandbox: panic requested"
	}

	return resp
}
