package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_installs",
		"List installs",
		"List installs in the current org. Optionally filter by app_id to see installs for a specific app. Returns install name, ID, app, status, and cloud platform."+apiPkg.MCPListToolHint,
	), s.mcpListInstalls)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_install",
		"Get install",
		"Get a compact overview of an install by name or ID, including app and branch identity, deployment and health rollups, input names, and the workflow or app-branch run that most recently updated it.",
	), s.mcpGetInstall)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_install_components",
		"List install components",
		"List all components deployed on an install with current deploy and health status plus the latest deploy. Use this to see what's deployed and whether components are healthy.",
	), s.mcpListInstallComponents)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_install_readme",
		"Get install README",
		"Get the app README rendered with the current state of an install. Returns Markdown and any interpolation warnings.",
	), s.mcpGetInstallReadme)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_install_health",
		"Get install health",
		"Get the current health rollup for an install and each component, including health descriptions, last report time, and cluster access errors.",
	), s.mcpGetInstallHealth)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_workflows",
		"List workflows",
		"List the 20 most recent workflows for an install. Returns workflow type, status, and timestamps. Use get_workflow for full step details on a specific workflow.",
	), s.mcpListWorkflows)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_workflow",
		"Get workflow",
		"Get a workflow by ID with a summary of all steps and their statuses. approval_option is prompt or approve-all. pending_approvals lists every step whose approval has no response, including stored change counts and next_actions. It does not include plan diffs. When you stop for user input (approval, retry, or any next_actions menu), do not report only completed_steps/total_steps. List every step from the steps array by name and status so the user can see what already finished before the waiting step. Keep that list compact (name and status only); do not dump stack_setup, tfvars, or outputs again unless the user is waiting on await install stack. Present each pending approval's next_actions as a menu. Do not call get_approval_plan_diff or render diffs until the user chooses review_plan_diff. Choosing review means load every diff: call get_approval_plan_diff and follow continue_review, using each response's offset, until continue_review is absent. A full page is not the end of the plan. Do not reply until every page is loaded, then show one review. Plan review is optional: if the user selects approve, reject, or approve_all from next_actions, that selection is confirmation. Call the selected tool immediately without loading diffs or asking again. Offer approve_all only when that action is present, and only once per workflow. Tell the user approve_all applies every remaining plan in this workflow before presenting that choice. A parallel group can have several entries. After one approval is resolved, call watch_workflow to find the next. When a step includes stack_setup, show that install-stack setup to the user: status, quick launch link, template URL, create and update commands, and the Terraform clone, backend, tfvars, and apply command when those fields are set. Then keep watching. That step stays in progress until the stack is applied and the runner connects. When a step is waiting for a retry, next_actions lists every menu choice: show_step_logs, retry when the step is retryable, skip when it is skippable, and cancel_workflow. Present every one of those actions. Do not show only Show step logs. Call get_workflow_step_logs when presenting the menu; showing logs does not retry, skip, or cancel, and the workflow stays parked. If the log page has has_more, offer to load older logs with next_cursor. If message says the step has no log stream, tell the user that. If the user selects retry, skip, or cancel_workflow, that selection is confirmation: call that action's tool immediately. When a step includes composite_error, show its message, type, and each section before the menu. hints.skip_auto_retry means the step parked on the first failure.",
	), s.mcpGetWorkflow)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_pending_approvals",
		"Get pending approvals",
		"List pending workflow step approvals across the org (id, type, step, workflow, stored change counts, and next_actions). Does not return plan diffs. Present each approval's next_actions as a menu. Do not call get_approval_plan_diff until the user chooses review_plan_diff. Choosing review means load every diff: call get_approval_plan_diff and follow continue_review, using each response's offset, until continue_review is absent. A full page is not the end of the plan. Do not reply until every page is loaded, then show one review. Plan review is optional: selecting approve, reject, or approve_all is confirmation; call that tool immediately without loading diffs or asking again. Offer approve_all only when that action is present."+apiPkg.MCPListToolHint,
	), s.mcpGetPendingApprovals)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"watch_workflow",
		"Watch workflow",
		"Watch an install workflow, including a runbook_run from run_runbook or an action_workflow_run from run_action. "+
			"Pass cursor from the previous response to wake on any step, approval, or retry change, even when the workflow status stays in_progress. "+
			"last_known_status is the older check and compares only the workflow status. "+
			"If neither is set, this returns the current snapshot and a cursor immediately. "+
			"Otherwise it polls every 3 seconds and returns when a step starts or finishes, or the timeout is reached. It does not return only to repeat an in-progress step. "+
			"step_progress lists only steps to show now: no pending or queued steps, and no step already reported at the same status. An in-progress step is included again. Write step_progress as given and do not add other steps from workflow.steps. If step_progress is empty, write nothing about steps. Pass reported from this response on the next watch. After every watch_workflow result, the next user-visible message is step_progress. Then call next_action. A row of watch_workflow calls with no step_progress between them is wrong. Do not end the turn after printing the list, and do not wait for the user. Only stop when the workflow is terminal or a step needs a decision. If steps is empty, say the workflow is still generating steps, then keep watching. "+
			"When you stop for user input (approval, retry, or any next_actions menu), do not report only completed_steps/total_steps. "+
			"List every step from the steps array by name and status so the user can see what already finished before the waiting step. "+
			"Keep that list compact (name and status only); do not dump stack_setup, tfvars, or outputs again unless the user is waiting on await install stack. "+
			"If a step includes stack_setup, show that install-stack setup once (status, quick launch link, template URL, create and update commands, and Terraform files when present), then keep watching. "+
			"An in-progress await-install-stack step is the customer applying the stack. A timeout there is not a stop. "+
			"When a step is waiting for a retry, next_actions lists every menu choice: show_step_logs, retry when the step is retryable, skip when it is skippable, and cancel_workflow. Present every one of those actions. Do not show only Show step logs. Call get_workflow_step_logs when presenting the menu; showing logs does not retry, skip, or cancel. If the user selects retry, skip, or cancel_workflow, that selection is confirmation: call that action's tool immediately. When a step includes composite_error, show its message, type, and each section before the menu.",
	), s.mcpWatchWorkflow)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_approval_plan_diff",
		"Get approval plan diff",
		"Get one page of unified diffs for a workflow step approval. Pages are for the agent, not the user. "+
			"When the user asks to review a plan, keep calling this until every change is loaded. "+
			"If next_actions contains continue_review, call that tool immediately with its arguments, including offset. "+
			"Do not stop after one page, and do not stop because a page returned 100 changes. "+
			"Do not reply, do not render a partial page, and do not show a menu while continue_review is present. "+
			"After the last page, review every change from every page together in one response. Put each changes[].diff in a markdown diff fence. "+
			"Lead with delete and replace counts. no-op and read changes are omitted unless action asks for them. "+
			"Then present the last page's next_actions as a menu. Those are approve, reject, stop, and approve_all when allowed. "+
			"If too_large is true, the plan exceeded the parse cap: report the stored counts and do not invent a diff. "+
			"Review every entry in get_workflow.pending_approvals separately. Selecting approve, reject, or approve_all from the menu is sufficient confirmation; do not ask again.",
	), s.mcpGetApprovalPlanDiff)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_workflow_step",
		"Get workflow step",
		"Get full details for a workflow step including target type, approval status, policy validation, and execution time. The await install stack step includes stack_setup: status, quick launch link, template URL, create and update commands, and Terraform files when present. Show that setup, then keep watching the workflow until the stack is applied. When a step is waiting for a retry, next_actions lists every menu choice: show_step_logs, retry when the step is retryable, skip when it is skippable, and cancel_workflow. Present every one of those actions. Do not show only Show step logs. show_step_logs is read-only. If the user selects retry, skip, or cancel_workflow, that selection is confirmation: call that action's tool immediately. When composite_error is set, show its message, type, and each section before the menu.",
	), s.mcpGetWorkflowStep)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_deploys",
		"List deploys",
		"List the 20 most recent deploys for an install. Returns deploy ID, component, build, status, and timestamp.",
	), s.mcpListDeploys)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_deploy",
		"Get deploy",
		"Get deploy details by ID including component info, build, status, and log stream ID for log fetching.",
	), s.mcpGetDeploy)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_install_inputs",
		"Get install inputs",
		"Get the current input values for an install by name or ID. Returns the latest input revision as a name-to-value map.",
	), s.mcpGetInstallInputs)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_available_roles",
		"List available roles",
		"List IAM roles an install can assume for an operation (same as GET /v1/installs/{id}/available-roles). "+
			"Pass operation_type matching the write you will call (reprovision, deprovision, deploy, trigger, provision, teardown). "+
			"Pass principal_type sandbox, component, or action — and principal_id for component or action — to mark the default role. "+
			"Use a returned name as role on write tools; omit role to use the default.",
	), s.mcpListAvailableRoles)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"update_install_inputs",
		"Update install inputs",
		"WRITE OPERATION: Update install input values (partial merge over current values). Starts an input-update workflow. "+
			"deploy_dependents defaults to true. Use get_install_inputs first to inspect current values. "+
			"Call list_available_roles (operation_type=deploy) before passing role. "+
			mcpWatchAfterStart,
		false,
		false,
	), s.mcpUpdateInstallInputs)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"deploy_install_components",
		"Deploy install components",
		"WRITE OPERATION: Deploy all components on an install. "+
			"Set plan_only to generate plans without applying. Call list_available_roles (operation_type=deploy, principal_type=component) before passing role. "+
			mcpWatchAfterStart,
		true,
		false,
	), s.mcpDeployInstallComponents)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"reprovision_install",
		"Reprovision install",
		"WRITE OPERATION: Reprovision an install (stack, sandbox, then components). "+
			"Set plan_only to generate plans without applying. Set stack_only to reprovision only the stack (runner infra), leaving sandbox and components unchanged. "+
			"Call list_available_roles (operation_type=reprovision, principal_type=sandbox) before passing role. "+
			mcpWatchAfterStart,
		true,
		false,
	), s.mcpReprovisionInstall)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"reprovision_sandbox",
		"Reprovision sandbox",
		"WRITE OPERATION: Reprovision only the install sandbox. Set skip_components to leave components unchanged after the sandbox apply. "+
			"Call list_available_roles (operation_type=reprovision, principal_type=sandbox) before passing role. "+
			mcpWatchAfterStart,
		true,
		false,
	), s.mcpReprovisionSandbox)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"deprovision_install",
		"Deprovision install",
		"WRITE OPERATION: Deprovision an install (tear down components and cloud resources). This is destructive. "+
			"confirm must be true to apply. Ask the user before setting confirm. plan_only does not require confirm. "+
			mcpWatchAfterStart,
		true,
		false,
	), s.mcpDeprovisionInstall)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"deprovision_sandbox",
		"Deprovision sandbox",
		"WRITE OPERATION: Deprovision only the install sandbox, leaving the stack. This is destructive. "+
			"confirm must be true to apply. Ask the user before setting confirm. plan_only does not require confirm. "+
			"Call list_available_roles (operation_type=deprovision, principal_type=sandbox) before passing role. "+
			mcpWatchAfterStart,
		true,
		false,
	), s.mcpDeprovisionSandbox)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"approve_step",
		"Approve step",
		"WRITE OPERATION: Approve one pending workflow step approval. This unblocks that step and lets the workflow continue (for example terraform apply or helm install). "+
			"The approval is irreversible. Plan review is optional. If the user selected Approve this step from next_actions, that selection is confirmation: call this tool immediately without loading the plan diff or asking again. "+
			"Requires the approval_id from get_workflow or get_pending_approvals. Optional note is stored on the approval response. "+
			"If this tool returns an error, the approval was not applied and the step is still pending. "+
			"On success, call the returned next_action (watch_workflow) immediately.",
		true,
		false,
	), s.mcpApproveStep)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"approve_all",
		"Approve all",
		"WRITE OPERATION: Switch one workflow to approve-all, matching the dashboard Approve all button. "+
			"Approves steps already waiting and auto-approves every later plan in this workflow. Does not change the install default. "+
			"Only valid when approval_option is prompt, the workflow is still running, it is not a drift scan, it is not cancelled, and it has an approval step. "+
			"Tell the user this applies every remaining plan in this workflow before presenting the choice. Selecting Approve all is confirmation; call this tool without asking again. "+
			"Requires the workflow_id from get_workflow.",
		true,
		false,
	), s.mcpApproveAll)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"reject_step",
		"Reject step",
		"WRITE OPERATION: Reject a pending workflow step approval. This stops the workflow from proceeding. "+
			"The rejection is irreversible for this workflow run — a new workflow must be triggered to retry. "+
			"Plan review is optional. If the user selected Reject this step from next_actions, that selection is confirmation: call this tool immediately without loading the plan diff or asking again. "+
			"Provide a reason to help the team understand why the approval was denied. "+
			"If this tool returns an error, the rejection was not applied and the step is still pending. "+
			"On success, call the returned next_action (watch_workflow) immediately.",
		true,
		false,
	), s.mcpRejectStep)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"retry_step",
		"Retry step",
		"WRITE OPERATION: Retry a failed workflow step. The step must be retryable. "+
			"If the user selected Retry step from next_actions, that selection is confirmation: call this tool immediately. "+
			"This creates a new attempt for the step and the workflow resumes from that point. "+
			"On success, call the returned next_action (watch_workflow) immediately.",
		false,
		false,
	), s.mcpRetryStep)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"skip_step",
		"Skip step",
		"WRITE OPERATION: Skip a failed workflow step and continue the workflow. The step must be skippable. "+
			"If the user selected Skip step from next_actions, that selection is confirmation: call this tool immediately. "+
			"Changes from the skipped step are not applied. "+
			"On success, call the returned next_action (watch_workflow) immediately.",
		false,
		false,
	), s.mcpSkipStep)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"cancel_workflow",
		"Cancel workflow",
		"WRITE OPERATION: Cancel an in-progress workflow. The workflow must be in a cancelable state "+
			"(in_progress, pending, awaiting_approval, or failed_pending_retry). "+
			"If the user selected Cancel workflow from next_actions, that selection is confirmation: call this tool immediately. "+
			"On success, call the returned next_action (watch_workflow) until the workflow reaches a terminal status.",
		true,
		false,
	), s.mcpCancelWorkflow)
}
