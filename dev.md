# Local Development Environment

The local dev environment uses **Nix flake** for dependencies and **process-compose** to orchestrate all services. No Docker required for the core stack.

## Prerequisites: Nix + direnv

You need **Nix** (with flakes enabled) installed on your machine. We recommend the [Determinate Systems Nix Installer](https://docs.determinate.systems/) — follow their [installation instructions](https://docs.determinate.systems/getting-started/individuals) for your platform. It enables flakes by default and is easy to uninstall.

Next, install **direnv** so the Nix dev shell is loaded automatically when you `cd` into the repo. See the [direnv installation docs](https://direnv.net/docs/installation.html) for your platform, then [hook it into your shell](https://direnv.net/docs/hook.html).

Then set up your local env files (both are gitignored — copy them from the committed examples and fill in any secrets):

```bash
cp .envrc.example .envrc
cp .env.dev.example .env.dev   # then fill in the `replace-me` secret placeholders
direnv allow
```

`.envrc` runs `use flake` and `dotenv_if_exists .env.dev`, so after `direnv allow` the Nix dev shell and your env vars load automatically whenever you enter the directory — no need to run `nix develop` manually.

## Quick Start

```bash
# Enter the nix dev shell (installs all dependencies)
nix develop

# Start the full stack
process-compose up
```

A `preflight` process runs first and kills any stale processes on dev ports from a previous crashed session. Then all services start in dependency order.

## Architecture

```
process-compose up
├── postgres-init → postgres → postgres-createdb
├── clickhouse → clickhouse-createdb
├── temporal (9 namespaces)
├── go-generate (./scripts/reset-generated-code — runs once)
│
├── ctl-api-startup (migrations, depends on all infra + go-generate)
│   ├── ctl-api (air hot-reload, :8081/:8082/:8083/:8084/:8087/:8089)
│   ├── ctl-api-worker (air hot-reload, :8086)
│   └── runner (air hot-reload, run-local org, :9090)
│
├── dashboard-ui-deps (bun install if needed)
│   ├── dashboard-ui (bun build watch + PostCSS + Bun dev-server SSE reload, :4001 → proxies :4000)
│   └── dashboard-ui-server (Go BFF, :4000 — serves SPA + reverse-proxies /v1/*)
│
├── admin-dashboard-deps (bun install if needed)
│   └── admin-dashboard (bun-run esbuild watch + BrowserSync, :9088 → proxies :8087)
│
└── codegen namespace (disabled, on-demand)
    ├── gen-temporal
    ├── gen-api
    ├── gen-mocks
    ├── gen-sdk
    └── gen-all
```

## Service URLs

| Service | URL |
|---|---|
| Dashboard UI (Go BFF) | http://localhost:4000 |
| Dashboard UI (Bun dev-server live-reload) | http://localhost:4001 |
| CTL-API Public | http://localhost:8081 |
| CTL-API Admin/Internal | http://localhost:8082 |
| CTL-API Runner | http://localhost:8083 |
| CTL-API Auth | http://localhost:8084 |
| CTL-API Worker readiness | http://localhost:8086/readyz |
| Admin Dashboard (BFF + SPA) | http://localhost:8087 |
| Admin Dashboard (browser-sync live-reload) | http://localhost:9088 |
| CTL-API Slack listener | http://localhost:8089 |
| Runner health | http://localhost:9090/health |
| Temporal UI | http://localhost:8233 |
| PostgreSQL | localhost:5432 |
| ClickHouse (native) | localhost:9000 |
| ClickHouse (HTTP) | localhost:8123 |

## Key Files

| File | Purpose |
|---|---|
| `flake.nix` | Nix dev shell with all dependencies (Go, Node.js, air, process-compose, terraform, etc.) |
| `process-compose.yml` | Orchestrates all services and on-demand codegen processes |
| `.env.dev` | Environment variables for ctl-api (sourced by nix shell) |
| `.air-api.toml` | Air hot-reload config for ctl-api |
| `.air-worker.toml` | Air hot-reload config for ctl-api-worker |
| `.air-runner.toml` | Air hot-reload config for runner |
| `.dev/` | Local data directory (gitignored) — postgres data, clickhouse data, temporal DB, air build artifacts |

## Hot Reload

### Go services (ctl-api, ctl-api-worker, runner)

Managed by **air**. On any `.go` file change in watched directories:
1. Air detects the change
2. Runs temporal code generation (`go run ./services/ctl-api/cmd/gen --targets temporal`) — for ctl-api and ctl-api-worker only
3. Builds a binary (`go build`)
4. Restarts the service

**Watched directories:**
- ctl-api & ctl-api-worker: `services/ctl-api/`, `pkg/`
- runner: `bins/runner/`, `pkg/`

**Excluded from watch:** `_gen.go` and `mock*.go` files (prevents rebuild loops from generated code).

### Dashboard UI

Managed by **bun** (`bun run dev`): Bun build watch + PostCSS watch + a Bun dev-server that proxies `:4001 → :4000` and triggers SSE-based browser reloads. Changes to files in `services/dashboard-ui/client/` are automatically rebuilt and the browser refreshes. The Go BFF (`dashboard-ui-server`) serves the compiled SPA on `:4000`.

## Code Generation

### Automatic (on startup)

`./scripts/reset-generated-code` runs once on `process-compose up` via the `go-generate` process. This generates everything:
- Temporal workflow/activity types
- Swagger/API docs
- All mocks (mockgen)
- Config schema docs
- SDK clients (nuon-go, nuon-runner-go)

### Automatic (on hot reload)

Temporal code generation (`go run ./services/ctl-api/cmd/gen --targets temporal`) runs automatically before every ctl-api and ctl-api-worker rebuild. This means changes to temporal annotations are picked up without manual intervention.

### On-demand (trigger when needed)

Trigger from another terminal or from the process-compose TUI (`F7`):

```bash
# Temporal workflows/activities only
process-compose process start gen-temporal

# Swagger/API types only
process-compose process start gen-api

# All mocks (mockgen)
process-compose process start gen-mocks

# SDKs (nuon-go + nuon-runner-go) — also regenerates API specs
process-compose process start gen-sdk

# Full regeneration (nuclear option)
process-compose process start gen-all
```

**When to use each:**

| Generator | Trigger when... |
|---|---|
| `gen-temporal` | You add/modify `@temporal-gen` annotations |
| `gen-api` | You change swagger annotations or API endpoint definitions |
| `gen-mocks` | You modify an interface in `pkg/` that has a `//go:generate mockgen` directive |
| `gen-sdk` | API spec changed and you need updated Go/Runner SDK clients |
| `gen-all` | Things are broken, generated files are out of sync, or you pulled new code |

After any on-demand generation, air automatically detects the new/changed `.go` files and rebuilds the affected services.

## Local Debugging

### PostgreSQL

```bash
psql -h 127.0.0.1 -p 5432 -U ctl_api -d ctl_api

# Example query
psql -h 127.0.0.1 -p 5432 -U ctl_api -d ctl_api \
  -c "select id, status, status_description from runner_jobs order by created_at desc limit 5;"
```

### ClickHouse

```bash
clickhouse-client --port 9000

# Example query
clickhouse-client --port 9000 \
  --query "select timestamp, severity_text, body from ctl_api.otel_log_records order by timestamp desc limit 50"
```

### Temporal

```bash
# List workflows
temporal workflow list --namespace components --address 127.0.0.1:7233

# Show workflow details
temporal workflow show --namespace components --workflow-id <workflow_id> --run-id <run_id> --output json
```

## Environment Variables

The `.env.dev` file is sourced by the nix shell and provides stub values for all `validate:"required"` config fields. Key settings:

- `FORCE_SANDBOX_MODE=true` — bypasses real cloud infrastructure
- `DISABLE_NOTIFICATIONS=true` / `DISABLE_ANALYTICS=true` — no external calls
- Database connections point to `127.0.0.1` (local process-compose services)
- Cloud infra stubs (management account, runner cluster, DNS) are placeholder values

## Cleaning Up

```bash
# Stop all services
process-compose down

# Wipe all local data and start fresh
rm -rf .dev/
```

## Troubleshooting

### Port already in use
The `preflight` process in process-compose.yml handles this automatically — it kills stale processes on all dev ports before starting. If it doesn't work for some reason, kill leftovers manually:
```bash
lsof -ti:5432 | xargs kill  # postgres
lsof -ti:9000 | xargs kill  # clickhouse
lsof -ti:7233 | xargs kill  # temporal
```

### Generated code out of sync
```bash
process-compose process start gen-all
```

### Air not picking up changes
Check that the file you edited is in a watched directory and isn't excluded by `exclude_regex`. Generated files (`_gen.go`, `mock*.go`) are intentionally excluded.

### ctl-api startup fails
Check that postgres, clickhouse, and temporal are all healthy in the process-compose TUI. The `ctl-api-startup` process runs migrations and requires all three.
