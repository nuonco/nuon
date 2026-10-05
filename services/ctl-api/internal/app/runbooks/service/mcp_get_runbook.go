package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpGetRunbookInput struct {
	RunbookID string `json:"runbook_id,omitempty" jsonschema:"runbook ID; use runbook when passing an install"`
	Runbook   string `json:"runbook,omitempty" jsonschema:"runbook name or ID"`
	Install   string `json:"install,omitempty" jsonschema:"install name or ID; returns the config pinned to this install"`
}

type mcpRunbookDetail struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Description        string            `json:"description,omitempty"`
	AppID              string            `json:"app_id,omitempty"`
	InstallID          string            `json:"install_id,omitempty"`
	InstallRunbookID   string            `json:"install_runbook_id,omitempty"`
	RunbookConfigID    string            `json:"runbook_config_id,omitempty"`
	InCurrentAppConfig bool              `json:"in_current_app_config"`
	Runnable           bool              `json:"runnable"`
	Message            string            `json:"message,omitempty"`
	Inputs             []mcpRunbookInput `json:"inputs"`
	Steps              []mcpRunbookStep  `json:"steps"`
}

type mcpRunbookInput struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Required    bool   `json:"required"`
	Sensitive   bool   `json:"sensitive,omitempty"`
	HasDefault  bool   `json:"has_default,omitempty"`
	Default     string `json:"default,omitempty"`
}

type mcpRunbookStep struct {
	StepID           string `json:"step_id"`
	Idx              int    `json:"idx"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	ComponentName    string `json:"component_name,omitempty"`
	ActionWorkflowID string `json:"action_workflow_id,omitempty"`
	PlanOnly         bool   `json:"plan_only,omitempty"`
}

func (s *service) mcpGetRunbook(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetRunbookInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	ref := in.Runbook
	if ref == "" {
		ref = in.RunbookID
	}
	if ref == "" {
		return nil, nil, fmt.Errorf("runbook or runbook_id is required")
	}

	if in.Install != "" {
		detail, err := s.getInstallPinnedRunbook(ctx, orgID, in.Install, ref)
		if err != nil {
			return nil, nil, err
		}
		return apiPkg.MCPJSONResult(detail)
	}

	var runbook app.Runbook
	err = s.db.WithContext(ctx).
		Preload("Configs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("created_at DESC").Limit(1)
		}).
		Preload("Configs.Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Configs.Inputs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Where(app.Runbook{OrgID: orgID}).
		Where("id = ? OR name = ?", ref, ref).
		First(&runbook).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to find runbook %q: %w", ref, err)
	}

	detail := mcpRunbookDetailFrom(runbook, nil)
	detail.Message = "Pass install to load the config pinned to an install before running. Step IDs here are from the latest app config."
	return apiPkg.MCPJSONResult(detail)
}

func (s *service) getInstallPinnedRunbook(ctx context.Context, orgID, installRef, runbookRef string) (*mcpRunbookDetail, error) {
	install, err := s.resolveInstallRef(ctx, orgID, installRef)
	if err != nil {
		return nil, err
	}

	var runbook app.Runbook
	err = s.db.WithContext(ctx).
		Where(app.Runbook{OrgID: orgID, AppID: install.AppID}).
		Where(s.db.Where(app.Runbook{ID: runbookRef}).Or(app.Runbook{Name: runbookRef})).
		First(&runbook).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find runbook %q: %w", runbookRef, err)
	}

	var installRunbook app.InstallRunbook
	err = s.db.WithContext(ctx).
		Where(app.InstallRunbook{OrgID: orgID, InstallID: install.ID, RunbookID: runbook.ID}).
		First(&installRunbook).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find install runbook %q: %w", runbookRef, err)
	}

	configQuery := s.db.WithContext(ctx).
		Preload("Steps", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Preload("Inputs", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("idx ASC")
		}).
		Where(app.RunbookConfig{RunbookID: runbook.ID, OrgID: orgID})

	var runbookConfig app.RunbookConfig
	if install.AppConfigID != "" {
		err = configQuery.Where(app.RunbookConfig{AppConfigID: install.AppConfigID}).First(&runbookConfig).Error
	} else {
		err = configQuery.Order("created_at DESC").First(&runbookConfig).Error
	}
	if err == nil {
		runbook.Configs = []app.RunbookConfig{runbookConfig}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to get runbook config: %w", err)
	}

	detail := mcpRunbookDetailFrom(runbook, install)
	detail.InstallRunbookID = installRunbook.ID
	if len(runbook.Configs) == 0 {
		detail.Runnable = false
		detail.InCurrentAppConfig = false
		detail.Message = "This runbook is not in the install's app config version."
	}
	return &detail, nil
}

func mcpRunbookDetailFrom(runbook app.Runbook, install *app.Install) mcpRunbookDetail {
	detail := mcpRunbookDetail{
		ID:          runbook.ID,
		Name:        runbook.Name,
		Description: runbook.Description,
		AppID:       runbook.AppID,
		Inputs:      []mcpRunbookInput{},
		Steps:       []mcpRunbookStep{},
	}
	if install != nil {
		detail.InstallID = install.ID
	}
	if len(runbook.Configs) == 0 {
		return detail
	}

	cfg := runbook.Configs[0]
	detail.RunbookConfigID = cfg.ID
	detail.InCurrentAppConfig = true
	detail.Runnable = true
	for _, input := range cfg.Inputs {
		item := mcpRunbookInput{
			Name:        input.Name,
			DisplayName: input.DisplayName,
			Description: input.Description,
			Type:        string(input.Type),
			Required:    input.Required,
			Sensitive:   input.Sensitive,
			HasDefault:  input.Default != "",
		}
		if !input.Sensitive {
			item.Default = input.Default
		}
		detail.Inputs = append(detail.Inputs, item)
	}
	for _, step := range cfg.Steps {
		detail.Steps = append(detail.Steps, mcpRunbookStep{
			StepID:           step.ID,
			Idx:              step.Idx,
			Name:             step.Name,
			Type:             string(step.Type),
			ComponentName:    step.ComponentName,
			ActionWorkflowID: step.ActionWorkflowID.ValueString(),
			PlanOnly:         step.PlanOnly,
		})
	}
	return detail
}
