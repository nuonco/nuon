# Architecture & security (language-neutral, non-negotiable)

This is the contract every generated integration must satisfy, regardless of
language or framework.

## Three tiers

```
Customer's browser  ─▶  Vendor server (proxy)  ─▶  Nuon ctl-api
   (untrusted)            (holds secrets)            (control plane)
```

- **Vendor UI** — operated by the vendor, used by their customers. Talks ONLY to
  the vendor server. Knows nothing about Nuon credentials or the ctl-api URL.
- **Vendor server** — operated by the vendor, hosted in their cloud account.
  Holds the Nuon API token + org ID. Exposes a narrow `/api/...` surface and
  proxies to ctl-api.
- **Nuon ctl-api** — operated by Nuon, hosted in the vendor's cloud account
  (usually the same account as the vendor server).

## Security invariants (the proxy MUST enforce all of these)

### 1. Credentials never reach the browser
The Nuon token + org ID are read from the server's secret store / env only.
They are never sent to the client, never embedded in the bundle, never returned
in a response body. The ctl-api base URL also stays server-side.

### 2. The customer is authorized before every forward
The proxy maps the authenticated customer (the vendor's own session/tenant model)
to what they may do:
- Which `app_id` / `org` they may target.
- Which installs they may read or mutate (tenant ownership).

Generate this as a **required, explicit** step with a clear TODO where the human
wires their tenant model. A proxy that forwards without an authorization check is
a defect — do not generate one.

### 3. Server owns sensitive fields; customer input is whitelisted
The customer request contains only:
- `name`
- `inputs` — restricted to the keys declared in the app's input schema, and with
  `internal` / non-`Sensitive`-appropriate handling per the schema.

The server injects everything else: `app_id`, the cloud account block, tenant
`labels`, `metadata`. The customer must not be able to set `app_id`, `labels`,
`metadata`, `install_config`, or unknown input keys. Reject unknown keys rather
than silently dropping them.

### 4. Errors are mapped, not forwarded
Upstream ctl-api errors are logged server-side and returned to the customer as
sanitized, mapped messages (see `ctl-api.md`). Never leak raw upstream bodies.

### 5. The write path is async
Creating an install returns immediately with an install in a provisioning state.
The UI must not present "created" as "ready" — it polls status (see
`lifecycle.md`).

## Minimal proxy surface for "create an install"

| Vendor endpoint            | Forwards to                                | Notes                              |
| -------------------------- | ------------------------------------------ | ---------------------------------- |
| `GET  /api/install-inputs` | `GET /v1/apps/{app_id}/input-latest-config`| drives dynamic form; token hidden  |
| `POST /api/installs`       | `POST /v1/apps/{app_id}/installs`          | authorize + whitelist + inject     |
| `GET  /api/installs/:id`   | `GET /v1/installs/{install_id}`            | status polling; tenant-scoped read |

`app_id` is a server config value, not a path/param the client controls.
