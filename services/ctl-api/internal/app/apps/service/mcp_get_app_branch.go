package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type mcpGetAppBranchInput struct {
	App    string `json:"app" jsonschema:"app name or ID"`
	Branch string `json:"branch" jsonschema:"app branch name or ID"`
}

type mcpGetAppBranchResult struct {
	App           mcpAppRef                  `json:"app"`
	Branch        mcpAppBranchOverview       `json:"branch"`
	VCS           *mcpAppBranchVCS           `json:"vcs,omitempty"`
	InstallGroups []mcpAppBranchInstallGroup `json:"install_groups,omitempty"`
	Answers       mcpAppBranchAnswers        `json:"answers"`
	LatestRun     *mcpAppBranchRunOverview   `json:"latest_run,omitempty"`
}

type mcpAppBranchAnswers struct {
	LastRunSucceeded  *bool  `json:"last_run_succeeded"`
	LastRunStatus     string `json:"last_run_status,omitempty"`
	ChangeSummary     string `json:"change_summary"`
	DeploymentSummary string `json:"deployment_summary"`
}

type mcpAppRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type mcpAppBranchOverview struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ManagedBy string `json:"managed_by"`
}

type mcpAppBranchInstallGroup struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Order         int      `json:"order"`
	AllInstalls   bool     `json:"all_installs"`
	InstallIDs    []string `json:"install_ids,omitempty"`
	LabelSelector bool     `json:"has_label_selector"`
}

type mcpAppBranchRunOverview struct {
	ID               string                     `json:"id"`
	Status           string                     `json:"status"`
	Succeeded        bool                       `json:"succeeded"`
	AwaitingApproval bool                       `json:"awaiting_approval"`
	RunType          string                     `json:"run_type"`
	Preview          bool                       `json:"preview"`
	PlanOnly         bool                       `json:"plan_only"`
	PRNumber         *int                       `json:"pr_number,omitempty"`
	HeadSHA          string                     `json:"head_sha,omitempty"`
	BaseBranch       string                     `json:"base_branch,omitempty"`
	ErrorMessage     string                     `json:"error_message,omitempty"`
	NoConfigChanges  bool                       `json:"no_config_changes"`
	WorkflowID       string                     `json:"workflow_id,omitempty"`
	CreatedAt        string                     `json:"created_at"`
	PreviewInstallID string                     `json:"preview_install_id,omitempty"`
	PreviewInstall   string                     `json:"preview_install_name,omitempty"`
	PreviewMode      string                     `json:"preview_mode,omitempty"`
	Workflow         *mcpAppBranchWorkflowProg  `json:"workflow,omitempty"`
	Comparison       *mcpAppBranchChangeSummary `json:"comparison,omitempty"`
	Deployment       []mcpAppBranchDeployGroup  `json:"deployment,omitempty"`
}

type mcpAppBranchWorkflowProg struct {
	Status            string `json:"status"`
	StatusDescription string `json:"status_description,omitempty"`
	CompletedSteps    int    `json:"completed_steps"`
	TotalSteps        int    `json:"total_steps"`
	CurrentStep       string `json:"current_step,omitempty"`
}

type mcpAppBranchChangeSummary struct {
	BaseSHA         string                      `json:"base_sha,omitempty"`
	HeadSHA         string                      `json:"head_sha,omitempty"`
	HasGitDiff      bool                        `json:"has_git_diff"`
	HasConfigDiff   bool                        `json:"has_config_diff"`
	NoConfigChanges bool                        `json:"no_config_changes"`
	GitFilesChanged int                         `json:"git_files_changed,omitempty"`
	GitPaths        []string                    `json:"git_paths,omitempty"`
	ConfigFile      string                      `json:"config_file,omitempty"`
	ConfigAdditions int                         `json:"config_additions,omitempty"`
	ConfigRemovals  int                         `json:"config_removals,omitempty"`
	ConfigChanged   int                         `json:"config_changed,omitempty"`
	ConfigSections  []mcpAppBranchConfigSection `json:"config_sections,omitempty"`
}

type mcpAppBranchConfigSection struct {
	Name      string   `json:"name"`
	Additions int      `json:"additions"`
	Removals  int      `json:"removals"`
	Changed   int      `json:"changed"`
	Entries   []string `json:"entries,omitempty"`
}

type mcpAppBranchDeployGroup struct {
	GroupName         string                      `json:"group_name"`
	Status            string                      `json:"status"`
	StatusDescription string                      `json:"status_description,omitempty"`
	TotalInstalls     int                         `json:"total_installs"`
	CompletedInstalls int                         `json:"completed_installs"`
	FailedInstalls    int                         `json:"failed_installs"`
	Installs          []mcpAppBranchDeployInstall `json:"installs,omitempty"`
}

