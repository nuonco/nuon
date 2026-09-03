package activities

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBuildPRCommentBodyIncludesInstallImpact(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrq7fplr1up5atx5zpxotbabm",
		Status:  PRCommentStatusSuccess,
		InstallImpact: []InstallGroupImpact{
			{
				GroupName: "canary",
				Installs: []InstallImpact{
					{InstallID: "insta", InstallName: "canary-1", Added: 1, Changed: 2, StackChanged: true},
				},
			},
			{
				GroupName: "prod",
				Installs: []InstallImpact{
					{InstallID: "instb", InstallName: "prod-1", Unchanged: 5},
				},
			},
		},
	})

	for _, want := range []string{
		"Install Impact — 2 install(s)",
		"nothing was applied",
		"canary",
		"canary-1",
		"prod-1",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("comment body missing %q\n%s", want, body)
		}
	}
}

func TestBuildPRCommentBodyOmitsEmptyInstallImpact(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrq7fplr1up5atx5zpxotbabm",
		Status:  PRCommentStatusSuccess,
	})

	if strings.Contains(body, "Install Impact") {
		t.Errorf("expected no install impact section when there are no installs\n%s", body)
	}
}

// A no-changes preview short-circuits before any install is resolved, so the
// impact section would be misleading there.
func TestBuildPRCommentBodySkippedHasNoInstallImpact(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrq7fplr1up5atx5zpxotbabm",
		Status:  PRCommentStatusSkipped,
		InstallImpact: []InstallGroupImpact{
			{GroupName: "prod", Installs: []InstallImpact{{InstallID: "instb", InstallName: "prod-1"}}},
		},
	})

	if strings.Contains(body, "Install Impact") {
		t.Errorf("skipped preview should not render install impact\n%s", body)
	}
}

func TestBuildPRCommentBodyIncludesModeRunLinkBuildLabelsAndStackWarning(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		OrgName:    "acme",
		AppName:    "payments",
		BranchName: "production",
		RunID:      "abrun-example",
		RunURL:     "https://app.example.com/org/apps/app/branches/branch/runs/workflow",
		Status:     PRCommentStatusSuccess,
		Mode:       app.AppBranchRunPreviewModeBuildOnly,
		Diff: &ComputeAppConfigDiffOutput{
			ConfigFile: "nuon.toml",
			Sections: []ConfigDiffSection{
				{Name: "Stack", Changed: 1},
			},
		},
		ComponentChanges: []ComponentBuildChange{
			{
				ComponentName: "api",
				ChangeReason:  ChangeReasonSourceChanged,
				BuildURL:      "https://app.example.com/org/apps/app/components/api/builds/build-api",
			},
			{ComponentName: "worker", ChangeReason: ChangeReasonConfigChanged},
			{ComponentName: "unchanged", ChangeReason: ChangeReasonNoChanges},
		},
	})

	for _, want := range []string{
		"## Nuon Preview \u2014 acme/payments/production (build and validate)",
		"[View preview run \u2192](https://app.example.com/org/apps/app/branches/branch/runs/workflow)",
		"\U0001f6a8 Stack changes require customers to reprovision the stack. Learn more [here](https://docs.nuon.co/concepts/stacks).",
		"### Builds",
		"| [`api`](https://app.example.com/org/apps/app/components/api/builds/build-api) | `Source changed` |",
		"| `worker` | `Config changed` |",
		"No install was planned or applied.",
		"### Debug with MCP",
		"Fetch the overview of app branch run abrun-example and diagnose any failures.",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("comment body missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "`unchanged`") {
		t.Errorf("comment body should omit unchanged components\n%s", body)
	}
	if strings.Contains(body, "View Run") {
		t.Errorf("comment footer should not contain a second run link\n%s", body)
	}
	if strings.Contains(body, "Updated:") {
		t.Errorf("comment body should not contain a redundant timestamp footer\n%s", body)
	}
}

