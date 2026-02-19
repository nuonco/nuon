package installs

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// Exit codes for workflow watch command
const (
	ExitCodeSuccess          = 0
	ExitCodeFailed           = 1
	ExitCodeCancelled        = 2
	ExitCodeApprovalRequired = 3
	ExitCodeStepFailed       = 4
	ExitCodeInterrupt        = 130
)

// WorkflowsWatch polls a workflow until it reaches a terminal state or requires approval
func (s *Service) WorkflowsWatch(ctx context.Context, installID, workflowID string, interval time.Duration, asJSON, quiet bool) (int, error) {
	view := ui.NewListView()

	// Set up signal handling for graceful interrupt
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Resolve workflow ID if not provided directly
	if workflowID == "" {
		if installID == "" {
			return ExitCodeFailed, view.Error(fmt.Errorf("either --workflow-id or --install-id is required"))
		}

		resolvedInstallID, err := lookup.InstallID(ctx, s.api, installID)
		if err != nil {
			return ExitCodeFailed, ui.PrintError(err)
		}

		// Get the latest/active workflow for this install
		workflows, _, err := s.listWorkflows(ctx, resolvedInstallID, 0, 1)
		if err != nil {
			return ExitCodeFailed, view.Error(err)
		}
		if len(workflows) == 0 {
			return ExitCodeFailed, view.Error(fmt.Errorf("no workflows found for install %s", installID))
		}
		workflowID = workflows[0].ID
	}

	isTTY := term.IsTerminal(int(os.Stdout.Fd()))
	firstRender := true

	for {
		select {
		case <-sigChan:
			if !quiet {
				fmt.Println("\nInterrupted")
			}
			return ExitCodeInterrupt, nil
		case <-ctx.Done():
			return ExitCodeInterrupt, nil
		default:
		}

		workflow, err := s.api.GetWorkflow(ctx, workflowID)
		if err != nil {
			return ExitCodeFailed, view.Error(err)
		}

		// Check for approval-pending steps
		pendingStep := findApprovalPendingStep(workflow.Steps)
		if pendingStep != nil {
			if asJSON {
				ui.PrintJSON(map[string]interface{}{
					"workflow":     workflow,
					"pending_step": pendingStep,
					"status":       "approval_required",
				})
			} else if !quiet {
				s.renderWorkflowStatus(workflow, isTTY, firstRender)
				fmt.Println()
				fmt.Printf("⏸️  Approval required for step: %s (%s)\n", pendingStep.Name, pendingStep.ID)
				fmt.Printf("   Run: nuon installs workflows steps approve -w %s -s %s\n", workflowID, pendingStep.ID)
			}
			return ExitCodeApprovalRequired, nil
		}

		// Check for failed steps
		failedStep := findFailedStep(workflow.Steps)
		if failedStep != nil {
			if asJSON {
				ui.PrintJSON(map[string]interface{}{
					"workflow":    workflow,
					"failed_step": failedStep,
					"status":      "step_failed",
				})
			} else if !quiet {
				s.renderWorkflowStatus(workflow, isTTY, firstRender)
				fmt.Println()
				fmt.Printf("❌ Step failed: %s (%s)\n", failedStep.Name, failedStep.ID)
			}
			return ExitCodeStepFailed, nil
		}

		// Check for terminal states
		if workflow.Status != nil {
			status := workflow.Status.Status
			switch status {
			case models.AppStatusSuccess:
				if asJSON {
					ui.PrintJSON(workflow)
				} else if !quiet {
					s.renderWorkflowStatus(workflow, isTTY, firstRender)
					fmt.Println()
					fmt.Println("✅ Workflow succeeded")
				}
				return ExitCodeSuccess, nil

			case models.AppStatusError:
				if asJSON {
					ui.PrintJSON(workflow)
				} else if !quiet {
					s.renderWorkflowStatus(workflow, isTTY, firstRender)
					fmt.Println()
					fmt.Println("❌ Workflow failed")
				}
				return ExitCodeFailed, nil

			case models.AppStatusCancelled:
				if asJSON {
					ui.PrintJSON(workflow)
				} else if !quiet {
					s.renderWorkflowStatus(workflow, isTTY, firstRender)
					fmt.Println()
					fmt.Println("🚫 Workflow cancelled")
				}
				return ExitCodeCancelled, nil
			}
		}

		// Render current status
		if !quiet {
			if asJSON {
				ui.PrintJSON(workflow)
			} else {
				s.renderWorkflowStatus(workflow, isTTY, firstRender)
			}
		}
		firstRender = false

		// Wait for next poll interval
		select {
		case <-sigChan:
			if !quiet {
				fmt.Println("\nInterrupted")
			}
			return ExitCodeInterrupt, nil
		case <-ctx.Done():
			return ExitCodeInterrupt, nil
		case <-time.After(interval):
			// Continue polling
		}
	}
}

// findApprovalPendingStep returns the first step that is waiting for approval
func findApprovalPendingStep(steps []*models.AppWorkflowStep) *models.AppWorkflowStep {
	for _, step := range steps {
		if step.Approval != nil && step.Status != nil {
			// Check if step is waiting for approval (approval-awaiting status)
			if step.Status.Status == models.AppStatusApprovalDashAwaiting && !step.Finished {
				return step
			}
		}
	}
	return nil
}

// findFailedStep returns the first step that has failed
func findFailedStep(steps []*models.AppWorkflowStep) *models.AppWorkflowStep {
	for _, step := range steps {
		if step.Status != nil && step.Status.Status == models.AppStatusError {
			return step
		}
	}
	return nil
}

// renderWorkflowStatus renders the workflow status table
func (s *Service) renderWorkflowStatus(workflow *models.AppWorkflow, isTTY, firstRender bool) {
	view := ui.NewListView()

	// Clear screen on TTY for clean refresh (except first render)
	if isTTY && !firstRender {
		fmt.Print("\033[H\033[2J") // ANSI escape to clear screen and move cursor to top
	}

	// Print workflow header
	status := ""
	if workflow.Status != nil {
		status = string(workflow.Status.Status)
	}
	fmt.Printf("Workflow: %s (%s)\n", workflow.Name, workflow.ID)
	fmt.Printf("Type:     %s\n", workflow.Type)
	fmt.Printf("Status:   %s\n", status)

	startedAt, _ := time.Parse(time.RFC3339Nano, workflow.StartedAt)
	if !startedAt.IsZero() {
		fmt.Printf("Started:  %s\n", startedAt.Format(time.Stamp))
	}

	if workflow.Finished {
		finishedAt, _ := time.Parse(time.RFC3339Nano, workflow.FinishedAt)
		fmt.Printf("Finished: %s\n", finishedAt.Format(time.Stamp))
		fmt.Printf("Duration: %s\n", time.Duration(workflow.ExecutionTime).String())
	} else {
		elapsed := time.Since(startedAt)
		fmt.Printf("Elapsed:  %s\n", elapsed.Round(time.Second).String())
	}
	fmt.Println()

	// Print steps table using existing formatter
	if len(workflow.Steps) > 0 {
		fmt.Println("Steps:")
		view.Render(formatWorkflowSteps(workflow.Steps))
	}
}
