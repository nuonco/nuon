# CTL-API Service

The **Control API (ctl-api)** is the Nuon control-plane backend: apps, components, installs, runners, workflows, and
org-scoped operations. Go service using Gin, PostgreSQL/GORM, Temporal, and FX dependency injection.

## API Surfaces

Each surface is a separate HTTP listener (defaults in `internal/config.go`). Run all locally with `go run . api` from
`services/ctl-api`.

| Server | Port | Route registration | Auth | Purpose |
|--------|------|--------------------|------|---------|
| Public API | 8081 | `RegisterPublicRoutes` | API key + `X-Nuon-Org-ID` | Customer API, CLI, `dashboard-ui` |
| Internal API | 8082 | `RegisterInternalRoutes` | Internal/admin middleware | Ops JSON API |
| Runner API | 8083 | `RegisterRunnerRoutes` | Runner token | Runner job/state callbacks |
| Auth API | 8084 | `RegisterAuthRoutes` | Session/OAuth/device flow | Login, device code, OAuth |
| Admin Dashboard | 8087 | `RegisterAdminDashboardRoutes` | Admin proxy + session cookie | React SPA + JSON BFF |
| MCP | 8088 | `RegisterMCPTools` on domain services | Bearer + org RBAC | Agent MCP (Streamable HTTP) |
| Slack API | 8089 | `RegisterSlackRoutes` | Slack signing / OAuth JWT | Slash commands, Events API |
| nuonctl MCP | 8091 | `RegisterMCPTools` on `nuonctl_mcp_services` | Bearer + employee | Employee-only nuonctl MCP |

Health: `/livez`, `/readyz` on each Gin server. Despite the method name, most authenticated customer routes live on the
public server via `RegisterPublicRoutes` at `/v1/*`. `RegisterAuthRoutes` is for the auth listener, not org-scoped API
auth.

Swagger: public server serves `/docs/*`, `/oapi/v2`, `/oapi/v3`. Specs under `docs/public/`, `docs/admin/`, `docs/runner/`.

## Stack and Consumers

- **HTTP:** Gin + per-listener middleware
- **DB:** PostgreSQL via GORM (`*gorm.DB` tagged `name:"psql"`)
- **Workflows:** Temporal (`go run . worker`)
- **Docs:** Swag annotations; regenerate with `go generate` (see below)

Primary consumers: `dashboard-ui`, `bins/cli` (`nuon`), `bins/runner`, Temporal workers in this service.

## Repository Layout

```
services/ctl-api/
├── main.go / public.go / admin.go / runner.go   # entry + swag meta
├── cmd/                                         # cobra: api, worker, consumer, mcp, gen, ...
├── internal/
│   ├── app/          # domain models + service|helpers|worker|signals
│   ├── pkg/          # api, authz, config, queue, flow, ...
│   ├── middlewares/
│   └── fxmodules/    # FX wiring (apis, services, helpers, workers, mcp)
├── docs/
└── internal/integration/
```

### Domain package layout

Under `internal/app/{domain}/`:

```
service/     # HTTP handlers; implements api.Service Register*Routes
helpers/     # cross-domain business logic (FX-provided)
worker/      # Temporal workflows and domain activities
signals/     # queue signal definitions (often signals/v2/)
```

Root models sit in `internal/app/*.go`.

### FX modules (`internal/fxmodules/`)

| Module | Role |
|--------|------|
| `apis.go` | One FX module per HTTP listener |
| `services.go` | `fx.Provide(api.AsService(domain.New))` |
| `helpers.go` | Domain helpers |
| `middlewares.go` | HTTP middleware providers |
| `workers_*.go` | Temporal worker registration |
| `mcp.go` | MCP HTTP server lifecycle |

## Route Registration (FX + `api.Service`)

Every HTTP domain implements `internal/pkg/api.Service` with `RegisterPublicRoutes`, `RegisterRunnerRoutes`,
`RegisterInternalRoutes`, `RegisterAuthRoutes`, `RegisterAdminDashboardRoutes`, and `RegisterSlackRoutes`.

1. Implement handlers under `internal/app/<domain>/service/`
2. Wire `var _ api.Service = (*service)(nil)` and `Register*Routes` (no-op unused contexts)
3. Add `fx.Provide(api.AsService(<domain>service.New))` in `internal/fxmodules/services.go`
4. Ensure the target listener's FX module includes that services module

**Routes are not registered unless the service is in `services.go`.** Wrong `Register*Routes` method means the route
never mounts on the intended port.

Optional MCP: implement `api.MCPService` with `RegisterMCPTools(*mcp.Server)`.

### Handler placement

- **Handlers:** parse/bind, read `cctx`, return JSON/errors
- **Private service methods:** domain-only logic on the same `service` struct
- **Helpers:** logic shared across domains (inject via `fx.In` on `Params`)

Keep handlers thin. Cross-domain side effects (e.g. journey updates) go through helpers and must not fail the primary
operation (log and continue).

## App Config Sync (IMPORTANT)

`internal/pkg/config/syncer` is the **only** implementation that turns app config into database records. All paths call
`syncer.Run` against the intermediate config on `AppConfig`.

