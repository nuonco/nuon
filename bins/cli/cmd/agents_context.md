# Nuon agent context

**Purpose:** this document is for you, the agent. `nuon agents help` prints it when it detects an agent. It templates out the current CLI selection, how to reach Nuon over MCP, the tool catalog, and links for everything else Nuon. Read it before creating or changing Nuon resources.

A person running `nuon agents help` in a normal terminal gets the setup guide instead of this document. Point them at that rather than improvising setup steps.

## Current CLI selection

| Field | Value |
| --- | --- |
| Authenticated | {{.Authed}} |
| API URL | `{{.APIURL}}` |
| MCP HTTP URL | `{{.MCPURL}}` |
| Org ID | `{{.OrgID}}` |
| App ID | `{{.AppID}}` |
| Install ID | `{{.InstallID}}` |
{{if .ConfigFile}}| CLI config (`-C`) | `{{.ConfigFile}}` |
{{end}}

{{if .OpenAPISchemaURL}}
## HTTP API schema

When you need the HTTP API calls you can make, fetch the OpenAPI schema at `{{.OpenAPISchemaURL}}`.

{{end}}
If auth or org is missing:

```bash
nuon auth login
nuon orgs select
nuon apps select
nuon installs select   # optional
```

## How to call MCP

**Preferred:** local stdio proxy. It injects the token and org ID from `~/.nuon` and forwards tools to the control plane:

```bash
nuon agents mcp --allow-writes
```

The proxy sets `X-Nuon-Org-ID` from the CLI config org. Pass `x_nuon_org_id` on a tool call to send a different org for that request only. The proxy does not store it. There is no `select_org` tool.

## Organization for this session

Before the first org-scoped tool call, tell the user the Org ID in the current CLI selection and ask whether this session should use it or a different org, then wait. When that field is `(none selected)`, ask which org this session should use. Do not print the full org list. If the user names an org, resolve it with `list_orgs` and `query`. If they want to browse, show one page and say when `has_more` is true. Do not keep paging unless they ask for everything. After they answer, pass that `x_nuon_org_id` on every later org-scoped call in this conversation, including when it matches the config org. Do not write the choice back to the CLI config. `whoami` and `list_orgs` work without an org.

### Registering the proxy with a client

Nuon MCP uses stdio. The client starts a local process and exchanges MCP messages on that process's standard input and output. Every MCP connection should spawn `nuon` with arguments `agents`, `mcp`, and `--allow-writes`.

**Claude Code.** `claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes`, or `.mcp.json` in a project:

```json
{
  "mcpServers": {
    "nuon": {
      "command": "nuon",
      "args": ["agents", "mcp", "--allow-writes"]
    }
  }
}
```

**Cursor.** No add command. Save to `~/.cursor/mcp.json` (all projects) or `.cursor/mcp.json` (one project), then run `agent mcp enable nuon`:

```json
{
  "mcpServers": {
    "nuon": {
      "command": "nuon",
      "args": ["agents", "mcp", "--allow-writes"]
    }
  }
}
```

**Amp.** `amp mcp add nuon -- nuon agents mcp --allow-writes` (add `--workspace` to scope it to one workspace), or `~/.config/amp/settings.json` (user) / `.amp/settings.json` (workspace). Amp's key is `amp.mcpServers`:

```json
{
  "amp.mcpServers": {
    "nuon": {
      "command": "nuon",
      "args": ["agents", "mcp", "--allow-writes"]
    }
  }
}
```

**Other clients.** Use that client's MCP docs for the file path and JSON key. Point it at the `nuon` binary with arguments `agents`, `mcp`, and `--allow-writes`.

### Overriding the MCP URL

The upstream URL (`{{.MCPURL}}`) is derived from `api_url` in the CLI config, turning `api.<host>` or `app.<host>` into `mcp.<host>/mcp`. `https://app.nuon.co` is the Nuon Cloud dashboard, and login accepts it as an API URL, so it resolves to `https://mcp.nuon.co/mcp`. Pass `--url` on the registered command when the MCP URL does not follow from the API URL, for example a self-hosted or Nuon BYOC control plane, or any deployment where the two hostnames differ. Pass `--name` to rename the server in the client's list. A non-default `-C` config goes on the same command.

## Creating something new (starter checklist)

