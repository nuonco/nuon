package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
)

type mcpListDeploysInput struct {
	InstallID string `json:"install_id" jsonschema:"install ID to list deploys for"`
}

type mcpDeployCommit struct {
	SHA       string `json:"sha,omitempty"`
	Message   string `json:"message,omitempty"`
	GitRef    string `json:"git_ref,omitempty"`
	SourceRef string `json:"source_ref,omitempty"`
}

type mcpDeployBranchRun struct {
	ID       string `json:"id,omitempty"`
	Branch   string `json:"branch,omitempty"`
	RunType  string `json:"run_type,omitempty"`
	Status   string `json:"status,omitempty"`
	Trigger  string `json:"trigger,omitempty"`
	PRNumber *int   `json:"pr_number,omitempty"`
	HeadSHA  string `json:"head_sha,omitempty"`
	PlanOnly bool   `json:"plan_only,omitempty"`
	Preview  bool   `json:"preview,omitempty"`
}

type mcpDeploySummary struct {
	ComponentName     string              `json:"component_name,omitempty"`
	ComponentType     string              `json:"component_type,omitempty"`
	Status            string              `json:"status"`
	StatusDescription string              `json:"status_description,omitempty"`
	Commit            *mcpDeployCommit    `json:"commit,omitempty"`
	AppBranchRun      *mcpDeployBranchRun `json:"app_branch_run,omitempty"`
	CreatedAt         string              `json:"created_at"`
	NextActions       []mcpNextAction     `json:"next_actions,omitempty"`
	ID                string              `json:"id"`
}

type mcpListDeploysResult struct {
	Deploys     []mcpDeploySummary `json:"deploys"`
	NextActions []mcpNextAction    `json:"next_actions"`
}

func (s *service) mcpListDeploys(ctx context.Context, _ *mcp.CallToolRequest, in mcpListDeploysInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
	if err != nil {
		return nil, nil, err
	}
	if in.InstallID == "" {
		return nil, nil, fmt.Errorf("install_id is required")
	}

	var deploys []app.InstallDeploy
	err = s.db.WithContext(ctx).
		Preload("InstallComponent").
		Preload("InstallComponent.Component").
		Preload("ComponentBuild").
		Preload("ComponentBuild.VCSConnectionCommit").
		Preload("ComponentBuild.AppBranchRun").
		Preload("ComponentBuild.AppBranchRun.AppBranch").
		Where(app.InstallDeploy{OrgID: orgID}).
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Where("install_components.install_id = ?", in.InstallID).
		Order("install_deploys.created_at DESC").
		Limit(20).
		Find(&deploys).Error
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list deploys: %w", err)
	}

	out := make([]mcpDeploySummary, 0, len(deploys))
	for _, d := range deploys {
		out = append(out, mcpDeploySummaryFrom(in.InstallID, d))
	}

	return apiPkg.MCPJSONResult(mcpListDeploysResult{
		Deploys: out,
		NextActions: []mcpNextAction{
			{
				Action: "back",
				Label:  "Back to install summary",
				Tool:   "install_summary",
				Arguments: map[string]any{
					"install": in.InstallID,
				},
			},
			{
				Action: "stop",
				Label:  "Do nothing",
			},
		},
	})
}

func mcpDeploySummaryFrom(installID string, d app.InstallDeploy) mcpDeploySummary {
	summary := mcpDeploySummary{
		Status:            string(d.Status),
		StatusDescription: d.StatusDescription,
		Commit:            mcpDeployCommitFrom(d.ComponentBuild),
		CreatedAt:         apiPkg.MCPTime(d.CreatedAt),
		NextActions:       mcpDeployActions(installID, d),
		ID:                d.ID,
	}
	if d.ComponentBuild != nil {
		summary.AppBranchRun = mcpDeployBranchRunFrom(d.ComponentBuild.AppBranchRun)
	}
	if d.InstallComponent != nil && d.InstallComponent.Component.ID != "" {
		summary.ComponentName = d.InstallComponent.Component.Name
		summary.ComponentType = string(d.InstallComponent.Component.Type)
	}
	return summary
}

func mcpDeployActions(installID string, d app.InstallDeploy) []mcpNextAction {
	actions := []mcpNextAction{{
		Action: "view_logs",
		Label:  "View logs",
		Tool:   "get_deploy_logs",
		Arguments: map[string]any{
			"deploy_id": d.ID,
		},
	}}
	componentInstallID := installID
	if d.InstallID != "" {
		componentInstallID = d.InstallID
	}
	if d.InstallComponentID != "" && componentInstallID != "" {
		actions = append(actions, mcpNextAction{
			Action: "view_resources",
			Label:  "View resources",
			Tool:   "list_install_resources",
			Arguments: map[string]any{
				"install":              componentInstallID,
				"install_component_id": d.InstallComponentID,
			},
		})
	}
	return actions
}

func mcpDeployCommitFrom(build *app.ComponentBuild) *mcpDeployCommit {
	if build == nil {
		return nil
	}
	commit := &mcpDeployCommit{SourceRef: build.SourceRef}
	if build.GitRef != nil {
		commit.GitRef = *build.GitRef
	}
	if build.VCSConnectionCommit != nil {
		commit.SHA = build.VCSConnectionCommit.SHA
		commit.Message = mcpFirstLine(build.VCSConnectionCommit.Message)
	}
	if commit.SHA == "" && build.AppBranchRun != nil {
		commit.SHA = firstNonEmpty(build.AppBranchRun.HeadSHA, build.AppBranchRun.Metadata.HeadSHA)
	}
	if commit.SHA == "" && commit.Message == "" && commit.GitRef == "" && commit.SourceRef == "" {
		return nil
	}
	return commit
}

func mcpDeployBranchRunFrom(run *app.AppBranchRun) *mcpDeployBranchRun {
	if run == nil || run.ID == "" {
		return nil
	}
	pr := run.PRNumber
	if pr == nil {
		pr = run.Metadata.PRNumber
	}
	return &mcpDeployBranchRun{
		ID:       run.ID,
		Branch:   run.AppBranch.Name,
		RunType:  string(run.RunType),
		Status:   run.Status,
		Trigger:  string(run.Metadata.Trigger),
		PRNumber: pr,
		HeadSHA:  firstNonEmpty(run.HeadSHA, run.Metadata.HeadSHA),
		PlanOnly: run.PlanOnly,
		Preview:  run.RunType == app.AppBranchRunTypeGitPreview,
	}
}

func mcpFirstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}
