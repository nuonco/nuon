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

Branch runs are `nuon branches`, not `nuon apps sync`.

#### `disable-app-sync`

When the org has `disable-app-sync` on, `nuon apps sync` does not upload the app config (`internal/services/apps/sync_disabled.go`).
It resolves the selected app (or the directory-name app) without the directory mismatch prompt. On a TTY it prints a short
note that app sync now goes through an app branch, then asks whether to create one (default yes). No prints the manual
`nuon branches sync --file branches/<name>.toml` path. Non-interactive, JSON, and agent mode print a short deprecation
error and write nothing.

The wizard reads the app config directory's git checkout (`remote.origin.url`, `HEAD`, and the directory relative to the
repo root) and `GET /v1/vcs/connections/{id}/repos` for each org VCS connection. A matching `full_name` uses
`[connected_repo]`. No match uses `[public_repo]` with the same `owner/repo`. It asks which app branch to write, defaulting
to `main`, with the current git branch (when it is not `main`) or a typed name as the other choices. A detached HEAD omits
the current-branch row. It prints that TOML and asks before writing
`branches/<name>.toml` (a `/` in the branch name is a subdirectory) and running `nuon branches sync` on that file.
Syncing the file reconciles only this branch. A root `branch.toml` already in the folder is refused, because the app
parser rejects `branch.toml` and `branches/` together. A manually managed remote branch with the same name cannot be
taken over.

Afterwards it asks whether to add installs to the branch (default no). On yes, a select list offers all, none, or one
install at a time, and repeats until none or all are chosen. `nuon installs sync` requires `app_branch` in each install
config while this flag is on.

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

### Agent detection

`pkg/agentclient` reads the process environment and names the coding agent that launched the CLI. Product variables win over `AI_AGENT`. Amp is checked first (`AMP_CURRENT_THREAD_ID`, or `AGENT=amp`) because it also sets `CLAUDECODE`. Then Cursor (`CURSOR_AGENT=1`, `CURSOR_INVOKED_AS=agent`, `CURSOR_EXTENSION_HOST_ROLE=agent-exec`, `CURSOR_TRACE_ID`), Gemini (`GEMINI_CLI=1`), Codex (`CODEX_THREAD_ID`, `CODEX_SANDBOX`, `CODEX_CI`), Antigravity (`ANTIGRAVITY_AGENT`), Augment (`AUGMENT_AGENT=1`), OpenCode (`OPENCODE_CLIENT`), Claude Code (`CLAUDECODE=1`, `CLAUDE_CODE=1`, `CLAUDE_CODE_CHILD_SESSION=1`; `CLAUDE_CODE_IS_COWORK` names it `cowork`), Replit (`REPL_ID`), and Copilot (`COPILOT_MODEL`, `COPILOT_ALLOW_ALL`, `COPILOT_GITHUB_TOKEN`). When none of those are set, `AI_AGENT` is the fallback name (`amp`, `claude-code`, `cursor-cli@1.2.3`; the slug before `@` is used). An `AI_AGENT` that is not a slug, such as `1` or `true`, is ignored. `NUON_AGENT_CLIENT=cursor|claude` overrides all of that; `NUON_AGENT_CLIENT=off` disables it. Devin is not detected.

Detection runs once in the persistent pre-run (`applyAgentMode` in `cmd/agent_use.go`). A match sets `Config.Agent` to the name and forces `Config.Interactive = false` for that process. Neither is written to `~/.nuon`. Commands check `cfg.Agent != ""` to take a prompt-free path instead of calling `agentclient.Detect()` again; the state file, REST attribution, and the MCP proxy all read `cfg.Agent`. This does not turn on the JSON envelope — that still needs `--output agent`. `nuon auth login` in this mode skips the deployment selector and URL confirm: it uses the configured `api_url` / `NUON_API_URL`, or `https://api.nuon.co` when none is set, then runs the normal browser sign-in.

Attribution headers are set by `internal/attribution.Apply`. Any code that builds its own `nuon.New` client (login does, twice) must call it, or those requests go out unattributed.

When a client is detected, `nuon agents` and `nuon agents --help` print `✓ agent (cursor) detected` or `✓ agent (claude) detected` at the top of the setup guide. `nuon agents help` prints the orientation markdown instead of that guide. The same name is sent as `X-Nuon-Agent` on control-plane REST requests and on the MCP proxy. A normal terminal does not set these variables, so the line and the header are omitted and `nuon agents help` prints the setup guide.

The first detected run for an agent writes `~/.nuon.agents/<agent>.yaml` (`agent`, `cli_version`, `app_id`, `first_seen`, `last_seen`). That write prints a one-time setup guide on stderr: docs, `nuon agents help`, the dashboard, and a short command flow. Later runs update `cli_version`, `app_id`, and `last_seen` and stay quiet. Agent requests also send `User-Agent: nuon-cli/<version> (<agent>)` and `X-Nuon-Command` set to the command path, with no arguments. The MCP proxy sends command `nuon agents mcp`.

### MCP (`nuon agents mcp`)

Preferred LLM surface is **`nuon agents`**:

- `nuon agents help`: the **human** setup guide (`cmd/agents_help.go`) when `cfg.Agent` is empty, the same check
  `nuon auth login` uses. On an interactive terminal the guide opens in an alt-screen pager (`internal/ui/pager`,
  Bubble Tea viewport). Agents and pipes still get the full text. `agentsSetupGuide` backs this command and the
  `agents` group's `Long`, so `nuon agents`
  and `nuon agents --help` print the same instructions (the help subcommand additionally renders the live sign-in,
  org, and resolved MCP URL). When `cfg.Agent` is set, `nuon agents help` prints the orientation markdown instead.
  That document lives in `cmd/agents_context.md`, embedded with `go:embed` and rendered as a `text/template` against
  the fields of `agentsContext` (`Authed`, `APIURL`, `MCPURL`, `OrgID`, `AppID`, `InstallID`) — edit the markdown,
  not Go string literals. Keep its tool table and timestamp rules in sync with `docs/guides/agents/tools.mdx`, and
  its per-client registration in sync with the setup guide. Both documents state their purpose up top (human vs.
  agent) because that split is what users get confused about. Do not duplicate client setup recipes, sample
  queries, or the deprecated `nuon mcp` alias in the markdown — those live in `docs/guides/agents/` and `cmd/mcp.go`.
  MCP timestamps are UTC RFC3339 (`…Z`); agents localize before naming a day or clock time. Extend the setup guide,
  not one of its callers. Document each client on its own (Claude Code, Cursor, Amp, and a catch-all that tells people to check that client's MCP docs). Do not
  present `mcpServers` as a universal schema. Every runnable example carries `--allow-writes`. `--url` is a general
  override when the MCP URL does not follow from the API URL (self-hosted and Nuon BYOC are examples).
  `mcpClientJSON` renders a block for a given key; reuse it rather than retyping JSON.
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

