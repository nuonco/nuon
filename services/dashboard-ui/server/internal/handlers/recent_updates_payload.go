package handlers

import (
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type recentUpdatesPayload struct {
	Updates []recentUpdateSource `json:"updates"`
}

type recentUpdateSource struct {
	AppID            string                   `json:"appId"`
	AppName          string                   `json:"appName"`
	BranchID         string                   `json:"branchId"`
	BranchName       string                   `json:"branchName"`
	RunID            string                   `json:"runId"`
	WorkflowID       string                   `json:"workflowId,omitempty"`
	Status           string                   `json:"status,omitempty"`
	CreatedAt        string                   `json:"createdAt,omitempty"`
	AwaitingApproval bool                     `json:"awaitingApproval,omitempty"`
	CommitMessage    string                   `json:"commitMessage,omitempty"`
	CommitSha        string                   `json:"commitSha,omitempty"`
	CommitAuthor     string                   `json:"commitAuthor,omitempty"`
	CommitAvatarURL  string                   `json:"commitAvatarUrl,omitempty"`
	Repo             string                   `json:"repo,omitempty"`
	ConfigGroups     []recentConfigGroup      `json:"configGroups"`
	GroupRuns        []recentGroupRun         `json:"groupRuns"`
	InstallsByID     map[string]recentInstall `json:"installsById"`
	Approvals        []recentApproval         `json:"approvals"`
}

type recentConfigGroup struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Order       int64  `json:"order,omitempty"`
	HasSelector bool   `json:"hasSelector,omitempty"`
}

type recentGroupRun struct {
	InstallGroupID   string               `json:"install_group_id,omitempty"`
	InstallGroupName string               `json:"install_group_name,omitempty"`
	TotalInstalls    int64                `json:"total_installs,omitempty"`
	InstallGroup     *recentInstallGroup  `json:"install_group,omitempty"`
	Installs         []recentGroupInstall `json:"installs,omitempty"`
}

type recentInstallGroup struct {
	Name          string `json:"name,omitempty"`
	Default       bool   `json:"default,omitempty"`
	LabelSelector any    `json:"label_selector,omitempty"`
}

type recentGroupInstall struct {
	InstallID  string `json:"install_id,omitempty"`
	Status     string `json:"status,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`
}

type recentInstall struct {
	Name   string `json:"name,omitempty"`
	Health string `json:"health,omitempty"`
}

type recentApproval struct {
	ID           string              `json:"id,omitempty"`
	Type         string              `json:"type,omitempty"`
	OwnerID      string              `json:"owner_id,omitempty"`
	OwnerType    string              `json:"owner_type,omitempty"`
	WorkflowStep *recentApprovalStep `json:"workflow_step,omitempty"`
}

type recentApprovalStep struct {
	InstallWorkflowID string `json:"install_workflow_id,omitempty"`
	OwnerID           string `json:"owner_id,omitempty"`
	OwnerType         string `json:"owner_type,omitempty"`
}

func buildRecentUpdate(
	ref recentRunRef,
	detail *models.AppWorkflow,
	config *models.AppAppBranchConfig,
	app *models.AppApp,
	installs []*models.AppInstall,
	groupRuns []*models.AppInstallGroupRun,
	approvals []*models.AppWorkflowStepApproval,
) recentUpdateSource {
	run := ref.run
	if detail != nil && len(detail.AppBranchRuns) > 0 && detail.AppBranchRuns[0] != nil {
		run = detail.AppBranchRuns[0]
	}

	appName := ref.appID
	if app != nil && app.Name != "" {
		appName = app.Name
	}

	source := recentUpdateSource{
		AppID:        ref.appID,
		AppName:      appName,
		BranchID:     ref.branchID,
		BranchName:   ref.branchName,
		RunID:        ref.run.ID,
		WorkflowID:   ref.workflow.ID,
		Status:       runStatus(run, ref),
		CreatedAt:    recentRunCreatedAt(ref),
		ConfigGroups: configGroups(config),
		GroupRuns:    groupRunSources(groupRuns),
		InstallsByID: installSources(installs),
		Approvals:    approvalSources(approvals),
	}
	if run != nil {
		source.AwaitingApproval = run.AwaitingApproval
		source.CommitSha = run.HeadSha
		if run.Metadata != nil && run.Metadata.HeadSha != "" {
			source.CommitSha = run.Metadata.HeadSha
		}
		if run.VcsConnectionCommit != nil {
			source.CommitMessage = run.VcsConnectionCommit.Message
			source.CommitAuthor = run.VcsConnectionCommit.AuthorName
			source.CommitAvatarURL = run.VcsConnectionCommit.AuthorAvatarURL
			if run.VcsConnectionCommit.Sha != "" {
				source.CommitSha = run.VcsConnectionCommit.Sha
			}
		}
	}
	source.Repo = repoSlug(config, run)
	return source
}

