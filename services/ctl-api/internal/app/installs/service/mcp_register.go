package service

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_installs",
		"List installs",
		"List all installs in the current org. Optionally filter by app_id to see installs for a specific app. Returns install name, ID, app, status, and cloud platform.",
	), s.mcpListInstalls)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_install",
		"Get install",
		"Get a single install by name or ID. Returns full install details including app info, cloud account, sandbox config, and component status. Accepts either the install name or ID.",
	), s.mcpGetInstall)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_install_components",
		"List install components",
		"List all components deployed on an install with their current deploy status and latest deploy info. Use this to see what's deployed and whether deploys are healthy.",
	), s.mcpListInstallComponents)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"list_workflows",
		"List workflows",
		"List the 20 most recent workflows for an install. Returns workflow type, status, and timestamps. Use get_workflow for full step details on a specific workflow.",
	), s.mcpListWorkflows)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_workflow",
		"Get workflow",
		"Get a workflow by ID with a summary of all steps and their statuses. If a step is awaiting approval, the pending_approval field will contain the approval_id needed to approve or reject it. Use this to follow a workflow's progress and find pending approvals.",
	), s.mcpGetWorkflow)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_pending_approvals",
		"Get pending approvals",
		"List all pending workflow step approvals across the entire org. Returns approvals that have not yet received a response. Use approve_step or reject_step to respond to them.",
	), s.mcpGetPendingApprovals)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"watch_workflow",
		"Watch workflow",
		"Watch a workflow for status changes. If last_known_status is provided and differs from current status, "+
			"returns immediately. Otherwise polls every 3 seconds until the status changes or the timeout is reached. "+
			"Use this to follow a deploy or provision workflow's progress without repeated polling.",
	), s.mcpWatchWorkflow)

	mcp.AddTool(server, apiPkg.MCPReadTool(
		"get_workflow_step",
		"Get workflow step",
		"Get full details for a workflow step including target type, approval status, policy validation, and execution time.",
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

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"update_install_inputs",
		"Update install inputs",
		"WRITE OPERATION: Update install input values (partial merge over current values). Starts an input-update workflow. "+
			"deploy_dependents defaults to true. Use get_install_inputs first to inspect current values.",
		false,
		false,
	), s.mcpUpdateInstallInputs)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"deploy_install_components",
		"Deploy install components",
		"WRITE OPERATION: Deploy all components on an install. Returns a workflow_id; use get_workflow and watch_workflow to follow progress. "+
			"Set plan_only to generate plans without applying.",
		true,
		false,
	), s.mcpDeployInstallComponents)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"reprovision_install",
		"Reprovision install",
		"WRITE OPERATION: Reprovision an install (stack, sandbox, then components). Returns a workflow_id. "+
			"Set plan_only to generate plans without applying.",
		true,
		false,
	), s.mcpReprovisionInstall)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"reprovision_sandbox",
		"Reprovision sandbox",
		"WRITE OPERATION: Reprovision only the install sandbox. Set skip_components to leave components unchanged after the sandbox apply. "+
			"Returns a workflow_id.",
		true,
		false,
	), s.mcpReprovisionSandbox)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"deprovision_install",
		"Deprovision install",
		"WRITE OPERATION: Deprovision an install (tear down components and cloud resources). This is destructive. "+
			"confirm must be true to apply. Ask the user before setting confirm. plan_only does not require confirm. Returns a workflow_id.",
		true,
		false,
	), s.mcpDeprovisionInstall)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"deprovision_sandbox",
		"Deprovision sandbox",
		"WRITE OPERATION: Deprovision only the install sandbox, leaving the stack. This is destructive. "+
			"confirm must be true to apply. Ask the user before setting confirm. plan_only does not require confirm. Returns a workflow_id.",
		true,
		false,
	), s.mcpDeprovisionSandbox)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"approve_step",
		"Approve step",
		"WRITE OPERATION: Approve a pending workflow step approval. This unblocks the workflow and allows it to proceed to the next step. "+
			"The approval is irreversible — once approved, the workflow will continue executing (e.g., terraform apply, helm install). "+
			"Always review the plan contents via get_workflow before approving. Requires the approval_id from get_workflow or get_pending_approvals.",
		true,
		false,
	), s.mcpApproveStep)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"reject_step",
		"Reject step",
		"WRITE OPERATION: Reject a pending workflow step approval. This stops the workflow from proceeding. "+
			"The rejection is irreversible for this workflow run — a new workflow must be triggered to retry. "+
			"Provide a reason to help the team understand why the approval was denied.",
		true,
		false,
	), s.mcpRejectStep)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"retry_step",
		"Retry step",
		"WRITE OPERATION: Retry a failed workflow step. The step must be retryable. "+
			"This creates a new attempt for the step and the workflow resumes from that point.",
		false,
		false,
	), s.mcpRetryStep)

	mcp.AddTool(server, apiPkg.MCPWriteTool(
		"cancel_workflow",
		"Cancel workflow",
		"WRITE OPERATION: Cancel an in-progress workflow. The workflow must be in a cancelable state "+
			"(in_progress, pending, awaiting_approval, or failed_pending_retry).",
		true,
		false,
	), s.mcpCancelWorkflow)
}