type mcpAppBranchDeployInstall struct {
	InstallID   string `json:"install_id"`
	InstallName string `json:"install_name,omitempty"`
	Status      string `json:"status"`
	Phase       string `json:"phase,omitempty"`
	WorkflowID  string `json:"workflow_id,omitempty"`
}

func (s *service) mcpGetAppBranch(ctx context.Context, _ *mcp.CallToolRequest, in mcpGetAppBranchInput) (*mcp.CallToolResult, any, error) {
	orgID, err := require.Read(ctx)
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

	a, err := s.findAppRef(ctx, orgID, in.App)
	if err != nil {
		return nil, nil, err
	}
	branch, err := s.findAppBranch(ctx, orgID, a.ID, in.Branch)
	if err != nil {
		return nil, nil, err
	}

	result := mcpGetAppBranchResult{
		App: mcpAppRef{ID: a.ID, Name: a.Name},
		Branch: mcpAppBranchOverview{
			ID:        branch.ID,
			Name:      branch.Name,
			ManagedBy: string(branch.ManagedBy),
		},
	}

	cfg, err := s.latestAppBranchConfig(ctx, branch.ID)
	if err != nil {
		return nil, nil, err
	}
	result.VCS = mcpBranchVCS(cfg)
	if cfg != nil {
		for _, g := range cfg.InstallGroups {
			result.InstallGroups = append(result.InstallGroups, mcpAppBranchInstallGroup{
				ID:            g.ID,
				Name:          g.Name,
				Order:         g.Order,
				AllInstalls:   g.AllInstalls,
				InstallIDs:    []string(g.InstallIDs),
				LabelSelector: g.LabelSelector != nil,
			})
		}
	}

	run, err := s.latestAppBranchRun(ctx, branch.ID)
	if err != nil {
		return nil, nil, err
	}
	if run != nil {
		overview, err := s.appBranchRunOverview(ctx, run)
		if err != nil {
			return nil, nil, err
		}
		result.LatestRun = overview
	}
	result.Answers = mcpBranchOverviewAnswers(result)

	return apiPkg.MCPJSONResult(result)
}

func (s *service) appBranchRunOverview(ctx context.Context, run *app.AppBranchRun) (*mcpAppBranchRunOverview, error) {
	if err := s.markRunAwaitingApproval(ctx, run); err != nil {
		return nil, err
	}

	headSHA := run.HeadSHA
	if run.VCSConnectionCommit != nil && run.VCSConnectionCommit.SHA != "" {
		headSHA = run.VCSConnectionCommit.SHA
	}

	out := &mcpAppBranchRunOverview{
		ID:               run.ID,
		Status:           run.Status,
		Succeeded:        run.Status == "success",
		AwaitingApproval: run.AwaitingApproval,
		RunType:          string(run.RunType),
		Preview:          run.IsPreview(),
		PlanOnly:         run.PlanOnly,
		PRNumber:         run.PRNumber,
		HeadSHA:          headSHA,
		BaseBranch:       run.BaseBranch,
		ErrorMessage:     run.ErrorMessage,
		NoConfigChanges:  run.NoConfigChanges,
		CreatedAt:        run.CreatedAt.String(),
	}
	if run.Preview != nil {
		out.PreviewInstallID = run.Preview.InstallID
		out.PreviewInstall = run.Preview.InstallName
		out.PreviewMode = string(run.Preview.Mode)
	}
	if run.WorkflowID != nil {
		out.WorkflowID = *run.WorkflowID
		prog, err := s.appBranchWorkflowProgress(ctx, *run.WorkflowID)
		if err != nil {
			return nil, err
		}
		out.Workflow = prog
	}

	if run.Comparison != nil {
		out.Comparison = &mcpAppBranchChangeSummary{
			HeadSHA:         headSHA,
			NoConfigChanges: run.NoConfigChanges,
			HasGitDiff:      run.Comparison.GitDiff != nil && run.Comparison.GitDiff.IsSet(),
			HasConfigDiff:   run.Comparison.ConfigDiff != nil && run.Comparison.ConfigDiff.IsSet(),
		}
		if run.Comparison.BaseRun != nil {
			out.Comparison.BaseSHA = comparisonRunSHA(run.Comparison.BaseRun)
		}
		s.fillComparisonDetails(ctx, run.Comparison, out.Comparison)
	}

	groups, err := s.appBranchRunDeployment(ctx, run.ID)
	if err != nil {
		return nil, err
	}
	out.Deployment = groups
	return out, nil
}

