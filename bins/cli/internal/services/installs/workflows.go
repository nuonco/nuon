package installs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/plandiff"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) WorkflowsList(ctx context.Context, installID string, offset, limit int, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	workflows, hasMore, err := s.listWorkflows(ctx, installID, offset, limit)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(workflows)
		return nil
	}

	view.RenderPaging(formatWorkflows(workflows), offset, limit, hasMore)
	return nil
}

func (s *Service) WorkflowsGet(ctx context.Context, workflowID string, asJSON bool) error {
	view := ui.NewListView()

	workflow, err := s.api.GetWorkflow(ctx, workflowID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(workflow)
		return nil
	}

	fmt.Printf("Workflow: %s\n", workflow.ID)
	fmt.Printf("Name:     %s\n", workflow.Name)
	fmt.Printf("Type:     %s\n", workflow.Type)
	if workflow.Status != nil {
		fmt.Printf("Status:   %s\n", workflow.Status.Status)
	}
	startedAt, _ := time.Parse(time.RFC3339Nano, workflow.StartedAt)
	finishedAt, _ := time.Parse(time.RFC3339Nano, workflow.FinishedAt)
	fmt.Printf("Started:  %s\n", startedAt.Format(time.Stamp))
	if workflow.Finished {
		fmt.Printf("Finished: %s\n", finishedAt.Format(time.Stamp))
		fmt.Printf("Duration: %s\n", time.Duration(workflow.ExecutionTime).String())
	}
	fmt.Println()

	if len(workflow.Steps) > 0 {
		fmt.Println("Steps:")
		view.Render(formatWorkflowSteps(workflow.Steps))
	}

	return nil
}

func (s *Service) WorkflowStepsList(ctx context.Context, workflowID string, asJSON bool) error {
	view := ui.NewListView()

	steps, err := s.api.GetWorkflowSteps(ctx, workflowID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(steps)
		return nil
	}

	view.Render(formatWorkflowSteps(steps))
	return nil
}

// getLastProcessedStepID returns the ID of the most recently processed step in a workflow.
// This is useful for viewing logs of the step that just ran, rather than the step awaiting action.
// Logic: 1) Find first not-attempted step, return step at index-1
// 2) If no not-attempted step, return highest index step that is finished, in-progress, or error
func (s *Service) getLastProcessedStepID(ctx context.Context, workflowID string) (string, error) {
	steps, err := s.api.GetWorkflowSteps(ctx, workflowID)
	if err != nil {
		return "", err
	}
	if len(steps) == 0 {
		return "", fmt.Errorf("no steps found for workflow %s", workflowID)
	}

	// Find the first not-attempted step by index
	var firstNotAttempted *models.AppWorkflowStep
	for _, step := range steps {
		if step.Status != nil && step.Status.Status == models.AppStatusNotDashAttempted {
			if firstNotAttempted == nil || step.Idx < firstNotAttempted.Idx {
				firstNotAttempted = step
			}
		}
	}

	// If we found a not-attempted step, return the step just before it
	if firstNotAttempted != nil && firstNotAttempted.Idx > 0 {
		targetIdx := firstNotAttempted.Idx - 1
		for _, step := range steps {
			if step.Idx == targetIdx {
				return step.ID, nil
			}
		}
	}

	// Fallback: find the highest index step that has been processed
	// (finished, in-progress, error, success, or any status that indicates execution)
	var lastProcessed *models.AppWorkflowStep
	for _, step := range steps {
		if step.Status == nil {
			continue
		}
		status := step.Status.Status
		// Skip steps that haven't been processed
		if status == models.AppStatusNotDashAttempted || status == models.AppStatusPending {
			continue
		}
		if lastProcessed == nil || step.Idx > lastProcessed.Idx {
			lastProcessed = step
		}
	}

	if lastProcessed != nil {
		return lastProcessed.ID, nil
	}

	// Final fallback: return the step with highest index
	latest := steps[0]
	for _, step := range steps {
		if step.Idx > latest.Idx {
			latest = step
		}
	}
	return latest.ID, nil
}