func TestBuildPRCommentBodyApplyNamesPreviewInstall(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		OrgName:            "acme",
		AppName:            "payments",
		BranchName:         "production",
		RunID:              "abrun-example",
		Status:             PRCommentStatusSuccess,
		Mode:               app.AppBranchRunPreviewModeApply,
		PreviewInstallName: "preview-us-west",
		PreviewInstallURL:  "https://app.example.com/org/installs/install/app-branch-runs",
		InstallApplied:     true,
	})

	if !strings.Contains(body, "## Nuon Preview \u2014 acme/payments/production (apply)") {
		t.Errorf("comment body missing apply mode\n%s", body)
	}
	if !strings.Contains(body, "Applied to [`preview-us-west`](https://app.example.com/org/installs/install/app-branch-runs).") {
		t.Errorf("comment body missing preview install\n%s", body)
	}
	if strings.Contains(body, "Preview only") {
		t.Errorf("apply comment must not claim nothing was applied\n%s", body)
	}
}

func TestBuildPRCommentBodyApplyOmitsInstallWhenNotApplied(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		OrgName:            "acme",
		AppName:            "payments",
		BranchName:         "production",
		RunID:              "abrun-example",
		Status:             PRCommentStatusSuccess,
		Mode:               app.AppBranchRunPreviewModeApply,
		PreviewInstallName: "preview-us-west",
		PreviewInstallURL:  "https://app.example.com/org/installs/install/app-branch-runs",
		// InstallApplied intentionally false
	})

	if strings.Contains(body, "Applied to") {
		t.Errorf("comment body should not include 'Applied to' when InstallApplied is false\n%s", body)
	}
}

func TestBuildPRCommentBodySkippedOmitsMCPPrompt(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "production",
		RunID:   "abrun-example",
		Status:  PRCommentStatusSkipped,
		Mode:    app.AppBranchRunPreviewModePlanOnly,
	})

	if strings.Contains(body, "Debug with MCP") {
		t.Errorf("skipped comment should not contain an MCP debug prompt\n%s", body)
	}
	if strings.Contains(body, "Updated:") {
		t.Errorf("skipped comment should not contain a redundant timestamp footer\n%s", body)
	}
}

func TestBuildPRCommentBodyRendersCollapsedDiff(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		OrgName:    "acme",
		AppName:    "payments",
		BranchName: "production",
		RunID:      "abrun-example",
		Status:     PRCommentStatusPending,
		Mode:       app.AppBranchRunPreviewModePlanOnly,
		Diff: &ComputeAppConfigDiffOutput{
			ConfigFile: "nuon.toml",
			Sections: []ConfigDiffSection{
				{
					Name:      "Components",
					Additions: 1,
					Changed:   1,
					Entries: []ConfigDiffEntry{
						{Name: "api", Op: "add"},
						{Name: "worker", Op: "change", Description: "image tag"},
					},
				},
				{
					Name:     "Installers",
					Removals: 1,
					Entries: []ConfigDiffEntry{
						{Name: "aws-eks", Op: "remove"},
					},
				},
			},
		},
	})

	for _, want := range []string{
		"<details>\n<summary><strong>Config changes</strong> <code>+1</code> <code>~1</code> <code>-1</code></summary>",
		"#### Components",
		"- `+` `api`",
		"- `~` `worker` \u2014 image tag",
		"#### Installers",
		"- `-` `aws-eks`",
		"</details>",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("comment body missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "| Section | Added | Changed | Removed |") {
		t.Errorf("comment body should not render the old per-section table\n%s", body)
	}
}

func TestBuildPRCommentBodyEmptyDiffCollapsesToNoChanges(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		BranchName: "production",
		RunID:      "abrun-example",
		Status:     PRCommentStatusPending,
		Mode:       app.AppBranchRunPreviewModePlanOnly,
		Diff:       &ComputeAppConfigDiffOutput{ConfigFile: "nuon.toml"},
	})

	if !strings.Contains(body, "<summary><strong>Config changes</strong> <code>no changes</code></summary>") {
		t.Errorf("comment body missing empty diff summary\n%s", body)
	}
	if !strings.Contains(body, "## Nuon Preview \u2014 production (plan-only)") {
		t.Errorf("heading should fall back to the branch name alone\n%s", body)
	}
}

