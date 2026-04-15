package nuonrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SandboxConfig represents a per-runner, per-job-type sandbox configuration
// fetched from the centralized API.
type SandboxConfig struct {
	ID        string    `json:"id"`
	RunnerID  string    `json:"runner_id"`
	JobType   string    `json:"job_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Preset       string        `json:"preset"`
	Duration     time.Duration `json:"duration"`
	FaultRate    float64       `json:"fault_rate"`
	ErrorMessage string        `json:"error_message,omitempty"`
	FailAtStep   string        `json:"fail_at_step,omitempty"`

	SleepDuration   time.Duration `json:"sleep_duration,omitempty"`
	Timeout         time.Duration `json:"timeout,omitempty"`
	TriggerShutdown bool          `json:"trigger_shutdown,omitempty"`

	LogLines     json.RawMessage `json:"log_lines,omitempty"`
	PlanContents string          `json:"plan_contents,omitempty"`
	Outputs      json.RawMessage `json:"outputs,omitempty"`
}

func (c *client) GetSandboxConfigs(ctx context.Context) ([]*SandboxConfig, error) {
	reqURL := fmt.Sprintf("%s/v1/runners/%s/sandbox-configs", c.APIURL, c.RunnerID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch sandbox configs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var configs []*SandboxConfig
	if err := json.NewDecoder(resp.Body).Decode(&configs); err != nil {
		return nil, fmt.Errorf("unable to decode sandbox configs: %w", err)
	}

	return configs, nil
}