// confirmStepAction displays step details and prompts for confirmation before taking an action.
func (s *Service) confirmStepAction(ctx context.Context, installID, workflowID, stepID, action string) (bool, error) {
	step, err := s.api.GetWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		return false, err
	}

	// Display step information
	fmt.Println()
	fmt.Printf("Step:   %s\n", step.Name)
	fmt.Printf("ID:     %s\n", step.ID)
	fmt.Printf("Index:  %d\n", step.Idx)
	if step.Status != nil {
		fmt.Printf("Status: %s\n", step.Status.Status)
	}

	// Try to display plan summary if available
	if step.StepTargetID != "" && step.StepTargetType == "install_deploys" && installID != "" {
		resolvedInstallID, err := lookup.InstallID(ctx, s.api, installID)
		if err == nil {
			deploy, err := s.api.GetInstallDeploy(ctx, resolvedInstallID, step.StepTargetID)
			if err == nil && len(deploy.RunnerJobs) > 0 {
				plan, err := s.api.GetRunnerJobPlan(ctx, deploy.RunnerJobs[0].ID)
				if err == nil && plan != "" {
					formatted, err := plandiff.FormatPlan(plan)
					if err == nil {
						fmt.Println()
						fmt.Println("Plan:")
						fmt.Println(formatted)
					}
				}
			}
		}
	}

	fmt.Println()

	prompt := fmt.Sprintf("Are you sure you want to %s step '%s'?", action, step.Name)
	return bubbles.Confirm(prompt, s.cfg.Interactive)
}

