# Nuon (Open Source)

Nuon is a BYOC (Bring Your Own Cloud) platform: software vendors deploy and operate their product in their customers'
cloud accounts. This repository is the open-source control plane, runner, CLI, and dashboard.

Product concepts and deployment options: [README.md](README.md) and https://docs.nuon.co.

When working in a package, read that package's `AGENTS.md`.

## If you are here to use Nuon

The rest of this file is for changing Nuon's code. To deploy an app with Nuon instead:

- Install the CLI: `brew install nuonco/tap/nuon`. Other options are in [docs/cli.mdx](docs/cli.mdx).
- Follow the quickstart: [docs/get-started/quickstart.mdx](docs/get-started/quickstart.mdx), published at
  https://docs.nuon.co/get-started/quickstart.md.
- Every docs page is published as markdown. The index is https://docs.nuon.co/llms.txt; the source is `docs/`.
- App config lives in git and ships through an app branch: add a branch file, run
  `nuon branches sync --file <branch file>` once, then push to the tracked git branch. See
  [docs/guides/app-branches.mdx](docs/guides/app-branches.mdx).

## Architecture

```
CLI / dashboard  →  ctl-api (control plane)  →  runner (in customer cloud)
                         ↓
              Temporal + PostgreSQL + ClickHouse
```

| Piece | Role |
| --- | --- |
| `services/ctl-api` | Control plane API: apps, installs, workflows, builds, authz. Public API (~8081), runner API (~8083). |
| `bins/runner` | Executes deploys/jobs in the customer's cloud; talks to the runner API. |
| `bins/cli` (`nuon`) | Public CLI; primary agent/MCP surface for operators. |
| `services/dashboard-ui` | Vendor dashboard (SPA under `client/`, served by a Go server). |
| Temporal | Orchestrates multi-step install/build/deploy workflows. |
| PostgreSQL | Source of truth for platform state. |
| ClickHouse | Logs, telemetry, and related high-volume reads/writes (e.g. runner heartbeats). |

Typical bug hunt: install/deploy failures → ctl-api workflow + runner logs; sync/config issues →
`services/ctl-api/internal/pkg/config/syncer` (server-side; CLI does not own conversion); auth/org access → ctl-api
account/RBAC models.

## Repository map

| Path | Contents |
| --- | --- |
| `bins/cli` | Public `nuon` CLI |
| `bins/runner` | Customer-cloud execution binary |
| `bins/lsp` | Language server for Nuon configs |
| `bins/telemetry-relay` | Telemetry relay |
| `services/ctl-api` | Control plane (Go) |
| `services/dashboard-ui` | Dashboard UI |
| `pkg/` | Shared Go libraries (API clients, kube, terraform, temporal, etc.) |
| `sdks/` | Generated SDKs (e.g. `nuon-go`, `nuon-runner-go`) |
| `docs/` | Product/API docs (including agent guides) |
| `images/` | Container image build contexts |
| `.agents/skills/` | Agent skills for this repo |

## Where to go next

| Topic | Read |
| --- | --- |
| CLI, `--output agent`, MCP, read-only mode | [bins/cli/AGENTS.md](bins/cli/AGENTS.md), [docs/guides/agents/](docs/guides/agents/) |
| Runner behavior | [bins/runner/AGENTS.md](bins/runner/AGENTS.md) |
| API, workflows, RBAC, signals/webhooks | [services/ctl-api/AGENTS.md](services/ctl-api/AGENTS.md) |
| Dashboard UI | [services/dashboard-ui/AGENTS.md](services/dashboard-ui/AGENTS.md) |
| Logging conventions | [conventions/logging.md](conventions/logging.md) |
| Contributing / issues | [CONTRIBUTING.md](CONTRIBUTING.md) |

`CLAUDE.md` files in this repo are one-line `@AGENTS.md` imports so Claude Code shares the same instructions.

## Agents and MCP

- Prefer the **Nuon MCP** when the client has it configured (read-only by default unless writes are enabled).
- Otherwise use the CLI: `nuon agents help` for orientation (an agent receives the markdown briefing); `nuon agents mcp` as the stdio MCP proxy;
  `--output agent` for a single JSON envelope on stdout; `--read-only` / `NUON_READ_ONLY=1` unless the task needs
  writes.
- Public walkthrough: https://docs.nuon.co and `docs/guides/agents/` in this repo.
- Auth: `nuon auth login`, then select org (and app/install as needed).

## Conventions

- In examples, fixtures, commits, and PRs, use fictional names (`acme`, `example.com`) or resource IDs — not real
  customer org, app, install, or domain names.
- Database-backed tests use PostgreSQL, not SQLite.
- After Go edits: `gofmt` and `goimports` on the package/directory (do not hand-manage imports).
- Prefer struct-based GORM `Where` clauses over raw SQL strings.
- Log with zap (see conventions above), not `fmt.Println`.
- Comment only when the *why* is non-obvious.
- After changing annotated types, Temporal `@temporal-gen` surfaces, or swagger endpoints: `go generate` in the
  affected package (see `services/ctl-api`, `sdks/`).
- Cloud-connection ACR tokens always use Entra scope `https://containerregistry.azure.net/.default`
  (`acr.RegistryScope`). That covers connection verification, the foreign-subject probe, and
  `GetRepositoryTokenWithCredential`, which posts the access token with `ConnectionExchangeClient`.
  Do not request an Azure Resource Manager management scope (`https://management.azure.com/` or
  `https://management.azure.com/.default`) on that path. Legacy `GetRepositoryToken` keeps
  `armTokenScope` and the default HTTP client so existing runners behind a proxy keep working.
