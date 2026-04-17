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
| `npm run test:e2e` | Playwright E2E tests |
| `npm run test:e2e:ui` | Playwright interactive UI mode |
| `npm run test:e2e:headed` | Playwright with visible browser |

## E2E tests

Playwright smoke tests that run against a live local (or staging) environment. Chromium only.

### Prerequisites

- Local dev stack running (dashboard-ui + ctl-api + postgres + temporal)
- An admin account email with access to the admin API

### Running

```bash
E2E_EMAIL=you@nuon.co E2E_ORG_ID=orgXXX npm run test:e2e
```

### Environment variables

| Variable | Default | Required | Purpose |
|----------|---------|----------|---------|
| `E2E_BASE_URL` | `http://127.0.0.1:4000` | no | Dashboard URL |
| `E2E_ADMIN_API_URL` | `http://127.0.0.1:8082` | no | Admin API for token generation |
| `E2E_EMAIL` | — | yes | Admin email (used to auth and generate token) |
| `E2E_ORG_ID` | — | yes | Org ID for test navigation |

### How auth works

The global setup calls `POST /v1/general/admin-static-token` on the admin API to generate a short-lived (1h) token, injects it as the `X-Nuon-Auth` cookie, and saves the browser state. All specs reuse this saved auth state.

### Flow docs

`e2e/flows/` contains markdown flow specs that describe test scenarios in a structured format. These serve as source-of-truth documentation — update the flow markdown and regenerate specs from it.
