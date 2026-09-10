# Dashboard UI Service

Nuon's primary web application: a **Go BFF + React SPA**. All production UI work goes in `client/`. There is no `src/`
directory and no Next.js app in this service.

```
services/dashboard-ui/
├── client/     ← Production SPA + Lite rebuild (`lite/`)
├── server/     ← Go BFF (Gin + Uber fx)
└── dist/       ← Compiled SPA assets served by the BFF
```

Entry: `client/index.tsx` (production `App` or `LiteApp` from runtime config).

**Working in `client/lite/`?** Follow [client/lite/AGENTS.md](./client/lite/AGENTS.md) and ignore this file. Production
code must not import from `client/lite/`.

## Go BFF (`server/`)

- Serves the compiled SPA from `dist/`
- Validates the `X-Nuon-Auth` httponly cookie
- Injects runtime config as `window.__NUON_CONFIG__`
- Reverse-proxies `/v1/*` to ctl-api (cookie → `Authorization: Bearer`)
- Exposes BFF-only `/api/*` (SSE streams, log download, etc.)

| Path prefix | Role |
|-------------|------|
| `/v1/*` | Proxied to ctl-api |
| `/api/*` | Handled by the BFF |

Log streams (`server/internal/handlers/log_streams.go`): SSE and download endpoints; `?job_output=true` filters to
`ScopeName == "oteljob"`. Runner OTEL scopes: `oteljob` (user-visible) and `system` (internal).

### SSE resource endpoints

~13 `/api/orgs/:orgId/.../sse` endpoints poll ctl-api, hash JSON, and emit events only on change.

**All shared plumbing is in `server/internal/handlers/sse.go`. Never hand-roll an SSE loop.** Auth with `sseAuth`, then
pass a `Fetch` closure to `runSSEStream`.

- First event in `Events` is the primary resource (`Finished` keys off it)
- `sseAuth` before `runSSEStream` — after SSE headers flush, errors are `fetch-error` events
- Wrap marshal failures with `errSSESilentRetry` for silent retry
- Paginated lists: `timelineQuery` + `timelineFetcher` from `timeline.go`
- `log_streams.go` is intentionally separate — do not fold into `runSSEStream`

## Production SPA (`client/`)

```
client/
├── components/     ← Reusable UI (domain + common/, layout/, surfaces/)
├── hooks/
├── lib/api.ts      ← Fetch wrapper (returns T, throws TAPIError)
├── lib/ctl-api/    ← Domain API functions
├── providers/
├── types/ctl-api.types.ts   ← Import these (not nuon-oapi-v3.d.ts)
├── views/          ← Route-level views only
└── index.tsx
```

### Views vs components (strict)

- **`views/`** — route content, layout wrappers, orchestration only
- **`views/` must never contain** modals, tables, reusable sub-components, or action buttons
- Feature UI belongs in `client/components/[domain]/`

## Layout system

Use sanctioned scaffolds — see [DESIGN.md](./DESIGN.md) §5.

| Archetype | Scaffold |
|-----------|----------|
| Org-level list | `ListPage variant="page"` |
| Child list via `<Outlet />` | `ListPage` default `variant="section"` |
| Non-list section | `PageSection` + `SectionHeader` |
| Run page | `DetailPage` + `DetailHeader` + routed `TabNav` |
| Entity page | `DetailPage` + `HistoryRail` |

Resource identity header → `DetailHeader`. Section name only → `SectionHeader`. `PageTitle` / `Breadcrumbs` are
headless setters rendered as siblings **before** the scaffold.

Do not: import `<BackToTop />` in views; use `!p-0` hacks (use `flush`); hand-assemble headers; use unrouted `Tabs`
for detail structure; nest `PageLayout` under a parent layout `Outlet`; put create actions only in table filters.

## Routing

React Router v7. Redirects: `loader` + `redirect` — never `<Navigate>`. See `client/views/install/routes.tsx`.

## Page titles

Every routed view sets its own title; layouts never set `document.title`. `PageTitleProvider` appends `| Nuon`. At most
two segments, most specific first. Org-level pages have no owner segment. Pass `install?.name` directly (never
`${x?.name}`). Put `<PageTitle>` first on every early-return branch (use a fragment).

## API integration

`api<T>()` returns **`T` directly** — not `{ data: T }`. Import from `@/lib`. On 401, `api.ts` redirects to login.

### Defensive data access

Treat API data as potentially undefined: optional chaining, guard before render, nullish coalescing, `enabled` for
non-null assertions in `queryFn`, provider hooks may be undefined (`org?.id`).

## State and TanStack Query

Providers via hooks (`useOrg()`, `useInstall()`, …) — never raw `useContext`. Always invalidate related queries on
mutation success. Lists: `placeholderData: keepPreviousData` when revisiting.

## SSE / real-time (client)

SSE writes into the TanStack Query cache via `setQueryData`. Use shared hooks — never hand-wire `EventSource` in a
provider: `useSSEResourceQuery`, `useSSETimelineQuery`, `createSSEQueryListener`. Use `isTerminalStatusV2` for
terminals. New SSE views need **both** a Go `runSSEStream` handler and a client hook with matching event names.

## Auth and config

- Cookie `X-Nuon-Auth`; BFF validates and proxies
- `useConfig()` reads `window.__NUON_CONFIG__` — never hardcode API URLs
- Feature flags on the org via `useOrg()`: `org?.features?.['flag-name']`
- New flags: `services/ctl-api/internal/app/org.go` — append to bottom of `GetFeatures()`

## TypeScript

- `T` prefix for data/API types; `I` for component props
- Import types from `ctl-api.types.ts` only