1. Confirm the session org from the table above. `whoami` reports the org on that request and an org count. Use `list_orgs` to find a different org.
2. Prefer MCP tools for reads; use `--allow-writes` (or a write-scoped token) only when mutating.
3. Typical create flow:
   - App config in a local directory → `nuon apps sync` (or ask the user to sync).
   - Installs: dashboard or CLI (`nuon installs create`).
   - Deploys, runbooks, and actions → the `run_runbook` description requires a prompt before the call: show every step from `get_runbook` as a checklist (number, name, type), all enabled, and let the user turn steps off. Omit `steps` to run all of them. To skip any, pass every `step_id` with `enabled` false for the ones to skip; a step left out of `steps` still runs. Offer an optional execution role from `list_available_roles` (`operation_type=trigger`), or omit `role` for the default. After `run_runbook`, `run_action`, `reprovision_install`, `reprovision_sandbox`, `deprovision_install`, `deprovision_sandbox`, `deploy_install_components`, `update_install_inputs`, or `preview_app_branch`, ask whether the user wants to watch and call `next_action` only if they do. Each watch returns when a step starts or finishes. After every watch result, the next user-visible message is `step_progress`, then call watch again with the returned `cursor`. A row of watch calls with no `step_progress` between them is wrong. Do not end the turn after printing the list. Only stop when the workflow is terminal or a step needs a decision. Runbooks and action runs are install workflows. When you stop for user input (approval, retry, or any `next_actions` menu), do not report only `completed_steps`/`total_steps`. List every step from the `steps` array by name and status so the user can see what already finished before the waiting step. Keep that list compact (name and status only). Present each waiting approval's `next_actions` as a menu. Plan review is optional: when the user chooses to review, call `get_approval_plan_diff` and follow `continue_review` until every page is loaded, then review those diffs together. Selecting approve, reject, or approve all is confirmation; call the selected tool immediately without loading diffs or asking again. Offer `approve_all` only when that action is present, and explain before presenting it that it applies every remaining plan in the workflow. Then watch the workflow for the next step. When a step includes `stack_setup`, show that install-stack setup (quick launch link, template URL, CLI commands, and Terraform files when present) and keep watching until the stack is applied. When a step includes `composite_error`, show its message, type, and each section before the retry menu. When a step is waiting for a retry, `next_actions` lists every menu choice: Show step logs, Retry step when the step is retryable, Skip step when the step is skippable, and Cancel workflow. Present every one of those actions. Call `get_workflow_step_logs` when presenting the menu. That does not retry, skip, or cancel. Selecting retry, skip, or cancel is confirmation: call that action's tool immediately, then watch again.
4. Do not invent IDs; resolve names via `list_*` / `get_*` tools first.
5. Keep responses trimmed; MCP tools already return compact JSON.

## Install summary flow

Use `install_summary` as the entry point for investigating or updating one install. Do not call `get_install` for an install summary. If the user did not name an install, ask whether to use the Install ID in the current CLI selection or pick one from `list_installs`, and wait for the answer. When that field is `(none selected)`, skip the current-install choice, call `list_installs`, and let them pick. A named install skips this question. Render the summary like the install detail page, in this order: identity, applied commit (and the expected commit when it differs), each component, health, any failing workflows, then inputs grouped by name with their redacted values. The commit is already in the page. Then present `next_actions` and wait: Health, Resources, Deployments, an update, a failing workflow only when one is listed, or no action. Do not chain every inspection automatically.

The Make an update choice has nested `next_actions`. Present those choices and wait again. They prepare existing write flows by loading current inputs or available IAM roles; opening either menu does not mutate the install. After an inspection, offer Back to install summary or Do nothing. Back calls `install_summary` again with the same install.

## Timestamps

Tool JSON timestamps are UTC (Zulu) RFC3339 and always end in `Z`, for example `2026-09-04T04:23:00Z`. The `Z` means UTC, not the user's local clock.

- If you name a calendar day, clock time, or age, convert the UTC instant to this machine's local timezone first. Example: `2026-09-04T04:23:00Z` is still the evening of September 3 in US Pacific.
- Never say "today" or "yesterday" from the UTC date digits. The UTC calendar day can be a day ahead of local time.

List tools are paginated (default 20, max 100). If `has_more` is true, tell the user there are more results. Do not keep paging until the list is complete unless they asked for everything; use `offset=next_offset` for the next page.

## Nuon CLI commands

{{.CLIInstructions}}

## IAM permission checks (local CLI extension)

These checks are not MCP tools. They run on the app directory via the `policies` CLI extension.

After authoring or changing `permissions/` or `break_glass/`, before syncing:

1. If `nuon policies --help` fails: `nuon ext install nuonco/nuon-ext-policies`
2. From the app directory (or pass `--app-dir`):
   - `nuon policies --output json check-overlap <role>.toml` for each role file (for example `maintenance.toml`)
   - `nuon policies --output json check-boundaries`
3. Fix overlapping IAM actions and high-severity boundary drift (maintenance-only actions missing from provision/deprovision) before syncing.

## Tools (control plane)

Writes are hidden from the stdio proxy unless `--allow-writes` is set. Descriptions start with `WRITE OPERATION:`.

`get_install_inputs` redacts values for inputs marked sensitive. It is safe to use for inspection, but do not use its redacted values to construct an update.

| Domain | Read | Write |
| --- | --- | --- |
| Orgs | `whoami`, `list_orgs` | |
| Apps | `list_apps`, `get_app`, `get_app_config_schema`, `list_app_branches`, `get_app_branch`, `list_app_branch_runs`, `get_app_branch_run`, `list_app_branch_preview_sources` | `preview_app_branch` |
| Components | `list_components`, `get_component`, `list_builds`, `get_build` | |
| Installs | `list_installs`, `install_summary`, `get_install`, `get_install_readme`, `get_install_health`, `list_install_resources`, `list_install_components`, `get_install_inputs`, `list_available_roles`, `list_workflows`, `get_workflow`, `get_workflow_step`, `watch_workflow`, `get_pending_approvals`, `get_approval_plan_diff`, `list_deploys`, `get_deploy` | `update_install_inputs`, `deploy_install_components`, `reprovision_install`, `reprovision_sandbox`, `reconcile_install_app_version`, `deprovision_install`, `deprovision_sandbox`, `approve_step`, `approve_all`, `reject_step`, `retry_step`, `skip_step`, `cancel_workflow` |
| Actions | `list_install_actions`, `get_action` | `run_action` |
| Logs | `get_workflow_step_logs`, `get_deploy_logs`, `get_build_logs` | |
| Runbooks | `list_runbooks`, `get_runbook` | `run_runbook` |

Catalog: https://docs.nuon.co/guides/agents/tools
Sample queries: https://docs.nuon.co/guides/agents/sample-queries
