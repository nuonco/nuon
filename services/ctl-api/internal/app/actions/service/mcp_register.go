package service

import "github.com/modelcontextprotocol/go-sdk/mcp"

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_install_actions",
		Description: "List actions available on an install. Returns action workflow IDs and names that can be inspected with get_action or run with run_action.",
	}, s.mcpListInstallActions)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_action",
		Description: "Get an install action by action_id (action_workflow_id). Returns whether it can be triggered manually and the install-pinned action_workflow_config_id.",
	}, s.mcpGetAction)

	mcp.AddTool(server, &mcp.Tool{
		Name: "run_action",
		Description: "WRITE OPERATION: Trigger a configured action on an install. Requires a manual trigger on the action config. " +
			"Pass install_id and action_id (action_workflow_id). Resolves the install-pinned config like the CLI create-run flow.",
	}, s.mcpRunAction)
}
