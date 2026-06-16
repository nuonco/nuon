package app

import (
	"strings"

	"github.com/nuonco/nuon/pkg/generics"
)

// computeWorkflowName renders the human-readable workflow title shown in the
// dashboard, CLI, and search index. It is the single source of truth — the
// install_workflows.name column is populated by Workflow.BeforeSave so the
// stored value stays in sync with this logic.
func computeWorkflowName(w *Workflow) string {
	suffix := workflowNameSuffix(w)

	if w.Type == WorkflowTypeActionWorkflowRun {
		if adhoc := metaValue(w, "adhoc_action"); adhoc != "" {
			actionName := metaValue(w, "install_action_workflow_name")
			return "Adhoc action run (" + actionName + ")" + suffix
		}
	}

	if w.Type == WorkflowTypeAppBranchesRun {
		return appBranchRunName(w) + suffix
	}

	var base string
	if !w.FinishedAt.IsZero() {
		base = w.Type.PastTenseName()
	} else {
		base = w.Type.Name()
	}
	if base == "" {
		base = strings.ReplaceAll(string(w.Type), "_", " ")
	}

	return base + suffix
}

func appBranchRunName(w *Workflow) string {
	eventType := metaValue(w, "event_type")
	commitSha := metaValue(w, "commit_sha")

	var base string
	switch eventType {
	case "push":
		base = "VCS push"
	case "pull_request":
		prNum := metaValue(w, "pr_number")
		if prNum != "" {
			base = "PR #" + prNum
		} else {
			base = "Pull request"
		}
	case "onboarding":
		base = "Onboarding run"
	default:
		base = "Manual run"
	}

	if commitSha != "" {
		if len(commitSha) > 7 {
			commitSha = commitSha[:7]
		}
		base += " (" + commitSha + ")"
	}

	return base
}

// workflowNameSuffix builds the trailing "(...)" segments appended to the
// base title — workflow-name-suffix metadata for all types, plus the
// action/runbook name where applicable.
func workflowNameSuffix(w *Workflow) string {
	var b strings.Builder

	if v := metaValue(w, WorkflowMetadataKeyWorkflowNameSuffix); v != "" {
		b.WriteString(" (")
		b.WriteString(v)
		b.WriteString(")")
	}

	switch w.Type {
	case WorkflowTypeActionWorkflowRun:
		// Skip if the adhoc branch already appended the action name.
		if metaValue(w, "adhoc_action") == "" {
			if v := metaValue(w, "install_action_workflow_name"); v != "" {
				b.WriteString(" (")
				b.WriteString(v)
				b.WriteString(")")
			}
		}
	case WorkflowTypeRunbookRun:
		if v := metaValue(w, "runbook_name"); v != "" {
			b.WriteString(" (")
			b.WriteString(v)
			b.WriteString(")")
		}
	}

	return b.String()
}

func metaValue(w *Workflow, key string) string {
	if w.Metadata == nil {
		return ""
	}
	return generics.FromPtrStr(w.Metadata[key])
}
