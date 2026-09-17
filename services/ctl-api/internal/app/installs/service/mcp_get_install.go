package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetInstallInput struct {
	Install string `json:"install" jsonschema:"install name or ID"`
}

type mcpInstallOverview struct {
	ID                         string                `json:"id"`
	Name                       string                `json:"name"`
	App                        mcpInstallAppRef      `json:"app"`
	Description                string                `json:"description,omitempty"`
	CloudPlatform              string                `json:"cloud_platform,omitempty"`
	Lifecycle                  mcpInstallLifecycle   `json:"lifecycle"`
	AppConfigID                string                `json:"app_config_id,omitempty"`
	AppBranch                  *mcpInstallBranchRef  `json:"app_branch,omitempty"`
	SandboxStatus              string                `json:"sandbox_status,omitempty"`
	SandboxStatusDescription   string                `json:"sandbox_status_description,omitempty"`
	ComponentStatus            string                `json:"component_status,omitempty"`
	ComponentStatusDescription string                `json:"component_status_description,omitempty"`
	HealthStatus               string                `json:"health_status,omitempty"`
	HealthStatusDescription    string                `json:"health_status_description,omitempty"`
	RunnerStatus               string                `json:"runner_status,omitempty"`
	RunnerStatusDescription    string                `json:"runner_status_description,omitempty"`
	InputNames                 []string              `json:"input_names,omitempty"`
	LastUpdate                 *mcpInstallLastUpdate `json:"last_update,omitempty"`
}

type mcpInstallAppRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type mcpInstallBranchRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type mcpInstallLifecycle struct {
	Phase       string `json:"phase,omitempty"`
	Description string `json:"description,omitempty"`
	UpdatedAt   int64  `json:"updated_at,omitempty"`
}

type mcpInstallLastUpdate struct {
	WorkflowID    string `json:"workflow_id"`
	WorkflowType  string `json:"workflow_type"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	BranchRunID   string `json:"app_branch_run_id,omitempty"`
	BranchRunPR   *int   `json:"pr_number,omitempty"`
	BranchRunType string `json:"app_branch_run_type,omitempty"`
}

func (s *service) mcpGetInstall(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetInstallInput) (*mcp.CallToolResult, any, error) {
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

	result := mcpInstallOverview{
		ID:                         install.ID,
		Name:                       install.Name,
		App:                        mcpInstallAppRef{ID: install.AppID, Name: install.App.Name},
		CloudPlatform:              string(install.CloudPlatform),
		Lifecycle:                  mcpInstallLifecycle{Phase: string(install.LifecyclePhase.Phase), Description: install.LifecyclePhase.Description, UpdatedAt: install.LifecyclePhase.UpdatedAt},
		AppConfigID:                install.AppConfigID,
		SandboxStatus:              string(install.SandboxStatus),
		SandboxStatusDescription:   install.SandboxStatusDescription,
		ComponentStatus:            string(install.CompositeComponentStatus),
		ComponentStatusDescription: install.CompositeComponentStatusDescription,
		HealthStatus:               string(install.CompositeHealthStatus),
		HealthStatusDescription:    install.CompositeHealthStatusDescription,
		RunnerStatus:               string(install.RunnerStatus),
		RunnerStatusDescription:    install.RunnerStatusDescription,
	}
	if install.App.Description.Valid {
		result.Description = install.App.Description.String
	}
	if install.AppBranch != nil {
		result.AppBranch = &mcpInstallBranchRef{ID: install.AppBranch.ID, Name: install.AppBranch.Name}
	}
	if len(install.InstallInputs) > 0 {
		for name := range install.InstallInputs[0].ValuesRedacted {
			result.InputNames = append(result.InputNames, name)
		}
		sort.Strings(result.InputNames)
	}

	lastUpdate, err := s.mcpInstallLastUpdate(ctx, orgID, install.ID)
	if err != nil {
		return nil, nil, err
	}
	result.LastUpdate = lastUpdate

	return apiPkg.MCPJSONResult(result)
}

func (s *service) mcpInstallLastUpdate(ctx context.Context, orgID, installID string) (*mcpInstallLastUpdate, error) {
	var workflow app.Workflow
	err := s.db.WithContext(ctx).
		Where(app.Workflow{OrgID: orgID, OwnerID: installID}).
		Order("created_at DESC").
		First(&workflow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get latest install workflow: %w", err)
	}

	result := &mcpInstallLastUpdate{
		WorkflowID:   workflow.ID,
		WorkflowType: string(workflow.Type),
		Status:       string(workflow.Status.Status),
		CreatedAt:    apiPkg.MCPTime(workflow.CreatedAt),
	}
	if value := workflow.Metadata["app_branch_run_id"]; value != nil {
		result.BranchRunID = *value
		var run app.AppBranchRun
		err := s.db.WithContext(ctx).
			Where(app.AppBranchRun{ID: result.BranchRunID, OrgID: orgID}).
			First(&run).Error
		if err == nil {
			result.BranchRunPR = run.PRNumber
			result.BranchRunType = string(run.RunType)
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("unable to get app branch run for latest install workflow: %w", err)
		}
	}
	return result, nil
}
