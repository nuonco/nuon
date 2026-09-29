package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/links"
)

const maxCommitStatusContextLen = 255

type SetGithubCommitStatusInput struct {
	VcsConfigID string `json:"vcs_config_id" validate:"required"`
	CommitSHA   string `json:"commit_sha" validate:"required"`
	State       string `json:"state" validate:"required"`
	Description string `json:"description"`

	AppBranchID string `json:"app_branch_id"`
	RunID       string `json:"run_id"`

	Preview     bool                        `json:"preview,omitempty"`
	PreviewMode app.AppBranchRunPreviewMode `json:"preview_mode,omitempty"`

	Context string `json:"context,omitempty"`

	TargetURL string `json:"target_url,omitempty"`
}

func CommitStatusContext(orgName, appName, branchName string, preview bool, previewMode app.AppBranchRunPreviewMode) string {
	parts := make([]string, 0, 4)
	parts = append(parts, "nuon")
	for _, p := range []string{orgName, appName, branchName} {
		if p != "" {
			parts = append(parts, p)
		}
	}

	ctxStr := strings.Join(parts, "/")
	if preview {
		ctxStr += " preview"
		if label := previewMode.Label(); label != "" {
			ctxStr += fmt.Sprintf(" (%s)", label)
		}
	}

	runes := []rune(ctxStr)
	if len(runes) > maxCommitStatusContextLen {
		ctxStr = string(runes[:maxCommitStatusContextLen])
	}

	return ctxStr
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) SetGithubCommitStatus(ctx context.Context, input *SetGithubCommitStatusInput) error {
	owner, repo, client, err := a.resolveAuthenticatedGithubClient(ctx, input.VcsConfigID)
	if err != nil {
		a.l.Info("skipping commit status: no authenticated GitHub client available")
		return nil
	}

	statusContext, targetURL := input.Context, input.TargetURL
	if (statusContext == "" || targetURL == "") && input.AppBranchID != "" {
		var branch app.AppBranch
		if res := a.db.WithContext(ctx).
			Preload("Org").
			Preload("App").
			Where(app.AppBranch{ID: input.AppBranchID}).
			First(&branch); res.Error == nil {
			if statusContext == "" {
				statusContext = CommitStatusContext(branch.Org.Name, branch.App.Name, branch.Name, input.Preview, input.PreviewMode)
			}
			if targetURL == "" {
				workflowID := ""
				var run app.AppBranchRun
				if runRes := a.db.WithContext(ctx).
					Select("workflow_id").
					Where(app.AppBranchRun{ID: input.RunID}).
					First(&run); runRes.Error == nil && run.WorkflowID != nil {
					workflowID = *run.WorkflowID
				}
				targetURL = links.AppBranchRunUILink(a.cfg.AppURL, branch.OrgID, branch.AppID, branch.ID, workflowID)
			}
		} else {
			a.l.Warn("unable to load app branch for commit status context")
		}
	}
	if statusContext == "" {
		statusContext = CommitStatusContext("", "", "", input.Preview, input.PreviewMode)
	}

	status := &github.RepoStatus{
		State:       &input.State,
		Context:     &statusContext,
		Description: &input.Description,
	}
	if targetURL != "" {
		status.TargetURL = &targetURL
	}

	_, _, err = client.Repositories.CreateStatus(ctx, owner, repo, input.CommitSHA, status)
	if err != nil {
		if nrErr := nonRetryableGitHubError(err); nrErr != nil {
			return nrErr
		}
		return fmt.Errorf("unable to create commit status: %w", err)
	}

	return nil
}
