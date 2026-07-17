# Service-account token for the server (language-neutral)

The vendor server authenticates to ctl-api with a long-lived bearer token
(`NUON_API_TOKEN`). Getting this credential right is a security decision, not a
config detail.

## Use a dedicated service account — not a person's token

- Nuon accounts have types; automation should use an `AccountType=Service`
  account, distinct from a human (`Auth0`) account.
- A service-account token is not tied to an individual, survives people leaving,
  and can be scoped and rotated independently.
- **Never ship a developer's personal token** (e.g. the one in `~/.nuon` from
  `nuon auth login`). That token is fine ONLY for local, throwaway verification
  while building the integration — never for the committed or deployed server.

## How the token is issued — the admin API (two steps)

A vendor operating a BYOC control plane has access to its **admin API** and mints
a durable token in two calls: create a service account, then create a static
token for it. **The skill does not run these for the user** — it displays the
directions and ready-to-paste curl commands, and the user runs them (they hold
the admin credentials and the token must not pass through the skill).

Both endpoints live on the **admin API**, not the public API:

- `admin_api_url` is the control plane's admin API base — separate from the
  public API and network-restricted. Per-deployment (BYOC); confirm with the
  user, do not assume.
- In-request auth is the `X-Nuon-Admin-Email` header naming an existing admin
  account; the real gate is network access to the admin API.

### Step 1 — create (or fetch) the service account

`POST {admin_api_url}/v1/orgs/{org_id}/admin-service-account` (empty body). It is
idempotent — returns the existing account if already created. The account's email
is `{org_id}-admin-service-account@serviceaccount.nuon.co`, and it is granted the
**org-admin** role for that org (this endpoint does not offer finer scoping; note
the blast radius to the user).

```bash
curl -sS -X POST "$ADMIN_API_URL/v1/orgs/$ORG_ID/admin-service-account" \
  -H "X-Nuon-Admin-Email: $ADMIN_EMAIL" \
  -H "Content-Type: application/json" \
  -d '{}'
# → { "id": "...", "email": "<org_id>-admin-service-account@serviceaccount.nuon.co",
#     "subject": "<org_id>-admin-service-account", ... }
```

### Step 2 — mint a static token for that account

`POST {admin_api_url}/v1/general/admin-static-token` with the service account's
email (or subject) from step 1. `duration` is optional (defaults to 1 year).

```bash
curl -sS -X POST "$ADMIN_API_URL/v1/general/admin-static-token" \
  -H "X-Nuon-Admin-Email: $ADMIN_EMAIL" \
  -H "Content-Type: application/json" \
  -d '{"email_or_subject":"'"$ORG_ID"'-admin-service-account@serviceaccount.nuon.co","duration":"8760h"}'
# → { "api_token": "<token>" }
```

The returned `api_token` is a long-lived `TokenTypeStatic` token. The user puts it
into the server's secret store as `NUON_API_TOKEN` (see Storage below). If the
vendor prefers, the dashboard equivalent (Org settings → API tokens) also works.

## Storage

- Put the token in the project's existing secret mechanism (env var, secrets
  manager, platform secret) as `NUON_API_TOKEN`. Keep `.env` gitignored.
- The token is read server-side only and never returned to the browser
  (see `architecture-and-security.md`).
- Plan for rotation: the token can expire or be revoked; the server should fail
  clearly (surface a mapped 502) rather than silently.