func TestComponentChangesFromMetadataBuildsURLs(t *testing.T) {
	changes := componentChangesFromMetadata(map[string]any{
		"builds": []any{
			map[string]any{
				"component_name": "api",
				"component_id":   "component-api",
				"build_id":       "build-api",
				"change_reason":  ChangeReasonSourceChanged,
			},
			map[string]any{
				"component_name": "worker",
				"component_id":   "component-worker",
				"change_reason":  ChangeReasonNoChanges,
			},
			map[string]any{
				"component_name": "Sandbox",
				"component_id":   SandboxComponentID,
				"change_reason":  ChangeReasonSourceChanged,
			},
		},
	}, "https://app.example.com", "org", "app", "sandbox-build")

	if len(changes) != 2 {
		t.Fatalf("expected two changed builds, got %d", len(changes))
	}
	if changes[0].BuildURL != "https://app.example.com/org/apps/app/components/component-api/builds/build-api" {
		t.Errorf("unexpected build URL %q", changes[0].BuildURL)
	}
	if changes[1].BuildURL != "https://app.example.com/org/apps/app/sandbox/builds/sandbox-build" {
		t.Errorf("unexpected sandbox build URL %q", changes[1].BuildURL)
	}
}

func TestBuildPRCommentBodyPhaseChecksTable(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrun-example",
		Status:  PRCommentStatusPending,
		Mode:    app.AppBranchRunPreviewModePlanOnly,
		Phases: &PRCommentPhases{
			Config:  PRCommentPhaseValidating,
			Builds:  PRCommentPhaseBuilding,
			Install: PRCommentPhaseConfiguring,
		},
	})

	if !strings.Contains(body, "| Check | Status |") {
		t.Errorf("phase checks table header missing\n%s", body)
	}
	if !strings.Contains(body, "| Config | \u23f3 Validating |") {
		t.Errorf("config phase row missing\n%s", body)
	}
	if !strings.Contains(body, "| Builds | \u23f3 Building |") {
		t.Errorf("builds phase row missing\n%s", body)
	}
	if !strings.Contains(body, "| Install | \u23f3 Configuring |") {
		t.Errorf("install phase row missing\n%s", body)
	}
}

func TestBuildPRCommentBodyPhaseChecksValidStatuses(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrun-example",
		Status:  PRCommentStatusSuccess,
		Mode:    app.AppBranchRunPreviewModePlanOnly,
		Phases: &PRCommentPhases{
			Config:  PRCommentPhaseValid,
			Builds:  PRCommentPhaseValid,
			Install: PRCommentPhaseValid,
		},
	})

	validCount := strings.Count(body, "\u2705 Valid")
	if validCount != 3 {
		t.Errorf("expected 3 Valid rows, got %d\n%s", validCount, body)
	}
}

func TestBuildPRCommentBodyPhaseChecksInvalidStatus(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrun-example",
		Status:  PRCommentStatusFailed,
		Mode:    app.AppBranchRunPreviewModePlanOnly,
		Phases: &PRCommentPhases{
			Config:  PRCommentPhaseInvalid,
			Builds:  PRCommentPhaseBuilding,
			Install: "",
		},
	})

	if !strings.Contains(body, "| Config | \u274c Invalid |") {
		t.Errorf("config invalid row missing\n%s", body)
	}
	if strings.Contains(body, "| Install |") {
		t.Errorf("install row should be omitted when phase is empty\n%s", body)
	}
}

func TestBuildPRCommentBodyPhaseChecksOmittedWhenNil(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrun-example",
		Status:  PRCommentStatusPending,
		Mode:    app.AppBranchRunPreviewModePlanOnly,
		Phases:  nil,
	})

	if strings.Contains(body, "| Check | Status |") {
		t.Errorf("phase checks table should be omitted when Phases is nil\n%s", body)
	}
}

func TestBuildPRCommentBodyBuildOnlyOmitsInstallPhase(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrun-example",
		Status:  PRCommentStatusSuccess,
		Mode:    app.AppBranchRunPreviewModeBuildOnly,
		Phases: &PRCommentPhases{
			Config:  PRCommentPhaseValid,
			Builds:  PRCommentPhaseValid,
			Install: PRCommentPhaseConfiguring,
		},
	})

	if strings.Contains(body, "| Install |") {
		t.Errorf("install row should be omitted in build-only mode\n%s", body)
	}
	if !strings.Contains(body, "| Config | \u2705 Valid |") {
		t.Errorf("config valid row missing\n%s", body)
	}
}

func TestBuildPRCommentBodyMCPDocsLink(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrun-example",
		Status:  PRCommentStatusSuccess,
	})

	if !strings.Contains(body, "https://docs.nuon.co/guides/agents/overview") {
		t.Errorf("comment body missing MCP docs link\n%s", body)
	}
	if !strings.Contains(body, "[MCP-enabled assistant](https://docs.nuon.co/guides/agents/overview)") {
		t.Errorf("comment body MCP link has wrong format\n%s", body)
	}
}