func (s *service) appBranchWorkflowProgress(ctx context.Context, workflowID string) (*mcpAppBranchWorkflowProg, error) {
	var wf app.Workflow
	err := s.db.WithContext(ctx).
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_idx, group_retry_idx, idx, created_at asc")
		}).
		First(&wf, "id = ?", workflowID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get branch run workflow: %w", err)
	}

	prog := &mcpAppBranchWorkflowProg{
		Status:            string(wf.Status.Status),
		StatusDescription: wf.Status.StatusHumanDescription,
	}
	for _, step := range wf.Steps {
		prog.TotalSteps++
		if step.Status.Status == app.StatusSuccess {
			prog.CompletedSteps++
		} else if prog.CurrentStep == "" {
			prog.CurrentStep = step.Name
		}
	}
	return prog, nil
}

func (s *service) appBranchRunDeployment(ctx context.Context, runID string) ([]mcpAppBranchDeployGroup, error) {
	var groupRuns []app.InstallGroupRun
	res := s.db.WithContext(ctx).
		Where(app.InstallGroupRun{AppBranchRunID: runID}).
		Order("created_at ASC").
		Find(&groupRuns)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install group runs: %w", res.Error)
	}

	nameByID := s.installNamesByID(ctx, collectGroupRunInstallIDs(groupRuns))

	out := make([]mcpAppBranchDeployGroup, 0, len(groupRuns))
	for _, g := range groupRuns {
		item := mcpAppBranchDeployGroup{
			GroupName:         g.InstallGroupName,
			Status:            string(g.Status.Status),
			StatusDescription: g.Status.StatusHumanDescription,
			TotalInstalls:     g.TotalInstalls,
			CompletedInstalls: g.CompletedInstalls,
			FailedInstalls:    g.FailedInstalls,
		}
		for _, inst := range g.Installs {
			item.Installs = append(item.Installs, mcpAppBranchDeployInstall{
				InstallID:   inst.InstallID,
				InstallName: nameByID[inst.InstallID],
				Status:      inst.Status,
				Phase:       inst.Phase,
				WorkflowID:  inst.WorkflowID,
			})
		}
		out = append(out, item)
	}
	return out, nil
}

const mcpDiffPathLimit = 15
const mcpDiffEntryLimit = 12

func (s *service) fillComparisonDetails(ctx context.Context, comparison *app.AppBranchRunComparison, summary *mcpAppBranchChangeSummary) {
	if comparison == nil || summary == nil || s.blobSvc == nil {
		return
	}
	blobCtx := blobstore.WithBlobService(ctx, s.blobSvc)

	if comparison.GitDiff != nil && comparison.GitDiff.IsSet() {
		var git struct {
			ChangedPaths []string `json:"changed_paths"`
			FilesChanged int      `json:"files_changed"`
		}
		if err := loadBlobInto(blobCtx, comparison.GitDiff, &git); err == nil {
			summary.GitFilesChanged = git.FilesChanged
			if summary.GitFilesChanged == 0 {
				summary.GitFilesChanged = len(git.ChangedPaths)
			}
			summary.GitPaths = capStringSlice(git.ChangedPaths, mcpDiffPathLimit)
		}
	}

	if comparison.ConfigDiff != nil && comparison.ConfigDiff.IsSet() {
		var cfg struct {
			ConfigFile string `json:"config_file"`
			Additions  int    `json:"additions"`
			Removals   int    `json:"removals"`
			Changed    int    `json:"changed"`
			Sections   []struct {
				Name      string `json:"name"`
				Additions int    `json:"additions"`
				Removals  int    `json:"removals"`
				Changed   int    `json:"changed"`
				Entries   []struct {
					Op   string `json:"op"`
					Name string `json:"name"`
				} `json:"entries"`
			} `json:"sections"`
		}
		if err := loadBlobInto(blobCtx, comparison.ConfigDiff, &cfg); err == nil {
			summary.ConfigFile = cfg.ConfigFile
			summary.ConfigAdditions = cfg.Additions
			summary.ConfigRemovals = cfg.Removals
			summary.ConfigChanged = cfg.Changed
			for _, sec := range cfg.Sections {
				section := mcpAppBranchConfigSection{
					Name:      sec.Name,
					Additions: sec.Additions,
					Removals:  sec.Removals,
					Changed:   sec.Changed,
				}
				for i, e := range sec.Entries {
					if i >= mcpDiffEntryLimit {
						break
					}
					label := e.Name
					if e.Op != "" && e.Name != "" {
						label = e.Op + " " + e.Name
					}
					if label != "" {
						section.Entries = append(section.Entries, label)
					}
				}
				summary.ConfigSections = append(summary.ConfigSections, section)
			}
		}
	}
}