func (s *Service) WorkflowStepsGet(ctx context.Context, workflowID, stepID string, asJSON bool) error {
	view := ui.NewListView()

	// If stepID is not provided, use the last processed step
	if stepID == "" {
		var err error
		stepID, err = s.getLastProcessedStepID(ctx, workflowID)
		if err != nil {
			return view.Error(err)
		}
	}

	step, err := s.api.GetWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(step)
		return nil
	}

	fmt.Printf("Step:           %s\n", step.ID)
	fmt.Printf("Name:           %s\n", step.Name)
	fmt.Printf("Execution Type: %s\n", step.ExecutionType)
	if step.Status != nil {
		fmt.Printf("Status:         %s\n", step.Status.Status)
	}
	fmt.Printf("Index:          %d\n", step.Idx)
	fmt.Printf("Group Index:    %d\n", step.GroupIdx)
	fmt.Printf("Retryable:      %t\n", step.Retryable)
	fmt.Printf("Skippable:      %t\n", step.Skippable)
	fmt.Printf("Finished:       %t\n", step.Finished)

	startedAt, _ := time.Parse(time.RFC3339Nano, step.StartedAt)
	finishedAt, _ := time.Parse(time.RFC3339Nano, step.FinishedAt)
	if !startedAt.IsZero() {
		fmt.Printf("Started:        %s\n", startedAt.Format(time.Stamp))
	}
	if step.Finished && !finishedAt.IsZero() {
		fmt.Printf("Finished At:    %s\n", finishedAt.Format(time.Stamp))
		fmt.Printf("Duration:       %s\n", time.Duration(step.ExecutionTime).String())
	}

	if step.StepTargetID != "" {
		fmt.Printf("\nTarget:\n")
		fmt.Printf("  Type: %s\n", step.StepTargetType)
		fmt.Printf("  ID:   %s\n", step.StepTargetID)
	}

	if step.Approval != nil {
		fmt.Printf("\nApproval:\n")
		fmt.Printf("  ID:   %s\n", step.Approval.ID)
		fmt.Printf("  Type: %s\n", step.Approval.Type)
	}

	if len(step.Links) > 0 {
		fmt.Printf("\nLinks:\n")
		for key, value := range step.Links {
			fmt.Printf("  %s: %v\n", key, value)
		}
	}

	if len(step.Metadata) > 0 {
		fmt.Printf("\nMetadata:\n")
		for key, value := range step.Metadata {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	return nil
}

func formatWorkflows(workflows []*models.AppWorkflow) [][]string {
	data := [][]string{
		{
			"ID",
			"NAME",
			"TYPE",
			"STATUS",
			"STARTED AT",
			"FINISHED AT",
			"UPDATED AT",
		},
	}
	for _, workflow := range workflows {
		startedAt, _ := time.Parse(time.RFC3339Nano, workflow.StartedAt)
		finishedAt, _ := time.Parse(time.RFC3339Nano, workflow.FinishedAt)
		updatedAt, _ := time.Parse(time.RFC3339Nano, workflow.UpdatedAt)
		status := ""
		if workflow.Status != nil {
			statusText := string(workflow.Status.Status)
			if styles.IsActionableStatus(workflow.Status.Status) {
				status = styles.GetStatusStyle(workflow.Status.Status).Render("→ " + statusText)
			} else {
				status = styles.GetStatusStyle(workflow.Status.Status).Render(statusText)
			}
		}

		data = append(data, []string{
			workflow.ID,
			workflow.Name,
			string(workflow.Type),
			status,
			startedAt.Format(time.Stamp),
			finishedAt.Format(time.Stamp),
			updatedAt.Format(time.Stamp),
		})
	}

	return data
}

func formatWorkflowSteps(steps []*models.AppWorkflowStep) [][]string {
	data := [][]string{
		{
			"IDX",
			"ID",
			"NAME",
			"STATUS",
			"EXECUTION TYPE",
			"APPROVAL",
			"POLICY",
			"FINISHED",
		},
	}
	for _, step := range steps {
		// Skip hidden steps from CLI output
		if step.ExecutionType == models.AppWorkflowStepExecutionTypeHidden {
			continue
		}
		status := ""
		if step.Status != nil {
			statusText := string(step.Status.Status)
			if styles.IsActionableStatus(step.Status.Status) {
				status = styles.GetStatusStyle(step.Status.Status).Render("→ " + statusText)
			} else {
				status = styles.GetStatusStyle(step.Status.Status).Render(statusText)
			}
		}
		approval := "-"
		if step.Approval != nil {
			approval = string(step.Approval.Type)
		}
		finished := "no"
		if step.Finished {
			finished = "yes"
		}

		policy := "-"
		if step.Status != nil && step.Status.Metadata != nil {
			policy = getPolicyColumnValue(step.Status.Metadata)
		}

		data = append(data, []string{
			fmt.Sprintf("%d", step.Idx),
			step.ID,
			step.Name,
			status,
			string(step.ExecutionType),
			approval,
			policy,
			finished,
		})
	}

	return data
}

func (s *Service) listWorkflows(ctx context.Context, installID string, offset, limit int) ([]*models.AppWorkflow, bool, error) {
	workflows, hasMore, err := s.api.GetWorkflows(ctx, installID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, hasMore, err
	}
	return workflows, hasMore, nil
}

func (s *Service) WorkflowStepApprove(ctx context.Context, installID, workflowID, stepID, note string, skipConfirm, asJSON bool) error {
	view := ui.NewListView()

	// If stepID is not provided, use the last processed step and require confirmation
	stepWasProvided := stepID != ""
	if stepID == "" {
		var err error
		stepID, err = s.getLastProcessedStepID(ctx, workflowID)
		if err != nil {
			return view.Error(err)
		}
	}

	// Require confirmation when using auto-resolved step (unless --yes flag is set)
	if !stepWasProvided && !skipConfirm {
		confirmed, err := s.confirmStepAction(ctx, installID, workflowID, stepID, "approve")
		if err != nil {
			return view.Error(err)
		}
		if !confirmed {
			fmt.Println("Aborted.")
			return nil
		}
	}

	step, err := s.api.GetWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		return view.Error(err)
	}

	if step.Approval == nil {
		return view.Error(fmt.Errorf("step %s does not have an approval", stepID))
	}

	resp, err := s.api.CreateWorkflowStepApprovalResponse(ctx, workflowID, stepID, step.Approval.ID, &models.ServiceCreateWorkflowStepApprovalResponseRequest{
		ResponseType: models.AppWorkflowStepResponseTypeApprove,
		Note:         note,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(resp)
		return nil
	}

	fmt.Printf("Approved step %s\n", stepID)
	return nil
}

func (s *Service) WorkflowStepReject(ctx context.Context, installID, workflowID, stepID, note string, skipConfirm, asJSON bool) error {
	view := ui.NewListView()

	// If stepID is not provided, use the last processed step and require confirmation
	stepWasProvided := stepID != ""
	if stepID == "" {
		var err error
		stepID, err = s.getLastProcessedStepID(ctx, workflowID)
		if err != nil {
			return view.Error(err)
		}
	}

	// Require confirmation when using auto-resolved step (unless --yes flag is set)
	if !stepWasProvided && !skipConfirm {
		confirmed, err := s.confirmStepAction(ctx, installID, workflowID, stepID, "reject")
		if err != nil {
			return view.Error(err)
		}
		if !confirmed {
			fmt.Println("Aborted.")
			return nil
		}
	}

	step, err := s.api.GetWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		return view.Error(err)
	}

	if step.Approval == nil {
		return view.Error(fmt.Errorf("step %s does not have an approval", stepID))
	}

	resp, err := s.api.CreateWorkflowStepApprovalResponse(ctx, workflowID, stepID, step.Approval.ID, &models.ServiceCreateWorkflowStepApprovalResponseRequest{
		ResponseType: models.AppWorkflowStepResponseTypeDeny,
		Note:         note,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(resp)
		return nil
	}

	fmt.Printf("Rejected step %s\n", stepID)
	return nil
}

func (s *Service) WorkflowStepRetry(ctx context.Context, installID, workflowID, stepID string, skipConfirm, asJSON bool) error {
	view := ui.NewListView()

	// If stepID is not provided, use the last processed step and require confirmation
	stepWasProvided := stepID != ""
	if stepID == "" {
		var err error
		stepID, err = s.getLastProcessedStepID(ctx, workflowID)
		if err != nil {
			return view.Error(err)
		}
	}

	// Require confirmation when using auto-resolved step (unless --yes flag is set)
	if !stepWasProvided && !skipConfirm {
		confirmed, err := s.confirmStepAction(ctx, installID, workflowID, stepID, "retry")
		if err != nil {
			return view.Error(err)
		}
		if !confirmed {
			fmt.Println("Aborted.")
			return nil
		}
	}

	resp, err := s.api.RetryOwnerWorkflow(ctx, workflowID, &models.ServiceRetryWorkflowByIDRequest{
		Operation: "retry-step",
		StepID:    stepID,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(resp)
		return nil
	}

	fmt.Printf("Retrying step %s\n", stepID)
	return nil
}

func (s *Service) WorkflowStepPlan(ctx context.Context, installID, workflowID, stepID string, asJSON bool) error {
	view := ui.NewListView()

	// If stepID is not provided, use the last processed step
	if stepID == "" {
		var err error
		stepID, err = s.getLastProcessedStepID(ctx, workflowID)
		if err != nil {
			return view.Error(err)
		}
	}

	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	step, err := s.api.GetWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		return view.Error(err)
	}

	if step.StepTargetID == "" {
		return view.Error(fmt.Errorf("step %s does not have a target (no deploy associated)", stepID))
	}

	if step.StepTargetType != "install_deploys" {
		return view.Error(fmt.Errorf("step target type %s is not a deploy", step.StepTargetType))
	}

	deploy, err := s.api.GetInstallDeploy(ctx, installID, step.StepTargetID)
	if err != nil {
		return view.Error(err)
	}

	if len(deploy.RunnerJobs) == 0 {
		return view.Error(fmt.Errorf("no runner jobs found for deploy %s", step.StepTargetID))
	}

	runnerJob := deploy.RunnerJobs[0]
	plan, err := s.api.GetRunnerJobPlan(ctx, runnerJob.ID)
	if err != nil {
		return view.Error(err)
	}

	if plan == "" {
		fmt.Println("No plan available")
		policyNames, _ := s.getPolicyNameMap(ctx, installID, workflowID)
		displayPolicyViolationsIfPresent(step, policyNames)
		return nil
	}

	var policyNames map[string]string
	if !asJSON {
		policyNames, _ = s.getPolicyNameMap(ctx, installID, workflowID)
	}

	if asJSON {
		result := map[string]any{"plan": plan}
		if step.Status != nil && step.Status.Metadata != nil {
			denyViolations, warnViolations := extractPolicyViolations(step.Status.Metadata)
			if len(denyViolations) > 0 {
				result["deny_violations"] = denyViolations
			}
			if len(warnViolations) > 0 {
				result["warn_violations"] = warnViolations
			}
		}
		ui.PrintJSON(result)
		return nil
	}

	// Try to format the plan with human-readable diff output
	formatted, err := plandiff.FormatPlan(plan)
	if err != nil {
		// Fall back to raw plan output if formatting fails
		fmt.Println(plan)
		displayPolicyViolationsIfPresent(step, policyNames)
		return nil
	}

	fmt.Println(formatted)
	displayPolicyViolationsIfPresent(step, policyNames)
	return nil
}

// displayPolicyViolationsIfPresent checks step metadata for policy violations and displays them.
func displayPolicyViolationsIfPresent(step *models.AppWorkflowStep, policyNames map[string]string) {
	if step.Status == nil || step.Status.Metadata == nil {
		return
	}

	denyViolations, warnViolations := extractPolicyViolations(step.Status.Metadata)
	output := formatPolicyViolationsDisplay(denyViolations, warnViolations, policyNames)
	if output != "" {
		fmt.Print(output)
	}
}

func (s *Service) WorkflowSetApprovalOption(ctx context.Context, workflowID string, approveAll, prompt, asJSON bool) error {
	view := ui.NewListView()

	var approvalOption models.AppInstallApprovalOption
	if approveAll {
		approvalOption = models.AppInstallApprovalOptionApproveDashAll
	} else if prompt {
		approvalOption = models.AppInstallApprovalOptionPrompt
	} else {
		return view.Error(fmt.Errorf("must specify either --approve-all or --prompt"))
	}

	workflow, err := s.api.UpdateWorkflow(ctx, workflowID, &models.ServiceUpdateWorkflowRequest{
		ApprovalOption: &approvalOption,
	})
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(workflow)
		return nil
	}

	fmt.Printf("Updated workflow %s approval option to %s\n", workflowID, approvalOption)
	return nil
}