func runStatus(run *models.AppAppBranchRun, ref recentRunRef) string {
	if run != nil && run.Status != "" {
		return run.Status
	}
	return recentRunStatus(ref)
}

func repoSlug(config *models.AppAppBranchConfig, run *models.AppAppBranchRun) string {
	if run != nil && run.AppBranchConfig != nil {
		if repo := configRepo(run.AppBranchConfig); repo != "" {
			return repo
		}
	}
	return configRepo(config)
}

func configRepo(config *models.AppAppBranchConfig) string {
	if config == nil {
		return ""
	}
	if config.ConnectedGithubVcsConfig != nil && config.ConnectedGithubVcsConfig.Repo != "" {
		return config.ConnectedGithubVcsConfig.Repo
	}
	if config.PublicGitVcsConfig != nil {
		return config.PublicGitVcsConfig.Repo
	}
	return ""
}

func configGroups(config *models.AppAppBranchConfig) []recentConfigGroup {
	if config == nil {
		return []recentConfigGroup{}
	}
	groups := append([]*models.AppAppBranchInstallGroup{}, config.InstallGroups...)
	for i := 1; i < len(groups); i++ {
		j := i
		for j > 0 && groups[j] != nil && groups[j-1] != nil && groups[j].Order < groups[j-1].Order {
			groups[j], groups[j-1] = groups[j-1], groups[j]
			j--
		}
	}
	out := make([]recentConfigGroup, 0, len(groups))
	for _, group := range groups {
		if group == nil {
			continue
		}
		out = append(out, recentConfigGroup{
			ID:          group.ID,
			Name:        group.Name,
			Order:       group.Order,
			HasSelector: group.LabelSelector != nil || group.Default,
		})
	}
	return out
}

func groupRunSources(groupRuns []*models.AppInstallGroupRun) []recentGroupRun {
	out := make([]recentGroupRun, 0, len(groupRuns))
	for _, groupRun := range groupRuns {
		if groupRun == nil {
			continue
		}
		installs := make([]recentGroupInstall, 0, len(groupRun.Installs))
		for _, install := range groupRun.Installs {
			if install == nil {
				continue
			}
			installs = append(installs, recentGroupInstall{
				InstallID:  install.InstallID,
				Status:     install.Status,
				WorkflowID: install.WorkflowID,
			})
		}
		source := recentGroupRun{
			InstallGroupID:   groupRun.InstallGroupID,
			InstallGroupName: groupRun.InstallGroupName,
			TotalInstalls:    groupRun.TotalInstalls,
			Installs:         installs,
		}
		if groupRun.InstallGroup != nil {
			var selector any
			if groupRun.InstallGroup.LabelSelector != nil {
				selector = groupRun.InstallGroup.LabelSelector
			}
			source.InstallGroup = &recentInstallGroup{
				Name:          groupRun.InstallGroup.Name,
				Default:       groupRun.InstallGroup.Default,
				LabelSelector: selector,
			}
		}
		out = append(out, source)
	}
	return out
}

func installSources(installs []*models.AppInstall) map[string]recentInstall {
	out := map[string]recentInstall{}
	for _, install := range installs {
		if install == nil || install.ID == "" {
			continue
		}
		out[install.ID] = recentInstall{
			Name:   install.Name,
			Health: install.CompositeHealthStatus,
		}
	}
	return out
}

func approvalSources(approvals []*models.AppWorkflowStepApproval) []recentApproval {
	out := make([]recentApproval, 0, len(approvals))
	for _, approval := range approvals {
		if approval == nil {
			continue
		}
		source := recentApproval{
			ID:        approval.ID,
			Type:      string(approval.Type),
			OwnerID:   approval.OwnerID,
			OwnerType: approval.OwnerType,
		}
		step := approval.WorkflowStep
		if step == nil {
			step = approval.InstallWorkflowStep
		}
		if step != nil {
			source.WorkflowStep = &recentApprovalStep{
				InstallWorkflowID: step.InstallWorkflowID,
				OwnerID:           step.OwnerID,
				OwnerType:         step.OwnerType,
			}
		}
		out = append(out, source)
	}
	return out
}