| Entry point | Flow |
|-------------|------|
| CLI `nuon apps sync` | `POST /configs` → `POST /configs/:id/sync` → `appconfigsync` signal |
| CLI with default app branches | `POST /configs` + `app_branch_id` → branch run `sync_app_config` step |
| VCS branch sync | branch run fetch step → `branches/activities.syncAppConfig` |

The CLI does not walk per-resource `Create*Config` endpoints. Those remain public API but are not the sync path.

**Build scheduling:** `syncer.RunRequest.DispatchBuilds` is set on the standalone CLI sync path. Branch runs leave it
off; their builds step schedules builds. Enabling both would double-build.

**Single builder:** Config maps to `app.*` models only through `internal/pkg/config/build` (pure: no gorm/gin).
Validation lives in `internal/pkg/config/validation`, invoked from the builder.

When adding config fields:

1. Mapping in `internal/pkg/config/build`; caller resolves DB lookups
2. Validation in `internal/pkg/config/validation`
3. Wire HTTP handler **and** syncer step
4. Do not change `Create*Request` JSON shapes; map onto builder input
5. Cover the field in `build/build_test.go`

Never add a second sync implementation. `// Duplicates logic from ...` means extract to `build`.

## PostgreSQL, GORM, and Tests

- **No SQLite** for database-backed tests. Use PostgreSQL harnesses.
- **Struct-based `Where` clauses** — not raw SQL strings for GORM filters.
- Prefer one `Preload()` chain over multiple Temporal round-trips when relationships exist. Prefer pinned FKs over
  `ORDER BY created_at DESC LIMIT 1` rediscovery.

## Logging

Never use `fmt.Println`. See [conventions/logging.md](/conventions/logging.md).

| Context | Logger |
|---------|--------|
| HTTP services | `*zap.Logger` via FX (`s.l`) |
| Temporal workflows | `log.WorkflowLogger(ctx)` from `internal/pkg/log` |
| Temporal activities | See [activities AGENTS.md](internal/pkg/workflows/workflow/activities/AGENTS.md) |

## Auth and Multi-Tenancy (essentials)

- Org-scoped routes require `X-Nuon-Org-ID` unless marked global in `internal/middlewares/global/global.go`
- Global allowlist patterns must **exactly** match route registration (`:param` segments, not literal paths)
- RBAC: accounts → roles → policies in `internal/pkg/authz/`
- **`cctx.SetAccountContext(ctx, account)`** before authz/org mutations so `CreatedByID` hooks populate

## Workflow and Status Conventions

- Never use `step.Idx` in user-facing status strings; use `step.Name`.
- Update `CompositeStatus.Metadata` separately from status transitions (use `generics.MergeJSONBMetadata`).
- Feature flags (`internal/app/org.go`): add constant, **append** to bottom of `GetFeatures()`, add description, set
  default in `BeforeCreate`.

## MCP Server

Stateless Streamable HTTP MCP at port 8088 (`internal/app/mcp/server/`). Employee-only nuonctl MCP is a second listener in the same `api` process on port 8091 (`NewNuonctl`). Standalone: `go run . api-mcp` and `go run . nuonctl-mcp-api`.

- Auth: Bearer, `require.Read` / `require.Write`, org via `X-Nuon-Org-ID` / `select_org`
- Register with `api.MCPReadTool` / `api.MCPWriteTool`
- No durable session; prefer org header in multi-replica
- Timestamps: UTC RFC3339 `Z`

Adding a tool: [.agents/skills/mcp-api-tool/SKILL.md](../../.agents/skills/mcp-api-tool/SKILL.md). Keep
`docs/guides/agents/tools.mdx` and CLI agent context in sync.

## Swagger and Code Generation

- Handlers need swag annotations (`@Router`, `@Security APIKey`, `@Security OrgID` for authenticated public routes)
- Description markdown under `docs/*/descriptions/` — missing files cause startup parse errors
- Regenerate: `go generate ./services/ctl-api/...`
- Long swagger model names in generated SDKs are expected

## Development Commands

```bash
cd services/ctl-api
go run . api              # all listeners + MCP
go run . worker           # Temporal workers
go run . preflight        # config/deps check
```

After Go edits: `gofmt` and `goimports` on the package directory.

## Nested AGENTS.md

| Area | Path |
|------|------|
| Admin dashboard SPA + BFF | [internal/app/admin-dashboard/AGENTS.md](internal/app/admin-dashboard/AGENTS.md) |
| Queue/signal system | [internal/pkg/queue/AGENTS.md](internal/pkg/queue/AGENTS.md) |
| Workflow step execution | [internal/pkg/flow/AGENTS.md](internal/pkg/flow/AGENTS.md) |
| Shared Temporal activities | [internal/pkg/workflows/workflow/activities/AGENTS.md](internal/pkg/workflows/workflow/activities/AGENTS.md) |

## Signals and Notifications

New signals under `internal/app/**/signals/` must also wire webhook and Slack hooks in
`internal/pkg/queue/signal/hooks/`, and stay in sync with dashboard/CLI interest pickers when users configure
subscriptions. See [queue AGENTS.md](internal/pkg/queue/AGENTS.md) for queue/signal ownership patterns.

## Integration Tests

`internal/integration/` — HTTP-level tests against PostgreSQL. Mock persistence in unit tests; do not use SQLite for
SQL semantics.
