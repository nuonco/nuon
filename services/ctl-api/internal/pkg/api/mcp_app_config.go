package api

// MCPAppConfigInstructions is included in MCP server instructions so clients
// validate app config with the local CLI instead of uploading files.
const MCPAppConfigInstructions = "When validating an app config directory for this MCP server, use the Nuon CLI binary and -C value from the server instructions, running `apps validate` from that directory or with the directory as the argument. That match applies only to local CLI work done for this MCP server. It checks the files on disk and does not store or sync them. There is no MCP tool that accepts config file contents."
