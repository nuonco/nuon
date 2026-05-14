# Dashboard UI

Go BFF (Backend-for-Frontend) + React SPA for the Nuon platform.

## Architecture

```
server/     Go BFF (Gin + Uber fx) — serves the SPA, proxies /v1/* to ctl-api, injects runtime config
client/     React SPA (React Router v7, TanStack Query, Tailwind CSS v4, Bun bundler)
public/     Static assets (served by the Go server)
```

The Go server validates auth cookies, rewrites them into `Authorization` headers for the ctl-api proxy, and injects `window.__NUON_CONFIG__` into the HTML at serve time.

## Development

The Go server is started separately via `nctl`. Then run the SPA dev server:

```bash
bun run dev
```

This watches `client/` for changes, rebuilds JS (bun build) and CSS (PostCSS/Tailwind) into `dist/`, and proxies through a Bun-based dev server with SSE live reload.

## Build

```bash
bun run build
```

Produces minified, content-hashed JS and CSS in `dist/assets/`.

## Other scripts

| Script                    | Description                     |
| ------------------------- | ------------------------------- |
| `bun run dev:ladle`       | Component stories (Ladle v5)    |
| `bun run lint`            | ESLint on `client/`             |
| `bun run tsc`             | TypeScript type check           |
| `bun run fmt`             | Prettier on `client/`           |
| `bun run test`            | bun test (unit tests)           |
| `bun run test:e2e`        | Playwright E2E tests            |
| `bun run test:e2e:ui`     | Playwright interactive UI mode  |
| `bun run test:e2e:headed` | Playwright with visible browser |

## E2E tests

Playwright smoke tests that run against a live local (or staging) environment. Chromium only.

### Prerequisites

- Local dev stack running (dashboard-ui + ctl-api + postgres + temporal)
- An admin account email with access to the admin API
- Playwright browsers installed: `bunx playwright install chromium`

### Running

```bash
# Creates a fresh test org, runs tests, deletes org on teardown
E2E_EMAIL=you@nuon.co bun run test:e2e

# Use an existing org (skips create/teardown)
E2E_EMAIL=you@nuon.co E2E_ORG_ID=orgXXX bun run test:e2e
```

### Environment variables

| Variable             | Default                 | Required | Purpose                                                                      |
| -------------------- | ----------------------- | -------- | ---------------------------------------------------------------------------- |
| `E2E_BASE_URL`       | `http://127.0.0.1:4000` | no       | Dashboard URL                                                                |
| `E2E_ADMIN_API_URL`  | `http://127.0.0.1:8082` | no       | Admin API for token generation                                               |
| `E2E_PUBLIC_API_URL` | `http://127.0.0.1:8081` | no       | Public API for org creation                                                  |
| `E2E_EMAIL`          | —                       | yes      | Admin email (used to auth and generate token)                                |
| `E2E_ORG_ID`         | —                       | no       | Existing org ID (if omitted, a fresh org is created and deleted after tests) |

### How it works

1. Global setup generates a static token via the admin API
2. If no `E2E_ORG_ID` is set, creates a fresh org via the public API (the token user becomes org admin)
3. Injects the token as the `X-Nuon-Auth` cookie and saves browser state
4. Tests run against the org
5. Global teardown deletes the org if it was created by the setup

### Flow docs

`e2e/flows/` contains markdown flow specs that describe test scenarios in a structured format. These serve as source-of-truth documentation — update the flow markdown and regenerate specs from it.
