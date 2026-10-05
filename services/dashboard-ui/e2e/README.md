# Dashboard E2E

Playwright specs for the dashboard. `specs/` runs against a live dashboard and control plane
(`playwright.config.ts`); `specs-ladle/` runs against Ladle stories (`ladle.config.ts`). Flow docs that describe each
spec live in [flows/](./flows/README.md).

```bash
E2E_EMAIL=you@example.com bun run test:e2e                     # everything in specs/
E2E_EMAIL=you@example.com bunx playwright test -c e2e/playwright.config.ts specs/onboarding
bun run test:ladle                                             # specs-ladle/, needs Ladle running
```

`global-setup.ts` seeds a user, creates a sandbox org (unless `E2E_ORG_ID` is set), seeds an app, and writes the
session to `e2e/.auth/`. `global-teardown.ts` deletes the org it created.

## Environment variables

Every variable in [env.ts](./env.ts):

| Variable | Default | Used for |
|----------|---------|----------|
| `E2E_EMAIL` | none (required) | Admin email sent as `X-Nuon-Admin-Email` when seeding users and orgs. |
| `E2E_BASE_URL` | `http://127.0.0.1:4000` | The dashboard. |
| `E2E_ADMIN_API_URL` | `http://127.0.0.1:8082` | ctl-api admin API (seed users, static tokens, org features). |
| `E2E_PUBLIC_API_URL` | `http://127.0.0.1:8081` | ctl-api public API. |
| `E2E_ORG_ID` | unset | Reuse an existing org instead of creating one. |
| `E2E_APP_CONFIG` | `httpbin` | Which `nuonco/example-app-configs` directory global setup seeds. |
| `E2E_GITHUB_INSTALL_ID` | unset | GitHub App installation the onboarding own-app specs connect. Locally, `POST :8082/v1/general/seed-user` returns it as `github_install_id`. |
| `E2E_ONBOARDING_REPO_AWS` | unset | `owner/repo` with an AWS app config at its root, visible to that installation. Locally `nuonco/kitchen-sink` works. |
| `E2E_ONBOARDING_REPO_GCP` | unset | Same, for GCP. |
| `E2E_ONBOARDING_REPO_AZURE` | unset | Same, for Azure. |

An own-app onboarding spec skips, with the missing variable in its skip reason, when `E2E_GITHUB_INSTALL_ID` or its
cloud's repo variable is unset. The other onboarding specs need only `E2E_EMAIL`.

## Onboarding specs

`specs/onboarding/` covers the first-run flow ([flow](./flows/onboarding.flow.md)). The dashboard has to be started
with `NUON_ONBOARDING_FIRST_RUN=true`, or `/onboarding` serves the existing wizard. Each spec makes a brand-new
account (`e2e/onboarding.ts`), so it starts with no org, the way a new sign-up does, and does not use the shared
session from global setup. They expect ctl-api's `force_sandbox_mode` (on in the local stack), so the install stack
reports back without a real cloud account.

```bash
E2E_EMAIL=you@example.com \
E2E_GITHUB_INSTALL_ID=<id> \
E2E_ONBOARDING_REPO_AWS=nuonco/kitchen-sink \
bunx playwright test -c e2e/playwright.config.ts specs/onboarding --workers=1
```

`--workers=1` keeps the local control plane from queueing several app syncs at once; each spec waits up to five
minutes for its app config and stack.
