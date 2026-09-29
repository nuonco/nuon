package slackrender

import (
	"strings"
	"unicode"
)

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
	case KindLabelAdded:
		return "Install label added"
	case KindAppBranchChanged:
		return "Install app branch changed"
	}
	return "Event"
}

func sentenceCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

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
	case WorkflowTypeReprovisionStack:
		return "Reprovisioning stack"
	case WorkflowTypeAppConfigBuild:
		return "Building app config"
	case WorkflowTypeAppBranchesRun,
		WorkflowTypeAppBranchesConfigRepoUpdate,
		WorkflowTypeAppBranchesComponentRepoUpdate:
		return "Running app branch"
	case WorkflowTypeRunbookRun:
		return "Running runbook"
	case WorkflowTypeComponentEnabled:
		return "Enabling component"
	case WorkflowTypeComponentDisabled:
		return "Disabling component"
	}
	return ""
}

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
	case TransitionAwaitingRetry:
		return "Awaiting manual retry"
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