// policyViolation represents a policy violation from step metadata.
type policyViolation struct {
	PolicyID   string `json:"policy_id"`
	PolicyName string `json:"policy_name"`
	Message    string `json:"message"`
	Severity   string `json:"severity"`
}

// extractPolicyViolations extracts deny and warn violations from step metadata.
func extractPolicyViolations(metadata map[string]any) ([]policyViolation, []policyViolation) {
	var denyViolations, warnViolations []policyViolation

	if denyRaw, ok := metadata["deny_violations"]; ok {
		denyViolations = parsePolicyViolations(denyRaw)
	}
	if warnRaw, ok := metadata["warn_violations"]; ok {
		warnViolations = parsePolicyViolations(warnRaw)
	}

	return denyViolations, warnViolations
}

// parsePolicyViolations converts a raw interface{} to a slice of policy violations.
func parsePolicyViolations(raw any) []policyViolation {
	var violations []policyViolation

	data, err := json.Marshal(raw)
	if err != nil {
		return violations
	}

	_ = json.Unmarshal(data, &violations)
	return violations
}

// formatPolicyViolationsDisplay formats policy violations for CLI output.
func formatPolicyViolationsDisplay(denyViolations, warnViolations []policyViolation, policyNames map[string]string) string {
	if len(denyViolations) == 0 && len(warnViolations) == 0 {
		return ""
	}

	var output string
	separator := styles.TextSubtle.Render("─── Policy Violations ───────────────────────────────────────────")
	endSeparator := styles.TextSubtle.Render("──────────────────────────────────────────────────────────────────")

	output += "\n" + separator + "\n"

	if len(denyViolations) > 0 {
		output += fmt.Sprintf("\n%s\n", styles.TextError.Render(fmt.Sprintf("✗ DENY VIOLATIONS (%d)", len(denyViolations))))
		for _, v := range denyViolations {
			name := policyDisplayName(v, policyNames)
			output += fmt.Sprintf("\n  %s %s\n", styles.TextSubtle.Render("Policy:"), styles.TextPrimary.Render(name))
			output += fmt.Sprintf("  %s %s\n", styles.TextSubtle.Render("Message:"), v.Message)
		}
	}

	if len(warnViolations) > 0 {
		output += fmt.Sprintf("\n%s\n", styles.TextWarning.Render(fmt.Sprintf("⚠ WARNINGS (%d)", len(warnViolations))))
		for _, v := range warnViolations {
			name := policyDisplayName(v, policyNames)
			output += fmt.Sprintf("\n  %s %s\n", styles.TextSubtle.Render("Policy:"), styles.TextPrimary.Render(name))
			output += fmt.Sprintf("  %s %s\n", styles.TextSubtle.Render("Message:"), v.Message)
		}
	}

	output += "\n" + endSeparator + "\n"

	return output
}