func loadBlobInto(ctx context.Context, blob *blobstore.Blob, dest any) error {
	raw, err := blob.Get(ctx)
	if err != nil {
		return err
	}
	if raw == "" {
		return fmt.Errorf("empty blob")
	}
	return json.Unmarshal([]byte(raw), dest)
}

func collectGroupRunInstallIDs(groups []app.InstallGroupRun) []string {
	seen := map[string]struct{}{}
	var ids []string
	for _, g := range groups {
		for _, inst := range g.Installs {
			if inst.InstallID == "" {
				continue
			}
			if _, ok := seen[inst.InstallID]; ok {
				continue
			}
			seen[inst.InstallID] = struct{}{}
			ids = append(ids, inst.InstallID)
		}
	}
	return ids
}

func (s *service) installNamesByID(ctx context.Context, ids []string) map[string]string {
	out := map[string]string{}
	if len(ids) == 0 {
		return out
	}
	var installs []app.Install
	if err := s.db.WithContext(ctx).Select("id", "name").Where("id IN ?", ids).Find(&installs).Error; err != nil {
		return out
	}
	for _, inst := range installs {
		out[inst.ID] = inst.Name
	}
	return out
}

func capStringSlice(in []string, n int) []string {
	if len(in) <= n {
		return in
	}
	return in[:n]
}

func mcpBranchOverviewAnswers(result mcpGetAppBranchResult) mcpAppBranchAnswers {
	if result.LatestRun == nil {
		return mcpAppBranchAnswers{
			ChangeSummary:     "no branch runs yet",
			DeploymentSummary: "no branch runs yet",
		}
	}
	run := result.LatestRun
	succeeded := run.Succeeded
	answers := mcpAppBranchAnswers{
		LastRunSucceeded: &succeeded,
		LastRunStatus:    run.Status,
	}
	if run.AwaitingApproval {
		answers.LastRunStatus = run.Status + " (awaiting approval)"
	}

	switch {
	case run.NoConfigChanges:
		answers.ChangeSummary = "no config changes in the latest run"
	case run.Comparison == nil:
		answers.ChangeSummary = "latest run has no comparison yet"
	default:
		parts := []string{}
		if run.Comparison.GitFilesChanged > 0 {
			parts = append(parts, fmt.Sprintf("%d git files changed", run.Comparison.GitFilesChanged))
		} else if run.Comparison.HasGitDiff {
			parts = append(parts, "git diff present")
		}
		if run.Comparison.ConfigAdditions+run.Comparison.ConfigRemovals+run.Comparison.ConfigChanged > 0 {
			parts = append(parts, fmt.Sprintf("config +%d / -%d / ~%d", run.Comparison.ConfigAdditions, run.Comparison.ConfigRemovals, run.Comparison.ConfigChanged))
		} else if run.Comparison.HasConfigDiff {
			parts = append(parts, "config diff present")
		}
		if len(run.Comparison.ConfigSections) > 0 {
			names := make([]string, 0, len(run.Comparison.ConfigSections))
			for _, sec := range run.Comparison.ConfigSections {
				if sec.Name != "" {
					names = append(names, sec.Name)
				}
			}
			if len(names) > 0 {
				parts = append(parts, "sections: "+strings.Join(names, ", "))
			}
		}
		if len(parts) == 0 {
			answers.ChangeSummary = "comparison exists but diffs are empty or not stored yet"
		} else {
			answers.ChangeSummary = strings.Join(parts, "; ")
		}
	}

	if run.PreviewInstall != "" {
		answers.ChangeSummary += "; preview against install " + run.PreviewInstall
	}

	if len(run.Deployment) == 0 {
		if run.Workflow != nil && run.Workflow.TotalSteps > 0 {
			answers.DeploymentSummary = fmt.Sprintf("workflow %s (%d/%d steps", run.Workflow.Status, run.Workflow.CompletedSteps, run.Workflow.TotalSteps)
			if run.Workflow.CurrentStep != "" {
				answers.DeploymentSummary += ", current: " + run.Workflow.CurrentStep
			}
			answers.DeploymentSummary += "); install-group deploys have not started"
		} else {
			answers.DeploymentSummary = "install-group deploys have not started"
		}
		return answers
	}

	var groupBits []string
	for _, g := range run.Deployment {
		bit := fmt.Sprintf("%s %d/%d complete", g.GroupName, g.CompletedInstalls, g.TotalInstalls)
		if g.FailedInstalls > 0 {
			bit += fmt.Sprintf(" (%d failed)", g.FailedInstalls)
		}
		if g.Status != "" {
			bit += " [" + g.Status + "]"
		}
		groupBits = append(groupBits, bit)
	}
	answers.DeploymentSummary = strings.Join(groupBits, "; ")
	return answers
}
