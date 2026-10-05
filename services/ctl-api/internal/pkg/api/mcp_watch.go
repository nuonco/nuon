package api

// MCPWatchInstructions is included in MCP server instructions so clients
// show step progress between watch_workflow calls for the whole session.
const MCPWatchInstructions = "While watching a workflow, after every watch_workflow result the next user-visible message is step_progress. Write it as given: it omits pending steps and steps already shown at the same status, and repeats a step only while it is in progress. Do not add steps from workflow.steps. If step_progress is empty, do not list steps. Then call the returned next_action and pass its reported map. Stop only when the workflow is terminal or a step needs a decision."
