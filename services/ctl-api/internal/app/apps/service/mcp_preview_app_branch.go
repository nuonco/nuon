package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpPreviewAppBranchInput struct {
	App         string `json:"app" jsonschema:"app name or ID"`
	Branch      string `json:"branch" jsonschema:"app branch name or ID to preview against"`
	Install     string `json:"install,omitempty" jsonschema:"install name or ID to preview against (required unless mode is build-only)"`
	PRNumber    int    `json:"pr_number,omitempty" jsonschema:"pull request number to preview; mutually exclusive with git_ref and app_config_id"`
	GitRef      string `json:"git_ref,omitempty" jsonschema:"git branch or ref to preview; mutually exclusive with pr_number and app_config_id"`
	HeadSHA     string `json:"head_sha,omitempty" jsonschema:"optional commit SHA for the preview source"`
	Mode        string `json:"mode,omitempty" jsonschema:"plan-only (default), apply, or build-only. Ask the user before apply"`
	AppConfigID string `json:"app_config_id,omitempty" jsonschema:"synced app config ID for a local-source preview (after nuon apps sync). HTTP MCP cannot read the local workspace"`
	Force       bool   `json:"force,omitempty" jsonschema:"force rebuild all components"`
	AutoApprove bool   `json:"auto_approve,omitempty" jsonschema:"skip the approval gate before deploy steps"`
}

type mcpPreviewAppBranchResult struct {
	RunID       string `json:"run_id"`
	WorkflowID  string `json:"workflow_id,omitempty"`
	AppID       string `json:"app_id"`
	AppName     string `json:"app_name"`
	BranchID    string `json:"branch_id"`
	BranchName  string `json:"branch_name"`
	InstallID   string `json:"install_id,omitempty"`
	InstallName string `json:"install_name,omitempty"`
	Source      string `json:"source"`
	Mode        string `json:"mode,omitempty"`
	PRNumber    *int   `json:"pr_number,omitempty"`
	GitRef      string `json:"git_ref,omitempty"`
	HeadSHA     string `json:"head_sha,omitempty"`
}