func policyDisplayName(v policyViolation, policyNames map[string]string) string {
	if v.PolicyName != "" {
		if v.PolicyID != "" {
			return fmt.Sprintf("%s (%s)", v.PolicyName, v.PolicyID)
		}
		return v.PolicyName
	}
	if v.PolicyID == "" {
		return ""
	}
	if policyNames == nil {
		return v.PolicyID
	}
	if name, ok := policyNames[v.PolicyID]; ok && name != "" {
		return fmt.Sprintf("%s (%s)", name, v.PolicyID)
	}
	return v.PolicyID
}

func (s *Service) getPolicyNameMap(ctx context.Context, installID, workflowID string) (map[string]string, error) {
	resolvedInstallID := installID
	if resolvedInstallID == "" && workflowID != "" {
		workflow, err := s.api.GetWorkflow(ctx, workflowID)
		if err != nil {
			return nil, err
		}
		if workflow.OwnerType == "installs" {
			resolvedInstallID = workflow.OwnerID
		}
	}
	if resolvedInstallID == "" {
		return nil, nil
	}

	install, err := s.api.GetInstall(ctx, resolvedInstallID)
	if err != nil {
		return nil, err
	}
	if install.AppID == "" {
		return nil, nil
	}

	policiesConfig, err := s.api.GetLatestAppPoliciesConfig(ctx, install.AppID)
	if err != nil {
		return nil, err
	}
	policyNames := make(map[string]string)
	for _, policy := range policiesConfig.Policies {
		if policy == nil || policy.ID == "" {
			continue
		}
		name := policy.Name
		if name == "" {
			name = policy.ID
		}
		policyNames[policy.ID] = name
	}

	return policyNames, nil
}

// getPolicyColumnValue returns a formatted policy status string for table display.
func getPolicyColumnValue(metadata map[string]any) string {
	if metadata == nil {
		return "-"
	}

	denyViolations, warnViolations := extractPolicyViolations(metadata)
	denyCount := len(denyViolations)
	warnCount := len(warnViolations)

	if denyCount == 0 && warnCount == 0 {
		if _, hasDeny := metadata["deny_violations"]; hasDeny {
			return styles.TextSuccess.Render("✓")
		}
		if _, hasWarn := metadata["warn_violations"]; hasWarn {
			return styles.TextSuccess.Render("✓")
		}
		return "-"
	}

	var parts []string
	if denyCount > 0 {
		parts = append(parts, styles.TextError.Render(fmt.Sprintf("✗ %d", denyCount)))
	}
	if warnCount > 0 {
		parts = append(parts, styles.TextWarning.Render(fmt.Sprintf("⚠ %d", warnCount)))
	}

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}
	return result
}
