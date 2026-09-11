# Nuon CLI

Public `nuon` CLI (Cobra). Talks to `ctl-api`. Config/tokens in `~/.nuon/`.

```bash
cd bins/cli
go build -o nuon .
gofmt ./... && go vet ./...
```

Layout: command definitions under `cmd/`; business logic under `internal/services/` and `internal/` (auth, config, UI). Interactive TUIs live in `internal/ui/v3/` and shared bubbles in `internal/ui/bubbles/`.

Agent surfaces, sync behavior, and TUI conventions below are the rules that matter when changing the CLI.

## TUI Conventions

The CLI uses [Bubble Tea](https://github.com/charmbracelet/bubbletea) for interactive terminal user interfaces. Follow
these conventions when adding or modifying commands:

### TUI vs Non-TUI Command Pattern

**Base command → TUI, subcommands → non-TUI:**

```bash
# TUI - launches interactive interface
nuon installs workflows

# Non-TUI - standard CLI output
nuon installs workflows list
nuon installs workflows get
nuon installs workflows --help
```

**Convention:**

- **Base command** (e.g., `nuon installs workflows`): Launches the full interactive TUI experience
- **Subcommands** (e.g., `list`, `get`, `select`): Non-TUI, outputs JSON/table/text for scripting
- **`--help` flag**: Always non-TUI, shows command documentation
- **`--json` flag**: When available, forces non-TUI JSON output

**Example implementation pattern:**

```go
workflowsCmd := &cobra.Command{
    Use:   "workflows",
    Short: "Manage workflows",
    Long:  `By default, launches an interactive TUI...`,
    Run: func(cmd *cobra.Command, _ []string) {
        // Base command → TUI
        svc.WorkflowsTUI(cmd.Context(), id, workflowID)
    },
}

workflowsListCmd := &cobra.Command{
    Use:   "list",
    Short: "List workflows",
    Run: func(cmd *cobra.Command, _ []string) {
        // Subcommand → non-TUI output
        svc.WorkflowsList(cmd.Context(), id, offset, limit, PrintJSON)
    },
}
workflowsCmd.AddCommand(workflowsListCmd)
```

### Reusing TUI Components

**IMPORTANT**: Always reuse existing TUI components from `internal/ui/`:

| Component        | Location                   | Use Case                                        |
| ---------------- | -------------------------- | ----------------------------------------------- |
| Bubbles (shared) | `internal/ui/bubbles/`     | Selector, confirm dialog, spinner, table        |
| v3 Common        | `internal/ui/v3/common/`   | Progress, header, status line, full-page dialog |
| Workflow TUI     | `internal/ui/v3/workflow/` | Workflow viewing/management                     |
| Action TUI       | `internal/ui/v3/action/`   | Action workflows                                |
| Install TUI      | `internal/ui/v3/install/`  | Install management                              |
| Logs TUI         | `internal/ui/v3/logs/`     | Log streaming                                   |

**Before creating new components:**

1. Check `internal/ui/bubbles/` for reusable primitives (selector, confirm, spinner)
2. Check `internal/ui/v3/common/` for shared layouts (header, footer, progress)
3. Look at existing TUI implementations for patterns (workflow, action, install)

### TUI Structure Pattern

New TUI features should follow the established v3 pattern:

```
internal/ui/v3/<feature>/
├── main.go          # Entry point, Model definition, Init/Update/View
├── keys.go          # Key bindings
├── messages.go      # Custom message types
├── styles.go        # Lipgloss styles
├── actions.go       # Business logic commands
├── data.go          # Data fetching/transformation
├── footer.go        # Footer view component
├── header.go        # Header view component
└── selector/        # Sub-components if needed
```

### How `nuon apps sync` works

The CLI does **not** walk the API's per-resource `Create*Config` endpoints. It parses and validates the config
locally, then hands the whole thing to the API in one shot (`internal/services/apps/sync_push.go`):

1. `POST /v1/apps/:app_id/configs` with `intermediate_config_json` — the serialized parsed config.
2. `POST /v1/apps/:app_id/configs/:config_id/sync` — asks the API to apply it. Returns `202`; the work runs on the
   app's queue.
3. Poll `GET /v1/apps/:app_id/configs/:config_id` until `status` is `active` or `error`, surfacing
   `status_description` as it moves.
4. Read `state.result` off the synced config for the components that had builds scheduled and for resources orphaned
   by this sync, then wait on those builds.

All config-to-database conversion lives server-side in `services/ctl-api/internal/pkg/config/syncer`. **Do not add
config knowledge to the CLI beyond parsing and validation** — a client-side syncer (`pkg/config/sync/apisyncer`) is
exactly what this replaced, after it silently drifted from the server's conversion for months.

#### The app branch path (`default-app-branches`)

When the org has the `default-app-branches` feature flag on, or the user passes `--branch` / `--app-branch`, the sync
routes through an app branch run instead (`internal/services/apps/sync_branch.go`). This path adds **no** endpoints of
its own:

1. `GET /v1/orgs/current` for the flag, then `GET /v1/apps/:app_id/branches` for a branch named `default`. On the first
   sync it does not exist yet, so `POST /v1/apps/:app_id/branches` creates it and
   `POST /v1/apps/:app_id/branches/:branch_id/configs` gives it a single `all_installs` install group. A name collision
   on create means a concurrent sync won the race, so re-list and use theirs.
2. `POST /v1/apps/:app_id/configs` with `intermediate_config_json` and `app_branch_id`. The config is left unsynced.
   **Resolving the branch here as a side effect of an empty `app_branch_id` does not work**: an older CLI would get a
   branch-linked config and then call `/configs/:id/sync`, whose `finalizeAppConfigSync` skips the install rollout for
   branch-linked configs on the assumption a branch run owns it. No run exists, so installs silently never update.
3. `POST /v1/apps/:app_id/branches/:branch_id/runs` with `app_config_id` and `sync_app_config: true`. The run's
   `sync app config` step is what calls the syncer, so the config still moves `pending` to `syncing` to
   `active`/`error` and the status poll above is unchanged.
4. The run's builds step owns component builds. **Do not also call `POST /configs/:id/sync` on this path**: it
   dispatches builds too, and the run would build every changed component twice.
5. The CLI waits until the run's builds step reaches a terminal status, then returns. The install group plan and deploy
   steps that follow keep running server-side.

Exit codes are unchanged: 0 synced, 1 sync failed, 3 builds failed. `--auto-approve` sets `approve-all` on the run;
without it the gate follows the targeted installs' own `approval_option`, which defaults to `prompt`.

### Output format (`--output table|json|agent`)

The global `--output` flag selects the output format (default `table`). `--json`/`-j` is a **deprecated** shorthand for
`--output json` (still works, warns on use, removed in a future release). `resolveOutput` in `cmd/cli.go` resolves the
format with precedence `--output` → `--json` → `NUON_OUTPUT` → `NUON_AGENT` → default, then sets `PrintJSON` and/or
`internal/agentmode` accordingly.

`agent` is the machine-friendly format for LLM/agent callers: it forces non-interactive and turns on
`internal/agentmode`. When on, **stdout carries exactly one JSON envelope** and all progress/human output is routed to
stderr:

- Success: `{"ok":true,"data":<command output>}` — wraps whatever the command passes to `ui.PrintJSON`.
- Error: `{"ok":false,"error":{"code":"<code>","message":"..."}}` with a non-zero exit. Codes come from
  `classifyError` in `internal/ui/agent.go` (`not_found`, `unauthorized`, `forbidden`, `invalid_request`,
  `server_error`, `user_error`, `api_error`, `builds_failed`, `error`).

Commands can attach a stable code and a custom exit code to an error via `ui.ErrExitCode` (honored by `wrapCmd`).
`nuon sync` / `nuon apps sync` use this to split outcomes: exit 0 = synced (+ builds completed), exit 1 = sync
failed, exit 3 (`builds_failed`) = synced but scheduled component builds failed, were policy-blocked, or timed
out. `--no-wait` skips the build wait entirely. The success envelope's `data.builds` reports
`{scheduled, waited, components:[{component_id, component_name, status}]}`.

Routing lives in `internal/agentmode` (`HumanWriter()` returns stderr when enabled). Any new command output must go
through `ui.PrintJSON` / `ui.PrintError` to be enveloped; commands that write their own JSON or use `fmt.Println` bypass
the envelope. `--output json` is unchanged from the old `--json` — raw output, no envelope.

#### Output annotations (REQUIRED when adding commands)

Commands declare which `--output` formats they support via the annotations system (`cmd/annotations.go`). No
annotation = supports all of `table,json,agent`. A command that can't honor a format (TUI-only, raw text, protocol
on stdout like `nuon mcp`) MUST declare what it does support:

