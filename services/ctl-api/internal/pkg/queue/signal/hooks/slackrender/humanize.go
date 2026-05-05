package slackrender

import (
	"strings"
	"unicode"
)

// headerTitle returns a short, sentence-case human phrase describing what
// the workflow / step is doing, used as the bolded header of the parent
// post and inline in flat / child renders.
//
// Step-level events prefer the step's own display name (sentence-cased)
// when it carries information beyond the component name — this lets
// readers see "Provision runner service account" rather than the generic
// target-type fallback "Updating runner". Workflow-level events are
// routed by Workflow.Type. Approval events render as the underlying
// step's title — the approval is rendered through the transition /
// approval block, not the title.
func headerTitle(e Event) string {
	if (e.Kind == KindWorkflowStep || e.Kind == KindWorkflowStepApproval) && e.Step != nil {
		name := strings.TrimSpace(e.Step.Name)
		if name != "" && !strings.EqualFold(name, e.Step.ComponentName) {
			return sentenceCase(name)
		}
		if title := stepTitleFromTargetType(e.Step.TargetType); title != "" {
			return title
		}
	}

	if title := titleFromWorkflowType(e.Workflow.Type); title != "" {
		return title
	}

	if e.Workflow.Type != "" {
		return e.Workflow.Type
	}

	switch e.Kind {
	case KindWorkflow:
		return "Workflow"
	case KindWorkflowStep:
		return "Workflow step"
	case KindWorkflowStepApproval:
		return "Workflow step approval"
	}
	return "Event"
}

// sentenceCase upper-cases the first rune of s and leaves the rest
// unchanged. Step names from the DB are typically lower-case ("provision
// runner service account"), so a one-rune flip is enough to make them
// read as a sentence header.
func sentenceCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// titleFromWorkflowType maps the public workflow.type vocabulary to a
// short sentence-case phrase.
func titleFromWorkflowType(wfType string) string {
	switch wfType {
	case WorkflowTypeProvision:
		return "Provisioning install"
	case WorkflowTypeReprovision:
		return "Reprovisioning install"
	case WorkflowTypeManualDeploy:
		return "Deploying components"
	case WorkflowTypeDeployComponents:
		return "Deploying components"
	case WorkflowTypeTeardownComponent:
		return "Tearing down component"
	case WorkflowTypeTeardownComponents:
		return "Tearing down components"
	case WorkflowTypeActionWorkflowRun:
		return "Running action workflow"
	case WorkflowTypeDriftRun:
		return "Running drift check"
	case WorkflowTypeInputUpdate:
		return "Updating inputs"
	case WorkflowTypeSyncSecrets:
		return "Syncing secrets"
	case WorkflowTypeDeprovision:
		return "Deprovisioning install"
	case WorkflowTypeDeprovisionSandbox:
		return "Deprovisioning sandbox"
	case WorkflowTypeReprovisionSandbox:
		return "Reprovisioning sandbox"
	case WorkflowTypeAppConfigBuild:
		return "Building app config"
	case WorkflowTypeAppBranchesRun:
		return "Running app branch"
	}
	return ""
}

// stepTitleFromTargetType maps the step.target_type vocabulary to a
// sentence-case phrase.
func stepTitleFromTargetType(targetType string) string {
	switch targetType {
	case TargetTypeInstallDeploys:
		return "Deploying component"
	case TargetTypeInstallSandboxRuns:
		return "Running sandbox"
	case TargetTypeInstallActionWorkflowRuns:
		return "Running action workflow"
	case TargetTypeInstallCloudFormationStack,
		TargetTypeInstallRunnerUpdate:
		return "Updating runner"
	}
	return ""
}

// transitionPhrase returns a sentence-case verb phrase for a given
// transition.
func transitionPhrase(transition string) string {
	switch strings.TrimSpace(transition) {
	case TransitionStarted:
		return "Started"
	case TransitionSucceeded:
		return "Succeeded"
	case TransitionFailed:
		return "Failed"
	case TransitionCancelled:
		return "Cancelled"
	case TransitionRequested:
		return "Awaiting approval"
	case TransitionApproved:
		return "Approved"
	case TransitionRejected:
		return "Rejected"
	}
	if transition == "" {
		return ""
	}
	return transition
}
