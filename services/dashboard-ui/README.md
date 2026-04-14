# Dashboard UI

Go BFF (Backend-for-Frontend) + React SPA for the Nuon platform.

## Architecture

```
server/     Go BFF (Gin + Uber fx) — serves the SPA, proxies /v1/* to ctl-api, injects runtime config
client/     React SPA (React Router v7, TanStack Query, Tailwind CSS v4, ESBuild)
public/     Static assets (served by the Go server)
```

The Go server validates auth cookies, rewrites them into `Authorization` headers for the ctl-api proxy, and injects `window.__NUON_CONFIG__` into the HTML at serve time.

## Development

The Go server is started separately via `nctl`. Then run the SPA dev server:

```bash
npm run dev
```

This watches `client/` for changes, rebuilds JS/CSS into `dist/`, and proxies through BrowserSync.

## Build

```bash
npm run build
```

Produces minified JS and CSS in `dist/`.

## Other scripts

| Script | Description |
|--------|-------------|
| `npm run dev:ladle` | Component stories (Ladle v5) |
| `npm run lint` | ESLint on `client/` |
| `npm run tsc` | TypeScript type check |
| `npm run fmt` | Prettier on `client/` |
| `npm run test` | Vitest tests |
