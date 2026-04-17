package nuonrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SandboxConfig represents a per-runner, per-job-type sandbox configuration
// fetched from the centralized API.
type SandboxConfig struct {
	ID        string    `json:"id"`
	RunnerID  string    `json:"runner_id"`
	JobType   string    `json:"job_type"`
	Operation string    `json:"operation,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Enabled      bool          `json:"enabled"`
	Preset       string        `json:"preset"`
	Duration     time.Duration `json:"duration"`
	ErrorMessage string        `json:"error_message,omitempty"`
	FailAtStep   string        `json:"fail_at_step,omitempty"`

	SleepDuration   time.Duration `json:"sleep_duration,omitempty"`
	Timeout         time.Duration `json:"timeout,omitempty"`
	TriggerShutdown bool          `json:"trigger_shutdown,omitempty"`

	LogLines            json.RawMessage `json:"log_lines,omitempty"`
	PlanContents        string          `json:"plan_contents,omitempty"`
	PlanDisplayContents string          `json:"plan_display_contents,omitempty"`
	StateJSON           string          `json:"state_json,omitempty"`
	Outputs             json.RawMessage `json:"outputs,omitempty"`
}

// GetSandboxConfigs fetches all sandbox configs for this runner.
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

// GetSandboxConfig fetches the sandbox config for a specific job type and operation.
// The API performs fallback: if no config matches the exact operation, it returns
// the config for the job type with no operation set.
func (c *client) GetSandboxConfig(ctx context.Context, jobType, operation string) (*SandboxConfig, error) {
	params := url.Values{
		"job_type": {jobType},
	}
	if operation != "" {
		params.Set("operation", operation)
	}

	reqURL := fmt.Sprintf("%s/v1/runners/%s/sandbox-config?%s", c.APIURL, c.RunnerID, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch sandbox config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var cfg SandboxConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode sandbox config: %w", err)
	}

	return &cfg, nil
}
