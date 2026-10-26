package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

const (
	mcpSummaryFailingWorkflowLimit = 5
	mcpSummaryInputValueLimit      = 80
)

type mcpInstallSummaryInput struct {
	Install string `json:"install" jsonschema:"install name or ID"`
}

type mcpInstallSummary struct {
	ID                       string                       `json:"id"`
	Name                     string                       `json:"name"`
	App                      mcpInstallAppRef             `json:"app"`
	AppBranch                *mcpInstallBranchRef         `json:"app_branch,omitempty"`
	Description              string                       `json:"description,omitempty"`
	CloudPlatform            string                       `json:"cloud_platform,omitempty"`
	Lifecycle                mcpInstallLifecycle          `json:"lifecycle"`
	Version                  mcpInstallVersion            `json:"version"`
	Components               []mcpInstallSummaryComponent `json:"components"`
	Health                   mcpInstallHealthStatuses     `json:"health"`
	FailingWorkflows         []mcpInstallFailure          `json:"failing_workflows,omitempty"`
	SandboxStatus            string                       `json:"sandbox_status,omitempty"`
	SandboxStatusDescription string                       `json:"sandbox_status_description,omitempty"`
	RunnerStatus             string                       `json:"runner_status,omitempty"`
	RunnerStatusDescription  string                       `json:"runner_status_description,omitempty"`
	Inputs                   []mcpInstallSummaryInputRow  `json:"inputs"`
	LastUpdate               *mcpInstallLastUpdate        `json:"last_update,omitempty"`
	NextActions              []mcpNextAction              `json:"next_actions"`
}

type mcpInstallSummaryComponent struct {
	Name              string `json:"name"`
	Type              string `json:"type,omitempty"`
	Status            string `json:"status,omitempty"`
	StatusDescription string `json:"status_description,omitempty"`
	HealthStatus      string `json:"health_status,omitempty"`
}

type mcpInstallSummaryInputRow struct {
	Group       string `json:"group,omitempty"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Type        string `json:"type,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
	Set         bool   `json:"set"`
	Value       string `json:"value,omitempty"`
}

type mcpInstallVersion struct {
	CurrentAppConfigID  string                   `json:"current_app_config_id,omitempty"`
	AppliedAppConfigID  string                   `json:"applied_app_config_id,omitempty"`
	ExpectedAppConfigID string                   `json:"expected_app_config_id,omitempty"`
	Status              string                   `json:"status,omitempty"`
	AppliedCommit       *mcpInstallVersionCommit `json:"applied_commit,omitempty"`
	ExpectedCommit      *mcpInstallVersionCommit `json:"expected_commit,omitempty"`
}

