package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v50/github"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/links"
)

// GitHub rejects a commit status whose context exceeds 255 characters.
const maxCommitStatusContextLen = 255

type SetGithubCommitStatusInput struct {
	VcsConfigID string `json:"vcs_config_id" validate:"required"`
	CommitSHA   string `json:"commit_sha" validate:"required"`
	State       string `json:"state" validate:"required"`
	Description string `json:"description"`

	// AppBranchID identifies the branch whose org/app/branch names form the
	// status context and whose dashboard link becomes the status target URL.
	AppBranchID string `json:"app_branch_id"`
	RunID       string `json:"run_id"`

	// Preview suffixes the derived context so preview runs get their own check
	// rather than overwriting the branch's regular one.
	Preview     bool                        `json:"preview,omitempty"`
	PreviewMode app.AppBranchRunPreviewMode `json:"preview_mode,omitempty"`

	// Context overrides the derived nuon/{org}/{app}/{branch} context. Left
	// empty by callers that want the branch-derived one.
	Context string `json:"context,omitempty"`

	// TargetURL overrides the derived run link.
	TargetURL string `json:"target_url,omitempty"`
}

// CommitStatusContext is the check name GitHub displays. Scoping it to the
// org, app and branch keeps two branches of the same app from overwriting each
// other's status on a shared commit, and the preview suffix keeps a preview
// from overwriting the branch's regular status.
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