func TestBuildPRCommentBodySkippedOmitsMCPLink(t *testing.T) {
	body := BuildPRCommentBody(&PRCommentParams{
		AppName: "acme",
		RunID:   "abrun-example",
		Status:  PRCommentStatusSkipped,
	})

	if strings.Contains(body, "docs.nuon.co") {
		t.Errorf("skipped comment should not contain MCP docs link\n%s", body)
	}
}

func TestFinalizeFailedPhasesPendingToInvalid(t *testing.T) {
	phases := &PRCommentPhases{
		Config:  PRCommentPhaseValidating,
		Builds:  PRCommentPhaseBuilding,
		Install: PRCommentPhaseConfiguring,
	}
	FinalizeFailedPhases(phases)

	if phases.Config != PRCommentPhaseInvalid {
		t.Errorf("expected Config=Invalid, got %q", phases.Config)
	}
	if phases.Builds != PRCommentPhaseInvalid {
		t.Errorf("expected Builds=Invalid, got %q", phases.Builds)
	}
	if phases.Install != PRCommentPhaseInvalid {
		t.Errorf("expected Install=Invalid, got %q", phases.Install)
	}
}

func TestFinalizeFailedPhasesPreservesTerminalPhases(t *testing.T) {
	phases := &PRCommentPhases{
		Config:  PRCommentPhaseValid,
		Builds:  PRCommentPhaseInvalid,
		Install: "",
	}
	FinalizeFailedPhases(phases)

	if phases.Config != PRCommentPhaseValid {
		t.Errorf("Valid phase should not be changed, got %q", phases.Config)
	}
	if phases.Builds != PRCommentPhaseInvalid {
		t.Errorf("Invalid phase should not be changed, got %q", phases.Builds)
	}
	if phases.Install != "" {
		t.Errorf("empty phase should remain empty, got %q", phases.Install)
	}
}

func TestInstallImpactFromStepMetadata(t *testing.T) {
	metadata := map[string]any{
		"total_installs": float64(2),
		"install_groups": []any{
			map[string]any{
				"install_group_name": "canary",
				"installs": []any{
					map[string]any{
						"install_id":      "inst-1",
						"install_name":    "canary-1",
						"added":           float64(3),
						"changed":         float64(1),
						"removed":         float64(0),
						"unchanged":       float64(10),
						"sandbox_changed": false,
						"stack_changed":   true,
					},
				},
			},
			map[string]any{
				"install_group_name": "prod",
				"installs": []any{
					map[string]any{
						"install_id":   "inst-2",
						"install_name": "prod-1",
						"added":        float64(0),
						"changed":      float64(0),
						"removed":      float64(0),
						"unchanged":    float64(5),
					},
				},
			},
		},
	}

	groups := installImpactFromStepMetadata(metadata)

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].GroupName != "canary" {
		t.Errorf("expected canary group, got %q", groups[0].GroupName)
	}
	if len(groups[0].Installs) != 1 {
		t.Fatalf("expected 1 canary install, got %d", len(groups[0].Installs))
	}
	inst := groups[0].Installs[0]
	if inst.InstallID != "inst-1" || inst.InstallName != "canary-1" {
		t.Errorf("unexpected install: %+v", inst)
	}
	if inst.Added != 3 || inst.Changed != 1 || inst.Unchanged != 10 {
		t.Errorf("unexpected counts: added=%d changed=%d unchanged=%d", inst.Added, inst.Changed, inst.Unchanged)
	}
	if !inst.StackChanged {
		t.Errorf("expected stack_changed=true")
	}
	if groups[1].GroupName != "prod" {
		t.Errorf("expected prod group, got %q", groups[1].GroupName)
	}
}

func TestInstallImpactFromStepMetadataEmpty(t *testing.T) {
	if groups := installImpactFromStepMetadata(map[string]any{}); groups != nil {
		t.Errorf("expected nil for missing key, got %v", groups)
	}
	if groups := installImpactFromStepMetadata(map[string]any{"install_groups": []any{}}); groups != nil {
		t.Errorf("expected nil for empty slice, got %v", groups)
	}
}