```go
Annotations: outputsAnnotation(OutputTable)                          // table only
Annotations: annotations(tuiAnnotation(TUIAltScreen), outputsAnnotation(OutputTable, OutputJSON))
```

`resolveOutput` enforces this — requesting an unsupported format errors with the supported list. **When adding or
changing a command, set this annotation; it is metadata LLMs and completion rely on and is not inferred.**

### OIDC workload identity federation (CI auth without secrets)

Exchange an ambient OIDC ID token for a short-lived Nuon API token (`POST /v1/oidc/token`). Requires
`oidc_federation_enabled` on the control plane and an org trust policy (`nuon orgs oidc-trust-policies create`).

In GitHub Actions: `permissions: id-token: write` plus `NUON_ORG_ID` (and `NUON_API_URL` when not Nuon Cloud). Other CI:
`NUON_OIDC_TOKEN` / `NUON_OIDC_TOKEN_FILE`, or `nuon auth exchange-token`. Implementation: `internal/oidctoken`,
`internal/services/auth/exchange.go`, `tryAmbientOIDCExchange` in `cmd/cli.go`.

### Read-only mode (`--read-only` / `NUON_READ_ONLY=1`)

Safety guardrail for agent-driven use: blocks any command that may mutate remote state. Enforced in
`guardReadOnly` (`cmd/readonly.go`) via a **default-deny allowlist** of read-only leaf command names — a new read
command must be added to `readOnlyCommands` or it will be blocked in this mode. Local-only operations (`select`,
`config`-style targeting, `init`, `generate-config`) are allowed; anything that creates/updates/deletes remote
resources exits 2 with a clear error.

