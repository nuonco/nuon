package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type ToolExecutor struct {
	apiURL     string
	httpClient *http.Client
}

func NewToolExecutor(apiURL string) *ToolExecutor {
	return &ToolExecutor{
		apiURL:     strings.TrimRight(apiURL, "/"),
		httpClient: &http.Client{},
	}
}

type ProgressFunc func(status string)

func (e *ToolExecutor) newClient(token, orgID string) (nuon.Client, error) {
	return nuon.New(
		nuon.WithURL(e.apiURL),
		nuon.WithAuthToken(token),
		nuon.WithOrgID(orgID),
	)
}

// rawGet is a fallback for endpoints not covered by the SDK.
func (e *ToolExecutor) rawGet(ctx context.Context, token, orgID, path string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", e.apiURL+path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Nuon-Org-ID", orgID)
	req.Header.Set("Accept", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

// rawPost is a fallback for endpoints not covered by the SDK.
func (e *ToolExecutor) rawPost(ctx context.Context, token, orgID, path string, reqBody any) (string, error) {
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", e.apiURL+path, strings.NewReader(string(b)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Nuon-Org-ID", orgID)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

func (e *ToolExecutor) Execute(ctx context.Context, token, orgID string, tc ToolCall, onProgress ...ProgressFunc) (string, error) {
	var args map[string]any
	if tc.Args != "" {
		if err := json.Unmarshal([]byte(tc.Args), &args); err != nil {
			return "", fmt.Errorf("invalid tool args: %w", err)
		}
	}
	if args == nil {
		args = map[string]any{}
	}

	client, err := e.newClient(token, orgID)
	if err != nil {
		return "", fmt.Errorf("create client: %w", err)
	}

	str := func(key string) string {
		v, _ := args[key].(string)
		return v
	}

	var progress ProgressFunc
	if len(onProgress) > 0 {
		progress = onProgress[0]
	}

	switch tc.Name {
	case "list_apps":
		apps, _, err := client.GetApps(ctx, nil)
		return marshalResult(apps, err)

	case "get_app":
		app, err := client.GetApp(ctx, str("app_id"))
		return marshalResult(app, err)

	case "create_app":
		name := str("name")
		app, err := client.CreateApp(ctx, &models.ServiceCreateAppRequest{Name: &name})
		return marshalResult(app, err)

	case "get_app_config":
		return e.getLatestAppConfig(ctx, client, str("app_id"))

	case "list_components":
		comps, _, err := client.GetAppComponents(ctx, str("app_id"), nil)
		return marshalResult(comps, err)

	case "create_component":
		b := map[string]any{"name": str("name"), "kind": str("kind")}
		if deps, ok := args["dependencies"]; ok {
			b["dependencies"] = deps
		}
		return e.rawPost(ctx, token, orgID, "/v1/apps/"+str("app_id")+"/components", b)

	case "create_terraform_config":
		b := map[string]any{}
		for _, key := range []string{"connected_repo", "variables", "env_vars"} {
			if v, ok := args[key]; ok {
				b[key] = v
			}
		}
		return e.rawPost(ctx, token, orgID, fmt.Sprintf("/v1/apps/%s/components/%s/configs/terraform-module", str("app_id"), str("component_id")), b)

	case "create_helm_config":
		b := map[string]any{}
		for _, key := range []string{"connected_repo", "values"} {
			if v, ok := args[key]; ok {
				b[key] = v
			}
		}
		return e.rawPost(ctx, token, orgID, fmt.Sprintf("/v1/apps/%s/components/%s/configs/helm", str("app_id"), str("component_id")), b)

	case "create_k8s_manifest_config":
		b := map[string]any{}
		for _, key := range []string{"connected_repo", "manifest_contents"} {
			if v, ok := args[key]; ok {
				b[key] = v
			}
		}
		return e.rawPost(ctx, token, orgID, fmt.Sprintf("/v1/apps/%s/components/%s/configs/kubernetes-manifest", str("app_id"), str("component_id")), b)

	case "create_docker_build_config":
		b := map[string]any{}
		for _, key := range []string{"connected_repo", "dockerfile"} {
			if v, ok := args[key]; ok {
				b[key] = v
			}
		}
		return e.rawPost(ctx, token, orgID, fmt.Sprintf("/v1/apps/%s/components/%s/configs/docker-build", str("app_id"), str("component_id")), b)

	case "build_all_components":
		return e.rawPost(ctx, token, orgID, "/v1/apps/"+str("app_id")+"/components/build-all", nil)

	case "get_build":
		build, err := client.GetComponentBuild(ctx, str("component_id"), str("build_id"))
		return marshalResult(build, err)

	case "list_vcs_connections":
		conns, _, err := client.GetVCSConnections(ctx, nil)
		return marshalResult(conns, err)

	case "list_repos":
		return e.rawGet(ctx, token, orgID, "/v1/vcs/connections/"+str("connection_id")+"/repos")

	case "list_branches":
		return e.rawGet(ctx, token, orgID, fmt.Sprintf("/v1/vcs/connections/%s/branches?repo=%s", str("connection_id"), str("repo")))

	case "list_installs":
		query := &models.GetPaginatedQuery{Limit: 100}
		if q := str("q"); q != "" {
			query.Q = q
		}
		installs, _, err := client.GetAllInstalls(ctx, query)
		return marshalResult(installs, err)

	case "get_install":
		install, err := client.GetInstall(ctx, str("install_id"))
		return marshalResult(install, err)

	case "create_install":
		return e.createInstall(ctx, token, orgID, args, str)

	case "get_install_inputs":
		inputs, err := client.GetInstallCurrentInputs(ctx, str("install_id"))
		return marshalResult(inputs, err)

	case "get_cloud_regions":
		regions, err := client.GetCloudPlatformRegions(ctx, str("platform"))
		return marshalResult(regions, err)

	case "list_deploys":
		deploys, _, err := client.GetInstallDeploys(ctx, str("install_id"), nil)
		return marshalResult(deploys, err)

	case "get_workflows":
		workflows, _, err := client.GetWorkflows(ctx, str("install_id"), nil)
		return marshalResult(workflows, err)

	case "get_workflow_steps":
		steps, err := client.GetWorkflowSteps(ctx, str("workflow_id"))
		return marshalResult(steps, err)

	case "get_step_logs":
		logs, err := client.LogStreamReadLogs(ctx, str("log_stream_id"), "")
		return marshalResult(logs, err)

	case "get_runner_job_plan":
		plan, err := client.GetRunnerJobPlan(ctx, str("runner_job_id"))
		return plan, err

	case "get_runner":
		return e.rawGet(ctx, token, orgID, "/v1/runners/"+str("runner_id"))

	case "get_runner_health":
		checks, err := client.GetRunnerRecentHealthChecks(ctx, str("runner_id"), "")
		return marshalResult(checks, err)

	case "run_adhoc_action":
		return e.runAdhocActionBlocking(ctx, client, str("install_id"), args, str, progress)

	default:
		return "", fmt.Errorf("unknown tool: %s", tc.Name)
	}
}

func marshalResult(v any, err error) (string, error) {
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal response: %w", err)
	}
	return string(b), nil
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

func (e *ToolExecutor) getLatestAppConfig(ctx context.Context, client nuon.Client, appID string) (string, error) {
	configs, _, err := client.GetAppConfigs(ctx, appID, nil)
	if err != nil {
		return "", fmt.Errorf("list app configs: %w", err)
	}
	if len(configs) == 0 {
		return "", fmt.Errorf("app has no configs")
	}

	recurse := true
	config, err := client.GetAppConfig(ctx, appID, configs[0].ID, &recurse)
	if err != nil {
		return "", fmt.Errorf("get app config: %w", err)
	}

	summary := appConfigSummary{ConfigID: config.ID}
	if config.Runner != nil {
		summary.Platform = string(config.Runner.AppRunnerType)
	}
	if config.Input != nil {
		for _, inp := range config.Input.Inputs {
			if inp == nil {
				continue
			}
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

	return marshalResult(summary, nil)
}

func (e *ToolExecutor) createInstall(ctx context.Context, token, orgID string, args map[string]any, str func(string) string) (string, error) {
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

	return e.rawPost(ctx, token, orgID, "/v1/apps/"+str("app_id")+"/installs", b)
}

type adhocActionResult struct {
	RunID      string `json:"run_id"`
	InstallID  string `json:"install_id"`
	WorkflowID string `json:"workflow_id,omitempty"`
	Status     string `json:"status"`
	Logs       string `json:"logs,omitempty"`
	Error      string `json:"error,omitempty"`
}

func (e *ToolExecutor) runAdhocActionBlocking(ctx context.Context, client nuon.Client, installID string, args map[string]any, str func(string) string, onProgress ProgressFunc) (string, error) {
	enableKube := true
	req := &models.ServiceCreateAdHocActionRequest{
		InlineContents:   str("inline_contents"),
		EnableKubeConfig: &enableKube,
	}
	if name := str("name"); name != "" {
		req.Name = name
	}
	if timeout, ok := args["timeout"].(float64); ok && timeout > 0 {
		req.Timeout = int64(timeout)
	}

	created, err := client.CreateAdHocAction(ctx, installID, req)
	if err != nil {
		return "", fmt.Errorf("create adhoc action: %w", err)
	}

	var status string
	var logStreamID string
	deadline := time.After(5 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-deadline:
			result := adhocActionResult{RunID: created.ID, InstallID: created.InstallID, WorkflowID: created.WorkflowID, Status: "polling_timeout", Error: "Timed out waiting for action to complete after 5 minutes"}
			return marshalResult(result, nil)
		case <-ticker.C:
		}

		run, err := client.GetInstallActionWorkflowRun(ctx, installID, created.ID)
		if err != nil {
			continue
		}

		status = run.Status
		if run.LogStream != nil && run.LogStream.ID != "" {
			logStreamID = run.LogStream.ID
		}

		if onProgress != nil {
			onProgress(status)
		}

		switch status {
		case "finished", "error", "timed-out", "cancelled":
			goto done
		}
	}

done:
	result := adhocActionResult{RunID: created.ID, InstallID: created.InstallID, WorkflowID: created.WorkflowID, Status: status}

	if logStreamID != "" {
		time.Sleep(2 * time.Second)
		logs, err := client.LogStreamReadLogs(ctx, logStreamID, "")
		if err == nil && len(logs) > 0 {
			var lines []string
			for _, l := range logs {
				lines = append(lines, l.Body)
			}
			result.Logs = strings.Join(lines, "\n")
		}
	}

	if status == "error" || status == "timed-out" {
		result.Error = "Action " + status
	}

	return marshalResult(result, nil)
}
