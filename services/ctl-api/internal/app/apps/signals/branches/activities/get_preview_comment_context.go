package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/links"
)

// This must match the step name emitted by the app branch run workflow.
const (
	previewBuildsStepName = "building components and sandbox"
	SandboxComponentID    = "sandbox"
)

type GetPreviewCommentContextInput struct {
	RunID string `json:"run_id" validate:"required"`
}

type GetPreviewCommentContextOutput struct {
	RunURL             string                 `json:"run_url"`
	ComponentChanges   []ComponentBuildChange `json:"component_changes"`
	PreviewInstallName string                 `json:"preview_install_name"`
	PreviewInstallURL  string                 `json:"preview_install_url"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) GetPreviewCommentContext(ctx context.Context, input *GetPreviewCommentContextInput) (*GetPreviewCommentContextOutput, error) {
	var run app.AppBranchRun
	if err := a.db.WithContext(ctx).
		Preload("AppBranch").
		Preload("Preview").
		Where(app.AppBranchRun{ID: input.RunID}).
		First(&run).Error; err != nil {
		return nil, fmt.Errorf("unable to load app branch run: %w", err)
	}

	out := &GetPreviewCommentContextOutput{}
	if run.WorkflowID != nil {
		out.RunURL = links.AppBranchRunUILink(
			a.cfg.AppURL,
			run.OrgID,
			run.AppBranch.AppID,
			run.AppBranchID,
			*run.WorkflowID,
		)

		var step app.WorkflowStep
		if err := a.db.WithContext(ctx).
			Where(app.WorkflowStep{
				InstallWorkflowID: *run.WorkflowID,
				Name:              previewBuildsStepName,
			}).
			Order("created_at DESC").
			First(&step).Error; err == nil {
			sandboxBuildID, _ := step.Status.Metadata["sandbox_build_id"].(string)
			out.ComponentChanges = componentChangesFromMetadata(
				step.Status.Metadata,
				a.cfg.AppURL,
				run.OrgID,
				run.AppBranch.AppID,
				sandboxBuildID,
			)
		}
	}

	if run.Preview != nil && run.Preview.InstallID != "" {
		out.PreviewInstallURL = links.InstallAppBranchRunsUILink(
			a.cfg.AppURL,
			run.OrgID,
			run.Preview.InstallID,
		)
		var install app.Install
		if err := a.db.WithContext(ctx).
			Select("name").
			Where(app.Install{ID: run.Preview.InstallID}).
			First(&install).Error; err == nil {
			out.PreviewInstallName = install.Name
		}
	}

	return out, nil
}

func componentChangesFromMetadata(metadata map[string]any, appURL, orgID, appID, sandboxBuildID string) []ComponentBuildChange {
	raw, ok := metadata["builds"].([]any)
	if !ok {
		return nil
	}

	changes := make([]ComponentBuildChange, 0, len(raw))
	for _, value := range raw {
		entry, ok := value.(map[string]any)
		if !ok {
			continue
		}
		name, _ := entry["component_name"].(string)
		componentID, _ := entry["component_id"].(string)
		buildID, _ := entry["build_id"].(string)
		reason, _ := entry["change_reason"].(string)
		if name == "" || (reason != ChangeReasonSourceChanged && reason != ChangeReasonConfigChanged) {
			continue
		}
		buildURL := links.ComponentBuildUILink(appURL, orgID, appID, componentID, buildID)
		if componentID == SandboxComponentID {
			buildID = sandboxBuildID
			buildURL = links.SandboxBuildUILink(appURL, orgID, appID, buildID)
		}
		changes = append(changes, ComponentBuildChange{
			ComponentName: name,
			ComponentID:   componentID,
			BuildID:       buildID,
			ChangeReason:  reason,
			BuildURL:      buildURL,
		})
	}
	return changes
}
