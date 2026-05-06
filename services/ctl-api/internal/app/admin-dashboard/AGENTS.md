# Admin Dashboard (React + BFF)

Internal admin dashboard for Nuon operations. A React SPA backed by a JSON BFF served by ctl-api. Completely separate from the customer-facing `dashboard-ui` SPA.

- **Port**: 8087 (configured via `admin_dashboard_http_port` in `internal/config.go`)
- **URL**: `http://localhost:8087` when running locally
- **Auth**: `X-Nuon-Auth` cookie resolves to an account; falls back to a synthetic `admin-dashboard` account ID for hook attribution. The route group is gated by whatever middlewares the deploy mounts via `admin_dashboard_middlewares` (typically the `admin` middleware checking `X-Nuon-Admin-Email` from the reverse proxy).

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Server | Go (Gin) |
| Frontend | React 18 + react-router 7 |
| Data | TanStack Query |
| Styling | Tailwind CSS v4 (PostCSS) |
| Bundler | esbuild + browser-sync (live-reload) |
| Type-check | `tsc --noEmit` |

There is no templ, no HTMX, and no Vite. Old `views/*.templ` and `components/*.templ` were replaced by `client/views/**/*.tsx` and `client/components/**/*.tsx`.

## Directory Structure

```
admin-dashboard/
├── service/                    ← Go BFF (Gin handlers returning JSON)
│   ├── service.go              ← FX params, constructor, route registration
│   ├── spa.go                  ← static file serving + SPA fallback
│   ├── views/                  ← shared Go types used by handlers (NOT views)
│   │   ├── types.go
│   │   ├── workflow_info.go
│   │   ├── workflow_detail_types.go
│   │   └── temporal_workers_types.go
│   └── *.go                    ← one file per handler
├── client/                     ← React SPA
│   ├── index.html              ← shell (favicon, fonts, /styles.css, /app.js, /app.css)
│   ├── index.tsx               ← React mount
│   ├── App.tsx                 ← router + providers
│   ├── styles.css              ← Tailwind v4 entry
│   ├── tsconfig.json
│   ├── components/             ← shared UI (Badge, JsonViewer, StatusHistory, ...)
│   ├── views/                  ← page components, one folder per domain
│   ├── lib/
│   │   ├── api.ts              ← fetch wrapper (prefixes /api, sends credentials)
│   │   └── admin-api/          ← typed client per domain (orgs, queues, signals, ...)
│   ├── providers/config-provider.tsx
│   ├── types/admin.types.ts    ← shared response types
│   └── utils/format.ts
├── package.json                ← npm scripts (dev, build, tsc)
├── postcss.config.js
└── scripts/hash-assets.js      ← prod step: content-hashed asset filenames
```

## Backend handler pattern

Handlers return JSON. No `templ.Handler(...)`, no redirects.

```go
func (s *service) OrgDetail(c *gin.Context) {
    org, err := s.fetchOrg(c.Request.Context(), c.Param("id"))
    if err != nil {
        s.l.Error("failed to fetch org", zap.Error(err))
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch org"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"org": org, /* ... */})
}
```

Mutations return `{"status": "..."}` rather than redirecting — the React caller invalidates the relevant query key.

## Frontend pattern

Each page component uses `useQuery` for reads and `useMutation` for writes:

```tsx
const { data, isLoading } = useQuery({
    queryKey: ['org', id],
    queryFn: () => getOrgDetail(id!),
})

const restartMutation = useMutation({
    mutationFn: () => restartQueue(id!),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['queue', id] }),
})
```

Polling replaces HTMX `hx-trigger="every Ns"`:

```tsx
useQuery({ queryKey: [...], queryFn: ..., refetchInterval: 5000 })
```

The `lib/admin-api/` modules are the only place that knows BFF URL shape. Pages never call `fetch` directly.

## Route registration

All routes live under `/api/*` and are registered in `service.go` → `RegisterAdminDashboardRoutes()`. Anything not under `/api/` falls through to `spa.go`'s SPA fallback (serves `dist/index.html`).

The other four `Register*Routes` methods return `nil` — admin-dashboard only uses the dashboard route context.

## Local dev

```bash
cd services/ctl-api/internal/app/admin-dashboard
npm install
npm run dev   # esbuild --watch + postcss --watch + browser-sync
```

In a second terminal, run ctl-api so the BFF listens on `:8087`. Open `http://localhost:8088` (browser-sync proxy) for live-reload, or `http://localhost:8087` to hit the Go server directly.

`npm run dev` writes to `dist/`. `spa.go`'s `registerStaticSPA` serves that directory and falls back to `dist/index.html` for client-side routes.

## Production build

```bash
npm run build
```

Produces content-hashed `dist/assets/*` and rewrites `index.html` link/script tags via `scripts/hash-assets.js`. The Go server serves `/assets/*` with long cache headers and `index.html` with `no-cache`.

## Adding a new page

1. Add the BFF handler under `service/<feature>.go` returning JSON, register it in `service.go` under the `/api` group.
2. Add a typed client function in `client/lib/admin-api/<domain>.ts`.
3. Add the response type to `client/types/admin.types.ts` if it's shared.
4. Build the React view under `client/views/<domain>/<Page>.tsx` using `useQuery` / `useMutation`.
5. Wire the route in `client/App.tsx` and add a sidebar entry in `client/components/layout/AppLayout.tsx`.