func (s *service) mcpPreviewAppBranch(ctx context.Context, _ *mcp.CallToolRequest, in mcpPreviewAppBranchInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Write(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := s.requireAppBranches(ctx); err != nil {
		return nil, nil, err
	}
	if in.App == "" {
		return nil, nil, fmt.Errorf("app is required")
	}
	if in.Branch == "" {
		return nil, nil, fmt.Errorf("branch is required")
	}

	sourceCount := 0
	if in.PRNumber > 0 {
		sourceCount++
	}
	if in.GitRef != "" {
		sourceCount++
	}
	if in.AppConfigID != "" {
		sourceCount++
	}
	if sourceCount == 0 {
		return nil, nil, fmt.Errorf("preview source required: pass pr_number, git_ref, or app_config_id (local synced config)")
	}
	if sourceCount > 1 {
		return nil, nil, fmt.Errorf("specify only one of pr_number, git_ref, or app_config_id")
	}

	var mode app.AppBranchRunPreviewMode
	if in.Mode != "" {
		mode = app.AppBranchRunPreviewMode(in.Mode)
		if !mode.Valid() || mode == "" {
			return nil, nil, fmt.Errorf("invalid mode %q: use plan-only, apply, or build-only", in.Mode)
		}
	}

	a, err := s.findAppRef(ctx, orgID, in.App)
	if err != nil {
		return nil, nil, err
	}
	branch, err := s.findAppBranch(ctx, orgID, a.ID, in.Branch)
	if err != nil {
		return nil, nil, err
	}

	var queued app.AppBranch
	res := s.db.WithContext(ctx).
		Preload("Queue").
		First(&queued, "id = ?", branch.ID)
	if res.Error != nil {
		return nil, nil, fmt.Errorf("unable to load app branch queue: %w", res.Error)
	}
	if queued.Queue.ID == "" {
		return nil, nil, fmt.Errorf("app branch %q has no queue", branch.Name)
	}

	var config app.AppBranchConfig
	res = s.db.WithContext(ctx).
		Where("app_branch_id = ?", branch.ID).
		Order("config_number DESC").
		First(&config)
	if res.Error != nil {
		return nil, nil, fmt.Errorf("unable to find latest branch config: %w", res.Error)
	}

	if mode == "" {
		mode = helpers.BranchPreviewConfigOrDefault(&config).Mode
	}
	if mode != app.AppBranchRunPreviewModeBuildOnly && in.Install == "" {
		return nil, nil, fmt.Errorf("install is required unless mode is build-only")
	}

	previewInput := &helpers.PreviewRunInput{
		HeadSHA:          in.HeadSHA,
		InputAppConfigID: in.AppConfigID,
	}
	switch {
	case in.PRNumber > 0:
		pr := in.PRNumber
		previewInput.Source = app.AppBranchRunPreviewSourcePR
		previewInput.PRNumber = &pr
	case in.GitRef != "":
		previewInput.Source = app.AppBranchRunPreviewSourceBranch
		previewInput.GitRef = in.GitRef
	default:
		previewInput.Source = app.AppBranchRunPreviewSourceLocal
	}

	override := &app.AppBranchPreviewOverride{Mode: &mode}
	var resolvedInstall *app.Install
	if in.Install != "" {
		inst, err := s.findInstallOnApp(ctx, orgID, a.ID, in.Install)
		if err != nil {
			return nil, nil, err
		}
		resolvedInstall = inst
		override.InstallID = &inst.ID
	}
	if override.Mode != nil || override.InstallID != nil {
		previewInput.Override = override
	}

	runType := app.AppBranchRunTypeGitPreview
	if previewInput.Source == app.AppBranchRunPreviewSourceLocal {
		runType = app.AppBranchRunTypeManual
	}

	workflowMeta := map[string]string{
		"app_id":        a.ID,
		"config_id":     config.ID,
		"config_number": strconv.Itoa(config.ConfigNumber),
		"force":         strconv.FormatBool(in.Force),
		"event_type":    "manual",
	}
	if in.AppConfigID != "" {
		workflowMeta["app_config_id"] = in.AppConfigID
	}
	prNumber := previewInput.PRNumber
	headSHA := in.HeadSHA
	if prNumber != nil {
		workflowMeta["pr_number"] = strconv.Itoa(*prNumber)
	}
	if headSHA != "" {
		workflowMeta["head_sha"] = headSHA
	}

	approvalOption := app.InstallApprovalOptionApproveAll
	if !in.AutoApprove {
		approvalOption, err = s.helpers.ResolveAppBranchApprovalOption(ctx, a.ID, branch.ID, config.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to resolve approval option: %w", err)
		}
	}

	triggerResp, err := s.helpers.TriggerAppBranchRun(ctx, &helpers.TriggerAppBranchRunRequest{
		Run: helpers.CreateAppBranchRunRequest{
			AppBranchID:       branch.ID,
			AppBranchConfigID: config.ID,
			AppConfigID:       in.AppConfigID,
			Force:             in.Force,
			PlanOnly:          false,
			RunType:           runType,
			EventType:         "manual",
			PRNumber:          prNumber,
			HeadSHA:           headSHA,
			Preview:           previewInput,
		},
		QueueID:        queued.Queue.ID,
		Metadata:       workflowMeta,
		ApprovalOption: approvalOption,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to trigger app branch preview: %w", err)
	}

	run := triggerResp.Run
	var reloaded app.AppBranchRun
	if err := s.db.WithContext(ctx).
		Preload("Preview").
		First(&reloaded, "id = ?", run.ID).Error; err == nil {
		run = &reloaded
	}

	result := mcpPreviewAppBranchResult{
		RunID:      run.ID,
		AppID:      a.ID,
		AppName:    a.Name,
		BranchID:   branch.ID,
		BranchName: branch.Name,
		Source:     string(previewInput.Source),
		Mode:       string(mode),
		PRNumber:   prNumber,
		GitRef:     in.GitRef,
		HeadSHA:    headSHA,
	}
	if triggerResp.Workflow != nil {
		result.WorkflowID = triggerResp.Workflow.ID
	} else if run.WorkflowID != nil {
		result.WorkflowID = *run.WorkflowID
	}
	if resolvedInstall != nil {
		result.InstallID = resolvedInstall.ID
		result.InstallName = resolvedInstall.Name
	}
	if run.Preview != nil {
		if result.InstallID == "" {
			result.InstallID = run.Preview.InstallID
			result.InstallName = run.Preview.InstallName
		}
		if run.Preview.Mode != "" {
			result.Mode = string(run.Preview.Mode)
		}
		if result.GitRef == "" {
			result.GitRef = run.Preview.GitRef
		}
	}

	return apiPkg.MCPJSONResult(result)
}
