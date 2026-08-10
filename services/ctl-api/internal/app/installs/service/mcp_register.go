package service

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_installs",
		Description: "List all installs in the current org. Optionally filter by app_id to see installs for a specific app. Returns install name, ID, app, status, and cloud platform.",
	}, s.mcpListInstalls)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_install",
		Description: "Get a single install by name or ID. Returns full install details including app info, cloud account, sandbox config, and component status. Accepts either the install name or ID.",
	}, s.mcpGetInstall)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_install_components",
		Description: "List all components deployed on an install with their current deploy status and latest deploy info. Use this to see what's deployed and whether deploys are healthy.",
	}, s.mcpListInstallComponents)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_workflows",
		Description: "List the 20 most recent workflows for an install. Returns workflow type, status, and timestamps. Use get_workflow for full step details on a specific workflow.",
	}, s.mcpListWorkflows)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_workflow",
		Description: "Get a workflow by ID with a summary of all steps and their statuses. If a step is awaiting approval, the pending_approval field will contain the approval_id needed to approve or reject it. Use this to follow a workflow's progress and find pending approvals.",
	}, s.mcpGetWorkflow)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_pending_approvals",
		Description: "List all pending workflow step approvals across the entire org. Returns approvals that have not yet received a response. Use approve_step or reject_step to respond to them.",
	}, s.mcpGetPendingApprovals)

	mcp.AddTool(server, &mcp.Tool{
		Name: "watch_workflow",
		Description: "Watch a workflow for status changes. If last_known_status is provided and differs from current status, " +
			"returns immediately. Otherwise polls every 3 seconds until the status changes or the timeout is reached. " +
			"Use this to follow a deploy or provision workflow's progress without repeated polling.",
	}, s.mcpWatchWorkflow)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_workflow_step",
		Description: "Get full details for a workflow step including target type, approval status, policy validation, and execution time.",
	}, s.mcpGetWorkflowStep)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_deploys",
		Description: "List the 20 most recent deploys for an install. Returns deploy ID, component, build, status, and timestamp.",
	}, s.mcpListDeploys)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_deploy",
		Description: "Get deploy details by ID including component info, build, status, and log stream ID for log fetching.",
	}, s.mcpGetDeploy)

	mcp.AddTool(server, &mcp.Tool{
		Name: "approve_step",
		Description: "WRITE OPERATION: Approve a pending workflow step approval. This unblocks the workflow and allows it to proceed to the next step. " +
			"The approval is irreversible — once approved, the workflow will continue executing (e.g., terraform apply, helm install). " +
			"Always review the plan contents via get_workflow before approving. Requires the approval_id from get_workflow or get_pending_approvals.",
	}, s.mcpApproveStep)

	mcp.AddTool(server, &mcp.Tool{
		Name: "reject_step",
		Description: "WRITE OPERATION: Reject a pending workflow step approval. This stops the workflow from proceeding. " +
			"The rejection is irreversible for this workflow run — a new workflow must be triggered to retry. " +
			"Provide a reason to help the team understand why the approval was denied.",
	}, s.mcpRejectStep)

	mcp.AddTool(server, &mcp.Tool{
		Name: "retry_step",
		Description: "WRITE OPERATION: Retry a failed workflow step. The step must be retryable. " +
			"This creates a new attempt for the step and the workflow resumes from that point.",
	}, s.mcpRetryStep)

	mcp.AddTool(server, &mcp.Tool{
		Name: "cancel_workflow",
		Description: "WRITE OPERATION: Cancel an in-progress workflow. The workflow must be in a cancelable state " +
			"(in_progress, pending, awaiting_approval, or failed_pending_retry).",
	}, s.mcpCancelWorkflow)
}
