package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ToolExecutor struct {
	apiURL string
	client *http.Client
}

func NewToolExecutor(apiURL string) *ToolExecutor {
	return &ToolExecutor{
		apiURL: strings.TrimRight(apiURL, "/"),
		client: &http.Client{},
	}
}

func (e *ToolExecutor) Execute(ctx context.Context, token, orgID string, tc ToolCall) (string, error) {
	var args map[string]any
	if tc.Args != "" {
		if err := json.Unmarshal([]byte(tc.Args), &args); err != nil {
			return "", fmt.Errorf("invalid tool args: %w", err)
		}
	}
	if args == nil {
		args = map[string]any{}
	}

	method, path, body, err := e.resolveToolCall(tc.Name, args)
	if err != nil {
		return "", err
	}

	if method == "__get_app_config__" {
		return e.getLatestAppConfig(ctx, token, orgID, path)
	}

	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = strings.NewReader(string(b))
	}

	url := e.apiURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Nuon-Org-ID", orgID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return string(respBody), nil
}

func (e *ToolExecutor) apiGet(ctx context.Context, token, orgID, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", e.apiURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Nuon-Org-ID", orgID)
	req.Header.Set("Accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
	return body, nil
}

type appInputSummary struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Default     string `json:"default,omitempty"`
	Required    bool   `json:"required"`
}

type appConfigSummary struct {
	ConfigID string            `json:"config_id"`
	Platform string            `json:"platform,omitempty"`
	Inputs   []appInputSummary `json:"inputs"`
}

func (e *ToolExecutor) getLatestAppConfig(ctx context.Context, token, orgID, appID string) (string, error) {
	configsBody, err := e.apiGet(ctx, token, orgID, "/v1/apps/"+appID+"/configs")
	if err != nil {
		return "", fmt.Errorf("list app configs: %w", err)
	}

	var configs []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(configsBody, &configs); err != nil {
		return "", fmt.Errorf("parse app configs: %w", err)
	}
	if len(configs) == 0 {
		return "", fmt.Errorf("app has no configs")
	}

	configBody, err := e.apiGet(ctx, token, orgID, "/v1/apps/"+appID+"/config/"+configs[0].ID+"?recurse=true")
	if err != nil {
		return "", fmt.Errorf("get app config: %w", err)
	}

	var raw struct {
		ID    string `json:"id"`
		Input *struct {
			Inputs []struct {
				Name        string `json:"name"`
				DisplayName string `json:"display_name"`
				Description string `json:"description"`
				Type        string `json:"type"`
				Default     string `json:"default"`
				Required    bool   `json:"required"`
			} `json:"inputs"`
		} `json:"input"`
		RunnerConfig *struct {
			AppRunnerType string `json:"app_runner_type"`
		} `json:"runner_config"`
	}
	if err := json.Unmarshal(configBody, &raw); err != nil {
		return "", fmt.Errorf("parse app config: %w", err)
	}

	summary := appConfigSummary{ConfigID: raw.ID}
	if raw.RunnerConfig != nil {
		summary.Platform = raw.RunnerConfig.AppRunnerType
	}
	if raw.Input != nil {
		for _, inp := range raw.Input.Inputs {
			summary.Inputs = append(summary.Inputs, appInputSummary{
				Name:        inp.Name,
				DisplayName: inp.DisplayName,
				Description: inp.Description,
				Type:        inp.Type,
				Default:     inp.Default,
				Required:    inp.Required,
			})
		}
	}

	b, _ := json.Marshal(summary)
	return string(b), nil
}

