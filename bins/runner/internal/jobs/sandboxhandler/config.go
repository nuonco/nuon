package sandboxhandler

import (
	"encoding/json"
	"time"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

// Config holds the sandbox job configuration for a specific job type and operation.
type Config struct {
	Duration            time.Duration          `json:"duration_ms"`
	ErrorMessage        string                 `json:"error_message,omitempty"`
	Outputs             map[string]interface{} `json:"outputs,omitempty"`
	LogLines            []string               `json:"log_lines,omitempty"`
	PlanContents        string                 `json:"plan_contents,omitempty"`
	PlanDisplayContents string                 `json:"plan_display_contents,omitempty"`
	StateJSON           string                 `json:"state_json,omitempty"`
	FailAtStep          string                 `json:"fail_at_step,omitempty"`
	SleepDuration       time.Duration          `json:"sleep_duration,omitempty"`
	Timeout             time.Duration          `json:"timeout,omitempty"`
	TriggerShutdown     bool                   `json:"trigger_shutdown,omitempty"`
}

// ConfigFromAPI converts a runner SDK SandboxConfig into a handler Config.
func ConfigFromAPI(cfg *nuonrunner.SandboxConfig) *Config {
	var logLines []string
	if len(cfg.LogLines) > 0 {
		_ = json.Unmarshal(cfg.LogLines, &logLines)
	}

	var outputs map[string]interface{}
	if len(cfg.Outputs) > 0 {
		_ = json.Unmarshal(cfg.Outputs, &outputs)
	}

	return &Config{
		Duration:            cfg.Duration,
		ErrorMessage:        cfg.ErrorMessage,
		LogLines:            logLines,
		PlanContents:        cfg.PlanContents,
		PlanDisplayContents: cfg.PlanDisplayContents,
		StateJSON:           cfg.StateJSON,
		FailAtStep:          cfg.FailAtStep,
		SleepDuration:       cfg.SleepDuration,
		Timeout:             cfg.Timeout,
		TriggerShutdown:     cfg.TriggerShutdown,
		Outputs:             outputs,
	}
}
