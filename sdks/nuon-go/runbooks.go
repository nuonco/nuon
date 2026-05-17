package nuon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// InstallRunbook represents a runbook associated with an install.
type InstallRunbook struct {
	ID          string    `json:"id"`
	CreatedByID string    `json:"created_by_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	OrgID     string `json:"org_id,omitempty"`
	InstallID string `json:"install_id,omitempty"`
	RunbookID string `json:"runbook_id,omitempty"`

	Runbook Runbook             `json:"runbook"`
	Runs    []InstallRunbookRun `json:"runs,omitempty"`

	Status string `json:"status,omitempty"`
}

// Runbook represents a runbook definition.
type Runbook struct {
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
	AppID  string `json:"app_id,omitempty"`
	Status string `json:"status,omitempty"`

	Configs     []RunbookConfig `json:"configs,omitempty"`
	ConfigCount int             `json:"config_count,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RunbookConfig represents a versioned configuration for a runbook.
type RunbookConfig struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	RunbookID   string    `json:"runbook_id,omitempty"`
	AppConfigID string    `json:"app_config_id,omitempty"`

	Readme string              `json:"readme,omitempty"`
	Steps  []RunbookStepConfig `json:"steps,omitempty"`
}

// RunbookStepConfig represents a step in a runbook config.
type RunbookStepConfig struct {
	ID   string `json:"id"`
	Idx  int    `json:"idx"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`

	// deploy fields
	ComponentName      string `json:"component_name,omitempty"`
	DeployDependencies bool   `json:"deploy_dependencies,omitempty"`

	// action reference
	ActionWorkflowID string `json:"action_workflow_id,omitempty"`

	// inline action fields
	Command        string `json:"command,omitempty"`
	InlineContents string `json:"inline_contents,omitempty"`
	Role           string `json:"role,omitempty"`
}

// InstallRunbookRun represents a single run of a runbook on an install.
type InstallRunbookRun struct {
	ID          string    `json:"id"`
	CreatedByID string    `json:"created_by_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	OrgID            string `json:"org_id,omitempty"`
	InstallID        string `json:"install_id,omitempty"`
	InstallRunbookID string `json:"install_runbook_id,omitempty"`
	RunbookConfigID  string `json:"runbook_config_id,omitempty"`
	TriggeredByID    string `json:"triggered_by_id,omitempty"`

	Status            string `json:"status,omitempty"`
	StatusDescription string `json:"status_description,omitempty"`

	InstallWorkflowID *string          `json:"install_workflow_id"`
	InstallWorkflow   *RunbookWorkflow `json:"install_workflow,omitempty"`

	ExecutionTime int64 `json:"execution_time,omitempty"`
}

// RunbookWorkflow is a minimal workflow representation for runbook runs.
type RunbookWorkflow struct {
	ID                string `json:"id"`
	Status            string `json:"status,omitempty"`
	StatusDescription string `json:"status_description,omitempty"`
}

// GetInstallRunbooks retrieves all runbooks for an install.
func (c *client) GetInstallRunbooks(ctx context.Context, installID string) ([]*InstallRunbook, error) {
	reqURL := fmt.Sprintf("%s/v1/installs/%s/runbooks", c.APIURL, installID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var runbooks []*InstallRunbook
	if err := json.NewDecoder(resp.Body).Decode(&runbooks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return runbooks, nil
}

// GetInstallRunbook retrieves a single install runbook.
func (c *client) GetInstallRunbook(ctx context.Context, installID, runbookID string) (*InstallRunbook, error) {
	reqURL := fmt.Sprintf("%s/v1/installs/%s/runbooks/%s", c.APIURL, installID, runbookID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var runbook InstallRunbook
	if err := json.NewDecoder(resp.Body).Decode(&runbook); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &runbook, nil
}

// CreateInstallRunbookRun triggers a runbook run on an install.
func (c *client) CreateInstallRunbookRun(ctx context.Context, installID, runbookID string) (*InstallRunbookRun, error) {
	reqURL := fmt.Sprintf("%s/v1/installs/%s/runbooks/%s/runs", c.APIURL, installID, runbookID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var run InstallRunbookRun
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &run, nil
}