func (e *ToolExecutor) resolveToolCall(name string, args map[string]any) (method, path string, body any, err error) {
	str := func(key string) string {
		v, _ := args[key].(string)
		return v
	}

	switch name {
	case "list_apps":
		return "GET", "/v1/apps", nil, nil

	case "get_app":
		return "GET", "/v1/apps/" + str("app_id"), nil, nil

	case "create_app":
		return "POST", "/v1/apps", map[string]any{"name": str("name")}, nil

	case "get_app_config":
		return "__get_app_config__", str("app_id"), nil, nil

	case "list_components":
		return "GET", "/v1/apps/" + str("app_id") + "/components", nil, nil

	case "create_component":
		b := map[string]any{
			"name": str("name"),
			"kind": str("kind"),
		}
		if deps, ok := args["dependencies"]; ok {
			b["dependencies"] = deps
		}
		return "POST", "/v1/apps/" + str("app_id") + "/components", b, nil

	case "create_terraform_config":
		b := map[string]any{}
		for _, key := range []string{"connected_repo", "variables", "env_vars"} {
			if v, ok := args[key]; ok {
				b[key] = v
			}
		}
		return "POST", fmt.Sprintf("/v1/apps/%s/components/%s/configs/terraform-module", str("app_id"), str("component_id")), b, nil

	case "create_helm_config":
		b := map[string]any{}
		for _, key := range []string{"connected_repo", "values"} {
			if v, ok := args[key]; ok {
				b[key] = v
			}
		}
		return "POST", fmt.Sprintf("/v1/apps/%s/components/%s/configs/helm", str("app_id"), str("component_id")), b, nil

	case "create_k8s_manifest_config":
		b := map[string]any{}
		for _, key := range []string{"connected_repo", "manifest_contents"} {
			if v, ok := args[key]; ok {
				b[key] = v
			}
		}
		return "POST", fmt.Sprintf("/v1/apps/%s/components/%s/configs/kubernetes-manifest", str("app_id"), str("component_id")), b, nil

	case "create_docker_build_config":
		b := map[string]any{}
		for _, key := range []string{"connected_repo", "dockerfile"} {
			if v, ok := args[key]; ok {
				b[key] = v
			}
		}
		return "POST", fmt.Sprintf("/v1/apps/%s/components/%s/configs/docker-build", str("app_id"), str("component_id")), b, nil

	case "build_all_components":
		return "POST", "/v1/apps/" + str("app_id") + "/components/build-all", nil, nil

	case "get_build":
		return "GET", fmt.Sprintf("/v1/apps/%s/components/%s/builds/%s", str("app_id"), str("component_id"), str("build_id")), nil, nil

	case "list_vcs_connections":
		return "GET", "/v1/vcs/connections", nil, nil

	case "list_repos":
		return "GET", "/v1/vcs/connections/" + str("connection_id") + "/repos", nil, nil

	case "list_branches":
		return "GET", fmt.Sprintf("/v1/vcs/connections/%s/branches?repo=%s", str("connection_id"), str("repo")), nil, nil

	case "list_installs":
		return "GET", "/v1/installs", nil, nil

	case "get_install":
		return "GET", "/v1/installs/" + str("install_id"), nil, nil

	case "create_install":
		b := map[string]any{
			"app_id":   str("app_id"),
			"name":     str("name"),
			"metadata": map[string]any{"managed_by": "nuon/agent"},
		}

		approvalOption := "prompt"
		if autoApprove, ok := args["auto_approve"].(bool); ok && autoApprove {
			approvalOption = "approve-all"
		}
		b["install_config"] = map[string]any{"approval_option": approvalOption}

		if inputs, ok := args["inputs"].(map[string]any); ok && len(inputs) > 0 {
			strInputs := map[string]string{}
			for k, v := range inputs {
				if s, ok := v.(string); ok {
					strInputs[k] = s
				}
			}
			b["inputs"] = strInputs
		}

		if region := str("region"); region != "" {
			b["aws_account"] = map[string]any{"iam_role_arn": "", "region": region}
		}
		if location := str("location"); location != "" {
			b["azure_account"] = map[string]any{
				"location":                   location,
				"service_principal_app_id":   "",
				"service_principal_password": "",
				"subscription_id":            "",
				"subscription_tenant_id":     "",
			}
		}

		return "POST", "/v1/apps/" + str("app_id") + "/installs", b, nil

	case "get_install_inputs":
		return "GET", "/v1/installs/" + str("install_id") + "/inputs/current", nil, nil

	case "get_cloud_regions":
		return "GET", "/v1/general/cloud-platform/" + str("platform") + "/regions", nil, nil

	case "list_deploys":
		return "GET", "/v1/installs/" + str("install_id") + "/deploys", nil, nil

	case "get_workflows":
		return "GET", "/v1/installs/" + str("install_id") + "/workflows", nil, nil

	case "get_workflow_steps":
		return "GET", "/v1/install-workflows/" + str("workflow_id") + "/steps", nil, nil

	case "get_step_logs":
		return "GET", "/v1/log-streams/" + str("log_stream_id") + "/logs", nil, nil

	case "get_runner_job_plan":
		return "GET", "/v1/runner-jobs/" + str("runner_job_id") + "/plan", nil, nil

	case "get_runner":
		return "GET", "/v1/runners/" + str("runner_id"), nil, nil

	case "get_runner_health":
		return "GET", "/v1/runners/" + str("runner_id") + "/recent-health-checks", nil, nil

	default:
		return "", "", nil, fmt.Errorf("unknown tool: %s", name)
	}
}
