package installs

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
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

func (s *Service) WorkflowStepsGet(ctx context.Context, workflowID, stepID string, asJSON bool) error {
	view := ui.NewListView()

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

	if step.PolicyValidation != nil {
		fmt.Printf("\nPolicy Validation:\n")
		fmt.Printf("  ID:     %s\n", step.PolicyValidation.ID)
		fmt.Printf("  Status: %s\n", step.PolicyValidation.Status.Status)
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
			status = string(workflow.Status.Status)
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
			"FINISHED",
		},
	}
	for _, step := range steps {
		status := ""
		if step.Status != nil {
			status = string(step.Status.Status)
		}
		approval := "-"
		if step.Approval != nil {
			approval = string(step.Approval.Type)
		}
		finished := "no"
		if step.Finished {
			finished = "yes"
		}

		data = append(data, []string{
			fmt.Sprintf("%d", step.Idx),
			step.ID,
			step.Name,
			status,
			string(step.ExecutionType),
			approval,
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

func (s *Service) WorkflowStepApprove(ctx context.Context, workflowID, stepID, note string, asJSON bool) error {
	view := ui.NewListView()

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

func (s *Service) WorkflowStepReject(ctx context.Context, workflowID, stepID, note string, asJSON bool) error {
	view := ui.NewListView()

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

func (s *Service) WorkflowStepRetry(ctx context.Context, workflowID, stepID string, asJSON bool) error {
	view := ui.NewListView()

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
		return nil
	}

	if asJSON {
		ui.PrintJSON(map[string]string{"plan": plan})
		return nil
	}

	fmt.Println(plan)
	return nil
}

func (s *Service) WorkflowStepLogs(ctx context.Context, installID, workflowID, stepID string, asJSON bool) error {
	view := ui.NewListView()

	step, err := s.api.GetWorkflowStep(ctx, workflowID, stepID)
	if err != nil {
		return view.Error(err)
	}

	if step.StepTargetID == "" {
		return view.Error(fmt.Errorf("step %s does not have a target", stepID))
	}

	var logStreamID string
	switch step.StepTargetType {
	case "install_deploys":
		installID, err = lookup.InstallID(ctx, s.api, installID)
		if err != nil {
			return ui.PrintError(err)
		}
		deploy, err := s.api.GetInstallDeploy(ctx, installID, step.StepTargetID)
		if err != nil {
			return view.Error(err)
		}
		if deploy.LogStream != nil {
			logStreamID = deploy.LogStream.ID
		}
	case "install_sandbox_runs":
		return view.Error(fmt.Errorf("sandbox run logs not yet supported via this command"))
	default:
		return view.Error(fmt.Errorf("unsupported step target type: %s", step.StepTargetType))
	}

	if logStreamID == "" {
		return view.Error(fmt.Errorf("no log stream found for step %s", stepID))
	}

	logs, err := s.api.LogStreamReadLogs(ctx, logStreamID, "")
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(logs)
		return nil
	}

	for _, log := range logs {
		fmt.Println(log.Body)
	}
	return nil
}
