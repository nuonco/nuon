package nuon

import (
	"bytes"
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
	ID          string            `json:"id"`
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	AppID       string            `json:"app_id,omitempty"`
	Status      string            `json:"status,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`

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

// CreateRunbookRequest is the request body for creating a runbook.
type CreateRunbookRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// UpdateRunbookRequest is the request body for updating a runbook.
type UpdateRunbookRequest struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// CreateRunbookConfigRequest is the request body for creating a runbook config.
type CreateRunbookConfigRequest struct {
	AppConfigID *string                           `json:"app_config_id,omitempty"`
	Readme      string                            `json:"readme,omitempty"`
	Steps       []*CreateRunbookStepConfigRequest `json:"steps"`
}

// CreateRunbookStepConfigRequest is the request body for creating a runbook step config.
type CreateRunbookStepConfigRequest struct {
	Name               string            `json:"name"`
	Type               string            `json:"type"`
	Idx                int64             `json:"idx"`
	ComponentName      string            `json:"component_name,omitempty"`
	DeployDependencies bool              `json:"deploy_dependencies,omitempty"`
	ActionName         string            `json:"action_name,omitempty"`
	Command            string            `json:"command,omitempty"`
	InlineContents     string            `json:"inline_contents,omitempty"`
	EnvVars            map[string]string `json:"env_vars,omitempty"`
	Timeout            int64             `json:"timeout,omitempty"`
	Role               string            `json:"role,omitempty"`
}

// RunbookWorkflow is a minimal workflow representation for runbook runs.
type RunbookWorkflow struct {
	ID                string                 `json:"id"`
	Status            *RunbookWorkflowStatus `json:"status,omitempty"`
	StatusDescription string                 `json:"status_description,omitempty"`
}

// RunbookWorkflowStatus mirrors the API's composite status object (app.CompositeStatus).
type RunbookWorkflowStatus struct {
	Status                 string `json:"status,omitempty"`
	StatusHumanDescription string `json:"status_human_description,omitempty"`
}

// GetAppRunbook retrieves a runbook by name or ID for an app.
func (c *client) GetAppRunbook(ctx context.Context, appID, nameOrID string) (*Runbook, error) {
	reqURL := fmt.Sprintf("%s/v1/apps/%s/runbooks/%s", c.APIURL, appID, nameOrID)

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
		return nil, newHTTPAPIError(resp.StatusCode, string(body))
	}

	var runbook Runbook
	if err := json.NewDecoder(resp.Body).Decode(&runbook); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &runbook, nil
}

// CreateRunbook creates a new runbook for an app.
func (c *client) CreateRunbook(ctx context.Context, appID string, createReq *CreateRunbookRequest) (*Runbook, error) {
	reqURL := fmt.Sprintf("%s/v1/apps/%s/runbooks", c.APIURL, appID)

	body, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, newHTTPAPIError(resp.StatusCode, string(respBody))
	}

	var runbook Runbook
	if err := json.NewDecoder(resp.Body).Decode(&runbook); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &runbook, nil
}

// UpdateRunbook updates a runbook.
func (c *client) UpdateRunbook(ctx context.Context, runbookID string, updateReq *UpdateRunbookRequest) (*Runbook, error) {
	reqURL := fmt.Sprintf("%s/v1/apps/_/runbooks/%s", c.APIURL, runbookID)

	body, err := json.Marshal(updateReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, newHTTPAPIError(resp.StatusCode, string(respBody))
	}

	var runbook Runbook
	if err := json.NewDecoder(resp.Body).Decode(&runbook); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &runbook, nil
}

// CreateRunbookConfig creates a new config for a runbook.
func (c *client) CreateRunbookConfig(ctx context.Context, runbookID string, createReq *CreateRunbookConfigRequest) (*RunbookConfig, error) {
	reqURL := fmt.Sprintf("%s/v1/apps/_/runbooks/%s/configs", c.APIURL, runbookID)

	body, err := json.Marshal(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, newHTTPAPIError(resp.StatusCode, string(respBody))
	}

	var cfg RunbookConfig
	if err := json.NewDecoder(resp.Body).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &cfg, nil
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
		return nil, newHTTPAPIError(resp.StatusCode, string(body))
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
		return nil, newHTTPAPIError(resp.StatusCode, string(body))
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
		return nil, newHTTPAPIError(resp.StatusCode, string(body))
	}

	var run InstallRunbookRun
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &run, nil
}