## Comments

Comment only when the *why* is non-obvious. Do not narrate what the code does.

## Forms (TanStack Form + Zod)

Use `Form*` wrappers in `client/components/common/form/`. Canonical: `client/components/api-tokens/CreateApiToken/`.

Always TanStack Form + Zod unless E1 (zero controls) or E2 (type-to-confirm not sent in body). Allowlist:
`DeploymentPlanEditor` only — ask before adding another exception.

- Flat fields only (no nested objects — breaks `canSubmit` in Form v1)
- No native `required` — Zod only
- Errors → `FormErrorBanner`; success → close modal + toast
- Every form gets `.stories.tsx`

## Component patterns

Container/presentational split under `client/components/[domain]/MyComponent/`; barrel exports the Container as the
public name. Never shadow a directory with a flat sibling file.

Before building UI: check `common/` and domain dirs; read `.stories.tsx` first.

- Loading: primitive `loading` props / `<Table isLoading>` / `<Loading>` — no hand-built skeletons
- Icons: only `Icon` from `@/components/common/Icon`
- Links: `Link` from `@/components/common/Link` (`href`, not `to`)
- Disabled button reasons: `tooltipProps` on `Button`
- Admin tools: `AdminDashboardLink` / `TemporalLink` from `client/components/admin/`
- Modals/panels: `Modal` / `Panel` from `surfaces/` — never `*Base`
- Ladle v5: plain function exports; stories use presentational component + mocks
- Tab object keys: all-lowercase

## Design, copy, toasts, dates

- [DESIGN.md](./DESIGN.md), [COPY_STYLE.md](./COPY_STYLE.md)
- Toasts: heading + description; status transitions via `useStatusToast`
- Dates: Luxon via `<Time>` / `<Duration>` only

## Scripts

```bash
bun run dev
bun run lint
bunx tsc --noEmit --project client/tsconfig.json
bun run dev:ladle
bun run test
bun run test:e2e         # Playwright smokes (needs stack + E2E_EMAIL)
bun run test:e2e:ui
bun run test:e2e:headed
```

Do not run production builds (`build`, `build:js`, `build:css`) unless explicitly asked.

## Playwright (E2E and agent verification)

Committed smoke tests live under `e2e/` (`playwright.config.ts`, `global-setup.ts`, `flows/` → `specs/`). See
`e2e/flows/README.md` and `e2e/global-setup.ts` for setup details. Typical run:

```bash
E2E_EMAIL=you@example.com bun run test:e2e
# optional: E2E_ORG_ID=orgXXX  E2E_BASE_URL  E2E_ADMIN_API_URL  E2E_PUBLIC_API_URL
```

Defaults assume dashboard `:4000`, public API `:8081`, admin API `:8082` when a local control plane + dashboard are
available. Install browsers once: `bunx playwright install chromium`.

### Scratch checks and screenshots (agent workflow)

**Throwaway verification belongs in gitignored `tmp/` as a standalone `.mjs`, run with `bun run tmp/<name>.mjs` from
`services/dashboard-ui`.** Do not write scratch checks as `e2e/specs/*.spec.ts` (or anywhere else committed). `tmp/` is
gitignored so they cannot be committed by accident. Only graduate a check to `e2e/specs/` + `e2e/flows/` when it should
be a permanent smoke — and say so explicitly.

Do not use the Playwright test runner for scratch work (`playwright test` / `.spec.ts` outside `testDir`): a plain `.mjs`
needs no custom config.

Auth (same idea as `e2e/global-setup.ts`):

1. `POST http://127.0.0.1:8082/v1/general/seed-user` with `X-Nuon-Admin-Email` (`seed@nuon.co`, or `NUON_DEV_EMAIL` for
   your own orgs) → `{ api_token }`
2. Inject as cookie `X-Nuon-Auth` on `127.0.0.1`
3. Optional: call public API `:8081` with `Authorization: Bearer <token>` + `X-Nuon-Org-ID` to discover org/app IDs

Gotchas:

- Never `waitUntil: 'networkidle'` — the SPA polls (SSE + `refetchInterval`); use `domcontentloaded` then `waitFor` a
  page-specific selector
- Dark mode is `prefers-color-scheme` — set `colorScheme: 'dark'` on the Playwright context (no `.dark` class)
- From `tmp/`, `import { chromium } from 'playwright'` (Bun walks up to `node_modules`)
- Screenshots: `deviceScaleFactor: 2`+, write under `tmp/`

Minimal screenshot sketch:

```js
import { chromium } from 'playwright'

const ADMIN = 'http://127.0.0.1:8082'
const APP = 'http://127.0.0.1:4000'
const EMAIL = process.env.NUON_DEV_EMAIL ?? 'seed@nuon.co'

const seed = await fetch(`${ADMIN}/v1/general/seed-user`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json', 'X-Nuon-Admin-Email': EMAIL },
  body: '{}',
})
const seedBody = await seed.text()
const { api_token } = JSON.parse(seedBody.match(/^\{[^}]*\}/)[0])

const browser = await chromium.launch()
const ctx = await browser.newContext({ colorScheme: 'dark', deviceScaleFactor: 2 })
await ctx.addCookies([
  { name: 'X-Nuon-Auth', value: api_token, domain: '127.0.0.1', path: '/', httpOnly: true, sameSite: 'Lax' },
])
const page = await ctx.newPage()
await page.goto(`${APP}/<orgId>/...`, { waitUntil: 'domcontentloaded' })
await page.getByRole('heading').first().waitFor({ timeout: 15000 })
await page.screenshot({ path: 'tmp/shot.png', fullPage: true })
await browser.close()
```
