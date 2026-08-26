---
name: mcp-api-tool
description: Use this skill when adding or changing a ctl-api MCP tool (RegisterMCPTools, mcp_*.go handlers).
model: sonnet
color: blue
---

This skill enforces the ctl-api MCP tool registration pattern for the stateless Streamable HTTP server.

## Steps

1. Pick the domain service under `services/ctl-api/internal/app/<domain>/service/` that already implements `api.Service` and is in FX `group:"services"`.
2. Add `mcp_<name>.go` with an input struct (`json` + `jsonschema` tags) and a handler matching `func(ctx, *mcp.CallToolRequest, in T) (*mcp.CallToolResult, any, error)`.
3. Read org/account/token role from `keys.*FromContext(ctx)` — never Gin context.
4. Return payloads via `api.MCPJSONResult(...)` (compact JSON text content).
5. Register the tool in that domain’s `mcp_register.go` with `mcp.AddTool(server, &mcp.Tool{Name, Description}, handler)`.
6. For mutating tools: prefix `Description` with `WRITE OPERATION:` and call `requireWriteScope(ctx)` at the start of the handler (see installs write tools).
7. Prefer trimmed response shapes over full GORM models when the payload would be large for LLM context.
8. No new MCP-specific FX wiring — `RegisterMCPTools` is discovered via type assert on `api.MCPService`.

Reference examples: `apps/service/mcp_list_apps.go`, `apps/service/mcp_register.go`, `installs/service/mcp_approve_step.go`.

## Anti-Patterns

- **Do not** implement tools only in the CLI proxy — tools live in ctl-api; the CLI forwards them.
- **Do not** skip `requireWriteScope` on write tools — description prefix alone is not enough.
- **Do not** key request state on `Mcp-Session-Id` — the server is `Stateless: true`; use bearer token + `X-Nuon-Org-ID` / `select_org`.
- **Do not** print to stdout from tool handlers — MCP protocol owns the stream.
