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

	previewConfigStepNameFetch = "fetch app config"
	previewConfigStepNameSync  = "sync app config"

	previewInstallStepNamePlan   = "plan preview install"
	previewInstallStepNameApply  = "apply preview install"
	previewInstallStepNameImpact = "preview install impact"
)

type GetPreviewCommentContextInput struct {
	RunID string `json:"run_id" validate:"required"`
}

type GetPreviewCommentContextOutput struct {
	RunURL             string                 `json:"run_url"`
	ComponentChanges   []ComponentBuildChange `json:"component_changes"`
	PreviewInstallName string                 `json:"preview_install_name"`
	PreviewInstallURL  string                 `json:"preview_install_url"`
	// Diff holds the config diff computed from the run's app config IDs.
	// Nil when the run has no app config or the diff cannot be computed.
	Diff *ComputeAppConfigDiffOutput `json:"diff,omitempty"`
	// InstallImpact holds install impact groups parsed from the impact step metadata.
	// Nil when the impact step has not run or has no groups.
	InstallImpact []InstallGroupImpact `json:"install_impact,omitempty"`
	// Phases holds per-phase check statuses derived from workflow step records.
	// Nil when the run has no workflow ID or the steps cannot be queried.
	Phases *PRCommentPhases `json:"phases,omitempty"`
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

		var steps []app.WorkflowStep
		a.db.WithContext(ctx).
			Where(app.WorkflowStep{InstallWorkflowID: *run.WorkflowID}).
			Find(&steps)

		out.Phases = deriveCommentPhases(run.PreviewMode(), steps, run.AppConfigID)

		for _, step := range steps {
			if step.Name == previewBuildsStepName {
				sandboxBuildID, _ := step.Status.Metadata["sandbox_build_id"].(string)
				out.ComponentChanges = componentChangesFromMetadata(
					step.Status.Metadata,
					a.cfg.AppURL,
					run.OrgID,
					run.AppBranch.AppID,
					sandboxBuildID,
				)
			}
			if step.Name == previewInstallStepNameImpact {
				out.InstallImpact = installImpactFromStepMetadata(step.Status.Metadata)
			}
		}
	}

	out.Diff = a.computeCommentDiff(ctx, &run)

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

// deriveCommentPhases maps workflow step statuses to PR comment phase statuses.
// It is called from the GetPreviewCommentContext activity so the state is always
// read from the database — any signal that rewrites the comment can call the
// same activity and get a consistent view of all phases.
func deriveCommentPhases(mode app.AppBranchRunPreviewMode, steps []app.WorkflowStep, appConfigID string) *PRCommentPhases {
	byName := make(map[string]*app.WorkflowStep, len(steps))
	for i := range steps {
		s := &steps[i]
		byName[s.Name] = s
	}

	phases := &PRCommentPhases{}

	// Config phase
	if s, ok := findStep(byName, previewConfigStepNameFetch, previewConfigStepNameSync); ok {
		phases.Config = stepToPhaseStatus(s.Status.Status, "config")
	} else if appConfigID != "" {
		// Pre-existing config: no fetch/sync step, but config is implicitly valid.
		phases.Config = PRCommentPhaseValid
	}

	// Builds phase
	if s, ok := findStep(byName, previewBuildsStepName); ok {
		phases.Builds = stepToPhaseStatus(s.Status.Status, "builds")
	}

	// Install phase — only shown for non-build-only modes.
	if mode != app.AppBranchRunPreviewModeBuildOnly {
		if s, ok := findStep(byName, previewInstallStepNamePlan, previewInstallStepNameApply, previewInstallStepNameImpact); ok {
			phases.Install = stepToPhaseStatus(s.Status.Status, "install")
		}
	}

	return phases
}

func findStep(byName map[string]*app.WorkflowStep, names ...string) (*app.WorkflowStep, bool) {
	for _, name := range names {
		if s, ok := byName[name]; ok {
			return s, true
		}
	}
	return nil, false
}

// stepToPhaseStatus maps a workflow step CompositeStatus to a PR comment phase status.
// The kind parameter ("config", "builds", "install") picks the right in-progress label.
func stepToPhaseStatus(status app.Status, kind string) PRCommentPhaseStatus {
	switch status {
	case app.StatusSuccess:
		return PRCommentPhaseValid
	case app.StatusError, app.StatusCancelled:
		return PRCommentPhaseInvalid
	default:
		switch kind {
		case "config":
			return PRCommentPhaseValidating
		case "builds":
			return PRCommentPhaseBuilding
		case "install":
			return PRCommentPhaseConfiguring
		}
		return PRCommentPhaseValidating
	}
}

// computeCommentDiff resolves the baseline config and computes the app config diff.
// Returns nil when there is no app config on the run or any step fails.
func (a *Activities) computeCommentDiff(ctx context.Context, run *app.AppBranchRun) *ComputeAppConfigDiffOutput {
	if run.AppConfigID == "" || run.AppBranch.AppID == "" {
		return nil
	}

	var oldConfigID string
	if baseline, err := a.helpers.ResolvePreviewBaselineAppConfig(ctx, run.ID, run.AppBranchID); err == nil && baseline != nil {
		oldConfigID = baseline.AppConfigID
	}

	result, err := a.ComputeAppConfigDiff(ctx, &ComputeAppConfigDiffInput{
		AppID:       run.AppBranch.AppID,
		NewConfigID: run.AppConfigID,
		OldConfigID: oldConfigID,
	})
	if err != nil {
		return nil
	}
	return result
}

// installImpactFromStepMetadata parses InstallGroupImpact from the metadata
// written by previewimpact.updateStepMetadata.
func installImpactFromStepMetadata(metadata map[string]any) []InstallGroupImpact {
	raw, ok := metadata["install_groups"].([]any)
	if !ok || len(raw) == 0 {
		return nil
	}

	groups := make([]InstallGroupImpact, 0, len(raw))
	for _, gv := range raw {
		gm, ok := gv.(map[string]any)
		if !ok {
			continue
		}
		groupName, _ := gm["install_group_name"].(string)
		rawInstalls, _ := gm["installs"].([]any)

		installs := make([]InstallImpact, 0, len(rawInstalls))
		for _, iv := range rawInstalls {
			im, ok := iv.(map[string]any)
			if !ok {
				continue
			}
			installs = append(installs, InstallImpact{
				InstallID:      metaString(im, "install_id"),
				InstallName:    metaString(im, "install_name"),
				Added:          metaInt(im, "added"),
				Changed:        metaInt(im, "changed"),
				Removed:        metaInt(im, "removed"),
				Unchanged:      metaInt(im, "unchanged"),
				SandboxChanged: metaBool(im, "sandbox_changed"),
				StackChanged:   metaBool(im, "stack_changed"),
			})
		}

		if groupName != "" {
			groups = append(groups, InstallGroupImpact{GroupName: groupName, Installs: installs})
		}
	}
	return groups
}

func metaString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func metaInt(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	}
	return 0
}

func metaBool(m map[string]any, key string) bool {
	v, _ := m[key].(bool)
	return v
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