type mcpInstallVersionCommit struct {
	SHA       string `json:"sha,omitempty"`
	RunID     string `json:"run_id,omitempty"`
	Message   string `json:"message,omitempty"`
	Author    string `json:"author,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	RunStatus string `json:"run_status,omitempty"`
}

type mcpInstallFailure struct {
	WorkflowID   string `json:"workflow_id"`
	Name         string `json:"name,omitempty"`
	WorkflowType string `json:"workflow_type"`
	Status       string `json:"status"`
	Description  string `json:"description,omitempty"`
	CreatedAt    string `json:"created_at"`
}

func (s *service) mcpInstallSummary(ctx context.Context, _ *mcp.CallToolRequest, in mcpInstallSummaryInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.Install == "" {
		return nil, nil, fmt.Errorf("install is required")
	}

	install, err := s.findInstall(ctx, orgID, in.Install)
	if err != nil {
		return nil, nil, err
	}

	overview, err := s.mcpInstallOverviewFor(ctx, orgID, install)
	if err != nil {
		return nil, nil, err
	}
	versionOverview, err := s.getInstallOverview(ctx, orgID, install.ID, false)
	if err != nil {
		return nil, nil, err
	}
	components, err := s.mcpInstallSummaryComponents(ctx, install.ID)
	if err != nil {
		return nil, nil, err
	}
	inputs, err := s.mcpInstallSummaryInputs(ctx, install)
	if err != nil {
		return nil, nil, err
	}
	failing, err := s.mcpFailingInstallWorkflows(ctx, orgID, install.ID)
	if err != nil {
		return nil, nil, err
	}
	health := mcpInstallHealthStatuses{
		Runner: mcpInstallStatusAxis{
			Status:      string(install.RunnerStatus),
			Description: install.RunnerStatusDescription,
		},
		Sandbox: mcpSandboxStatus(install),
		Components: mcpInstallStatusAxis{
			Status:      string(install.CompositeComponentStatus),
			Description: install.CompositeComponentStatusDescription,
		},
	}
	if install.CompositeHealthStatus != "" {
		health.Health = &mcpInstallStatusAxis{
			Status:      string(install.CompositeHealthStatus),
			Description: install.CompositeHealthStatusDescription,
		}
	}

	return apiPkg.MCPJSONResult(mcpInstallSummary{
		ID:            overview.ID,
		Name:          overview.Name,
		App:           overview.App,
		AppBranch:     overview.AppBranch,
		Description:   overview.Description,
		CloudPlatform: overview.CloudPlatform,
		Lifecycle:     overview.Lifecycle,
		Version: mcpInstallVersion{
			CurrentAppConfigID:  versionOverview.ConfigDrift.CurrentAppConfigID,
			AppliedAppConfigID:  install.AppConfigRef.AppliedConfigID,
			ExpectedAppConfigID: install.AppConfigRef.ExpectedConfigID,
			Status:              versionOverview.BranchTracking.Status,
			AppliedCommit:       mcpInstallVersionCommitFrom(versionOverview.BranchTracking.AppliedCommit),
			ExpectedCommit:      mcpInstallVersionCommitFrom(versionOverview.BranchTracking.ExpectedCommit),
		},
		Components:               components,
		Health:                   health,
		FailingWorkflows:         failing,
		SandboxStatus:            overview.SandboxStatus,
		SandboxStatusDescription: overview.SandboxStatusDescription,
		RunnerStatus:             overview.RunnerStatus,
		RunnerStatusDescription:  overview.RunnerStatusDescription,
		Inputs:                   inputs,
		LastUpdate:               overview.LastUpdate,
		NextActions:              mcpInstallSummaryActions(install, failing),
	})
}

func mcpInstallVersionCommitFrom(commit *InstallOverviewCommit) *mcpInstallVersionCommit {
	if commit == nil {
		return nil
	}
	result := &mcpInstallVersionCommit{
		SHA:       commit.SHA,
		RunID:     commit.RunID,
		Message:   commit.Message,
		Author:    commit.Author,
		RunStatus: commit.RunStatus,
	}
	if commit.CreatedAt != nil {
		result.CreatedAt = apiPkg.MCPTime(*commit.CreatedAt)
	}
	return result
}

func (s *service) mcpFailingInstallWorkflows(ctx context.Context, orgID, installID string) ([]mcpInstallFailure, error) {
	var workflows []app.Workflow
	err := s.db.WithContext(ctx).
		Where(app.Workflow{OrgID: orgID, OwnerID: installID}).
		Where("status->>'status' IN ?", []string{string(app.StatusError), string(app.StatusFailedPendingRetry)}).
		Order("created_at DESC").
		Limit(mcpSummaryFailingWorkflowLimit).
		Find(&workflows).Error
	if err != nil {
		return nil, fmt.Errorf("unable to list failing install workflows: %w", err)
	}
	out := make([]mcpInstallFailure, 0, len(workflows))
	for _, workflow := range workflows {
		out = append(out, mcpInstallFailure{
			WorkflowID:   workflow.ID,
			Name:         workflow.Name,
			WorkflowType: string(workflow.Type),
			Status:       string(workflow.Status.Status),
			Description:  workflow.Status.StatusHumanDescription,
			CreatedAt:    apiPkg.MCPTime(workflow.CreatedAt),
		})
	}
	return out, nil
}

func (s *service) mcpInstallSummaryComponents(ctx context.Context, installID string) ([]mcpInstallSummaryComponent, error) {
	var components []app.InstallComponent
	err := s.db.WithContext(ctx).
		Preload("Component").
		Where(app.InstallComponent{InstallID: installID}).
		Find(&components).Error
	if err != nil {
		return nil, fmt.Errorf("unable to list install components: %w", err)
	}
	sort.Slice(components, func(i, j int) bool {
		return components[i].Component.Name < components[j].Component.Name
	})
	out := make([]mcpInstallSummaryComponent, 0, len(components))
	for _, component := range components {
		out = append(out, mcpInstallSummaryComponent{
			Name:              component.Component.Name,
			Type:              string(component.Component.Type),
			Status:            string(component.Status),
			StatusDescription: component.StatusDescription,
			HealthStatus:      string(component.HealthStatus),
		})
	}
	return out, nil
}

func (s *service) mcpInstallSummaryInputs(ctx context.Context, install *app.Install) ([]mcpInstallSummaryInputRow, error) {
	redacted := map[string]string{}
	if len(install.InstallInputs) > 0 {
		for name, value := range install.InstallInputs[0].ValuesRedacted {
			if value != nil {
				redacted[name] = *value
			}
		}
	}
	config, err := s.helpers.GetPinnedAppInputConfig(ctx, install.AppID, install.AppConfigID)
	if err != nil {
		return nil, err
	}
	if config == nil {
		return mcpSummaryInputsFromValues(redacted), nil
	}
	inputs := append([]app.AppInput(nil), config.AppInputs...)
	sort.Slice(inputs, func(i, j int) bool {
		if inputs[i].AppInputGroup.Index != inputs[j].AppInputGroup.Index {
			return inputs[i].AppInputGroup.Index < inputs[j].AppInputGroup.Index
		}
		return inputs[i].Index < inputs[j].Index
	})
	seen := map[string]bool{}
	out := make([]mcpInstallSummaryInputRow, 0, len(inputs))
	for _, input := range inputs {
		seen[input.Name] = true
		group := input.AppInputGroup.DisplayName
		if group == "" {
			group = input.AppInputGroup.Name
		}
		out = append(out, mcpSummaryInput(group, input.Name, input.DisplayName, string(input.Type), input.Required, input.Sensitive, redacted[input.Name]))
	}
	extra := mcpSummaryInputsFromValues(redacted)
	for _, input := range extra {
		if seen[input.Name] {
			continue
		}
		out = append(out, input)
	}
	return out, nil
}

func mcpSummaryInputsFromValues(values map[string]string) []mcpInstallSummaryInputRow {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make([]mcpInstallSummaryInputRow, 0, len(names))
	for _, name := range names {
		out = append(out, mcpSummaryInput("", name, "", "", false, false, values[name]))
	}
	return out
}

func mcpSummaryInput(group, name, displayName, inputType string, required, sensitive bool, value string) mcpInstallSummaryInputRow {
	row := mcpInstallSummaryInputRow{
		Group:       group,
		Name:        name,
		DisplayName: displayName,
		Type:        inputType,
		Required:    required,
		Sensitive:   sensitive,
		Set:         value != "",
		Value:       value,
	}
	if len(row.Value) > mcpSummaryInputValueLimit && row.Value != "********" {
		row.Value = fmt.Sprintf("set (%d characters)", len(value))
	}
	return row
}

func mcpInstallSummaryActions(install *app.Install, failing []mcpInstallFailure) []mcpNextAction {
	actions := []mcpNextAction{
		{
			Action: "inspect_health",
			Label:  "Health",
			Tool:   "get_install_health",
			Arguments: map[string]any{
				"install": install.ID,
			},
		},
		{
			Action: "inspect_resources",
			Label:  "Resources",
			Tool:   "list_install_resources",
			Arguments: map[string]any{
				"install": install.ID,
			},
		},
		{
			Action: "inspect_deployments",
			Label:  "Deployments",
			Tool:   "list_deploys",
			Arguments: map[string]any{
				"install_id": install.ID,
			},
		},
		{
			Action:      "make_update",
			Label:       "Make an update",
			NextActions: mcpInstallUpdateActions(install.ID),
		},
	}
	if len(failing) > 0 {
		actions = append(actions, mcpNextAction{
			Action: "inspect_failure",
			Label:  "Inspect failing workflow",
			Tool:   "get_workflow",
			Arguments: map[string]any{
				"workflow_id": failing[0].WorkflowID,
			},
		})
	}
	return append(actions, mcpNextAction{
		Action: "stop",
		Label:  "Do nothing",
	})
}

func mcpInstallUpdateActions(installID string) []mcpNextAction {
	return []mcpNextAction{
		{
			Action: "update_inputs",
			Label:  "Update install inputs",
			Tool:   "get_install_inputs",
			Arguments: map[string]any{
				"install": installID,
			},
		},
		{
			Action: "deploy_components",
			Label:  "Deploy components",
			Tool:   "list_available_roles",
			Arguments: map[string]any{
				"install":        installID,
				"operation_type": "deploy",
				"principal_type": "component",
			},
		},
		{
			Action: "reprovision_sandbox",
			Label:  "Reprovision sandbox",
			Tool:   "list_available_roles",
			Arguments: map[string]any{
				"install":        installID,
				"operation_type": "reprovision",
				"principal_type": "sandbox",
			},
		},
		{
			Action: "reprovision_install",
			Label:  "Reprovision install",
			Tool:   "list_available_roles",
			Arguments: map[string]any{
				"install":        installID,
				"operation_type": "reprovision",
				"principal_type": "sandbox",
			},
		},
		{
			Action: "back",
			Label:  "Back to install summary",
			Tool:   "install_summary",
			Arguments: map[string]any{
				"install": installID,
			},
		},
		{
			Action: "stop",
			Label:  "Do nothing",
		},
	}
}