### MCP (`nuon agents mcp`)

Preferred LLM surface is **`nuon agents`**:

- `nuon agents help`: the **human** setup guide (`cmd/agents_help.go`). `agentsSetupGuide` is the single source: it
  backs both this command and the `agents` group's `Long`, so `nuon agents`, `nuon agents --help`, and
  `nuon agents help` all print the same instructions (the subcommand additionally renders the live sign-in, org, and
  resolved MCP URL). Extend the guide, not one of its callers. Document each client on its own (Claude Code, Cursor,
  Amp, and a catch-all that tells people to check that client's MCP docs). Do not present `mcpServers` as a
  universal schema. Every runnable example carries `--allow-writes`. `--url` is a general override when the MCP URL
  does not follow from the API URL (self-hosted and Nuon BYOC are examples).
  `mcpClientJSON` renders a block for a given key; reuse it rather than retyping JSON.
- `nuon agents context` — markdown orientation (auth, selection, MCP URL, timestamps). The document lives in
  `cmd/agents_context.md`, embedded with `go:embed` and rendered as a `text/template` against the fields of
  `agentsContext` (`Authed`, `APIURL`, `MCPURL`, `OrgID`, `AppID`, `InstallID`) — edit the markdown, not Go string
  literals. Keep its tool table and timestamp rules in sync with `docs/guides/agents/tools.mdx`, and its per-client
  registration in sync with `nuon agents help`. Both documents state their purpose up top
  (human vs. agent) because that split is what users get confused about. Do not duplicate
  client setup recipes, sample queries, or the deprecated `nuon mcp` alias here — those live in
  `docs/guides/agents/` and `cmd/mcp.go`. MCP timestamps are UTC RFC3339 (`…Z`); agents localize before naming a
  day or clock time.
- `nuon agents mcp` — stdio proxy to ctl-api MCP (`internal/services/mcpserver/`). Auth from `~/.nuon`
  (`Authorization` + `X-Nuon-Org-ID`). Read-only unless `--allow-writes`. Register with the client:
  `claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes`, `amp mcp add nuon -- nuon agents mcp --allow-writes`,
  or write Cursor `~/.cursor/mcp.json` / `.cursor/mcp.json` with `"args": ["agents", "mcp", "--allow-writes"]` then `agent mcp enable nuon`.

`nuon mcp` is a deprecated alias of `nuon agents mcp`: it proxies only when stdio is piped (a real MCP
client); on a TTY it prints a notice and exits 0.

ctl-api MCP is **stateless** Streamable HTTP (`StreamableHTTPOptions.Stateless: true`):

- **Same as stateful:** `mcp.AddTool`, handler signatures, Bearer auth, write gating
  (`WRITE OPERATION:` + `require.Write`), org via header / `select_org` / single-org auto-select.
- **Different:** no durable `Mcp-Session-Id`; POST-only (GET/DELETE → 405); any replica can serve if
  auth is on the request; no server→client RPCs. Sticky org is in-process by token ID — send
  `X-Nuon-Org-ID` for multi-replica.

stdout is the MCP protocol: handlers and the proxy must never print.

Add API tools with `.agents/skills/mcp-api-tool`. Public docs: `docs/guides/agents/`.

### No-TTY / Non-Interactive Support

The CLI supports non-interactive environments (CI, pipes, cron). All `tea.NewProgram` call sites check `cfg.Interactive`
before launching a TUI.

#### Detection (`internal/config/tty.go`)

Priority: `NUON_NO_TTY=true` → `CI` env var set → `!term.IsTerminal(stdout)` → interactive. Stored in
`Config.Interactive` (resolved once in `NewConfig()`). All service structs access it via `s.cfg.Interactive`.

#### Pattern

Check `interactive` **before** creating a bubbletea program and use a different code path. `ui.NewProgram()`
(`internal/ui/program.go`) exists as a safety net that injects `WithInput(nil)` + `WithoutRenderer()`, but the preferred
pattern is to avoid bubbletea entirely when non-interactive.

#### Fallback behavior by component type

| Component                                                | Non-interactive behavior                                       |
| -------------------------------------------------------- | -------------------------------------------------------------- |
| **Spinners** (`bubbles/spinner.go`, `multi_spinner.go`)  | Print status lines: `Syncing...` → `✓ Syncing... completed`    |
| **Selectors** (`bubbles/selector.go`, v3 selectors)      | Return error: `"interactive terminal required; use --id flag"` |
| **Confirms** (`bubbles/confirm.go`, `confirm_dialog.go`) | Return error: `"use --yes flag to auto-approve"`               |
| **Display TUIs** (`watch`, `workflow`, `logs`)           | One-shot plain-text summary or streaming text output           |
| **Interactive table** (`bubbles/table.go`)               | Render static table via `v.Render()`                           |
| **Action TUIs** (`action/*`, `install/creator`)          | Error: `"interactive terminal required; use --json flag"`      |

#### Command annotations (`cmd/annotations.go`)

Commands are annotated with their TUI type via Cobra's `Annotations` map:

```go
Annotations: tuiAnnotation(TUIAltScreen)    // full-screen TUIs (workflows, watch, logs, actions, create)
Annotations: tuiAnnotation(TUIContextual)   // inline TUI elements (select, dev)
```

Use `annotations()` to merge multiple annotation maps (e.g.,
`annotations(skipAuthAnnotation(), tuiAnnotation(TUIAltScreen))`).

#### Testing

```bash
NUON_NO_TTY=true nuon <command>   # Explicit disable
CI=true nuon <command>             # CI simulation
nuon <command> | cat               # Pipe (auto-detected)
```

