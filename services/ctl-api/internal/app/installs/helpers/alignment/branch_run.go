package alignment

import (
	"strings"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type InstallOverviewCommit struct {
	SHA              string                   `json:"sha,omitempty"`
	RunID            string                   `json:"run_id,omitempty"`
	WorkflowID       string                   `json:"workflow_id,omitempty"`
	BranchID         string                   `json:"branch_id,omitempty"`
	AppConfigID      string                   `json:"app_config_id,omitempty"`
	Message          string                   `json:"message,omitempty"`
	Author           string                   `json:"author,omitempty"`
	CreatedAt        *time.Time               `json:"created_at,omitempty"`
	RunStatus        string                   `json:"run_status,omitempty"`
	AwaitingApproval bool                     `json:"awaiting_approval,omitempty"`
	Metadata         app.AppBranchRunMetadata `json:"metadata,omitzero"`
} // @name service.InstallOverviewCommit

func CommitFromRun(run *app.AppBranchRun) *InstallOverviewCommit {
	if run == nil {
		return nil
	}
	commit := &InstallOverviewCommit{
		RunID:       run.ID,
		BranchID:    run.AppBranchID,
		AppConfigID: run.AppConfigID,
		RunStatus:   run.Status,
		SHA:         run.RunMetadata().HeadSHA,
		Metadata:    run.RunMetadata(),
	}
	if run.WorkflowID != nil {
		commit.WorkflowID = *run.WorkflowID
	}
	if commit.SHA == "" {
		commit.SHA = run.HeadSHA
	}
	created := run.CreatedAt
	commit.CreatedAt = &created
	if run.VCSConnectionCommit != nil {
		commit.Author = run.VCSConnectionCommit.AuthorName
		commit.Message = firstLine(run.VCSConnectionCommit.Message)
		if commit.SHA == "" {
			commit.SHA = run.VCSConnectionCommit.SHA
		}
	}
	return commit
}

func firstLine(message string) string {
	message = strings.TrimSpace(message)
	if i := strings.IndexByte(message, '\n'); i >= 0 {
		return message[:i]
	}
	return message
}
