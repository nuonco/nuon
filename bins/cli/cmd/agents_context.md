# Nuon agent context

Use this document to orient before creating or changing Nuon resources.

## Current CLI selection

| Field | Value |
| --- | --- |
| Authenticated | {{.Authed}} |
| API URL | `{{.APIURL}}` |
| MCP HTTP URL | `{{.MCPURL}}` |
| Org ID | `{{.OrgID}}` |
| App ID | `{{.AppID}}` |
| Install ID | `{{.InstallID}}` |

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

The proxy sets `X-Nuon-Org-ID`. Do not call `select_org` unless that header is missing.

The upstream URL (`{{.MCPURL}}`) comes from `api_url` in the CLI config. Override with `--url` / `--name` on the registered command (`nuon agents mcp --allow-writes --url … --name …`). A non-default `-C` config goes on the same command.

**Direct HTTP:** point an MCP client at `{{.MCPURL}}` with `Authorization: Bearer <api_token>` and `X-Nuon-Org-ID: <org_id>` (required for multi-org accounts without `select_org`). Client setup: https://docs.nuon.co/guides/agents

## Creating something new (starter checklist)

1. Confirm org/app context (`nuon agents context` / `whoami`).
2. Prefer MCP tools for reads; use `--allow-writes` (or a write-scoped token) only when mutating.
3. Typical create flow:
   - App config in a local directory → `nuon apps sync` (or ask the user to sync).
   - Installs: dashboard or CLI (`nuon installs create`).
   - Deploys / workflow steps → list workflows and pending approvals, then approve/reject/retry/cancel as needed.
4. Do not invent IDs; resolve names via `list_*` / `get_*` tools first.
5. Keep responses trimmed; MCP tools already return compact JSON.

## Timestamps

Tool JSON timestamps are UTC (Zulu) RFC3339 and always end in `Z`, for example `2026-09-04T04:23:00Z`. The `Z` means UTC, not the user's local clock.

- If you name a calendar day, clock time, or age, convert the UTC instant to this machine's local timezone first. Example: `2026-09-04T04:23:00Z` is still the evening of September 3 in US Pacific.
- Never say "today" or "yesterday" from the UTC date digits. The UTC calendar day can be a day ahead of local time.

## Tools (control plane)

Writes are hidden from the stdio proxy unless `--allow-writes` is set. Descriptions start with `WRITE OPERATION:`.

| Domain | Read | Write |
| --- | --- | --- |
| Orgs | `whoami`, `list_orgs`, `select_org` | |
| Apps | `list_apps`, `get_app`, `list_app_branches`, `get_app_branch`, `list_app_branch_runs`, `get_app_branch_run`, `list_app_branch_preview_sources` | `preview_app_branch` |
| Components | `list_components`, `get_component`, `list_builds`, `get_build` | |
| Installs | `list_installs`, `get_install`, `get_install_readme`, `get_install_health`, `list_install_components`, `get_install_inputs`, `list_workflows`, `get_workflow`, `get_workflow_step`, `watch_workflow`, `get_pending_approvals`, `list_deploys`, `get_deploy` | `update_install_inputs`, `deploy_install_components`, `reprovision_install`, `reprovision_sandbox`, `deprovision_install`, `deprovision_sandbox`, `approve_step`, `reject_step`, `retry_step`, `cancel_workflow` |
| Actions | `list_install_actions`, `get_action` | `run_action` |
| Logs | `get_workflow_step_logs`, `get_deploy_logs`, `get_build_logs` | |
| Runbooks | `list_runbooks`, `get_runbook` | `run_runbook` |

Catalog: https://docs.nuon.co/guides/agents/tools
Sample queries: https://docs.nuon.co/guides/agents/sample-queries
