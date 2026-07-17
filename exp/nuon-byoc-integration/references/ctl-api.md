# ctl-api reference (protocol, language-neutral)

The Nuon control-plane API. Everything here is plain HTTP + JSON — applies to any
language.

## Base URL (BYOC — varies per deployment, NEVER assume)

Nuon is a BYOC platform: the control plane a customer integrates with is **not
always** `api.nuon.co`. It may be a Nuon-managed multi-tenant gateway
(`https://api.nuon.co`) or a dedicated/self-hosted BYOC control plane on the
vendor's own domain. **Always confirm the target API URL with the user before
generating anything.**

How to discover the current URL:

- `nuon --help` prints `✅ You are logged into <api_url>.` — this is the URL the
  CLI (and therefore your exploration commands) is currently pointed at.
- It is also stored in `~/.nuon` as `api_url`.

Confirm with the user that this is the control plane they want the integration
to target. If not, they must `nuon auth login` against the correct URL (or you
must use the URL they specify) before proceeding.

- The chosen URL is wired into the generated server as a `NUON_API_URL` (or
  equivalent) env var. Never hardcode it.

## Authentication

Every request carries two headers:

| Header           | Value                       | Source                    |
| ---------------- | --------------------------- | ------------------------- |
| `Authorization`  | `Bearer <NUON_API_TOKEN>`   | vendor server secret      |
| `X-Nuon-Org-ID`  | `<NUON_ORG_ID>`             | vendor server config      |

Also set `Content-Type: application/json` on requests with a body.

These credentials scope to a Nuon org and grant broad control-plane access. They
MUST stay server-side (see `architecture-and-security.md`).

## Error envelope

Non-2xx responses return a JSON body. Normalize it in the client wrapper to a
consistent shape for the rest of your code. Map upstream status codes:

| ctl-api status | Meaning                         | Suggested proxy response |
| -------------- | ------------------------------- | ------------------------ |
| 400            | invalid request body            | 422 to client            |
| 401 / 403      | bad/again token, org mismatch   | 500 / 502 (never leak)   |
| 404            | app/install not found           | 404                      |
| 409            | conflict (e.g. duplicate name)  | 409                      |
| 5xx            | upstream failure                | 502                      |

Never forward the raw ctl-api error body to the customer — it can reveal org
internals. Log it server-side, return a mapped, sanitized message.

## Endpoints used by this integration

| Purpose                  | Method + path                                   |
| ------------------------ | ----------------------------------------------- |
| Mint service-acct token  | `POST {admin_api_url}/v1/general/admin-static-token` (admin API; see `service-account-token.md`) |
| App input schema (latest)| `GET  /v1/apps/{app_id}/input-latest-config`    |
| Create install           | `POST /v1/apps/{app_id}/installs`               |
| Get install (status)     | `GET  /v1/installs/{install_id}`                |
| List app installs        | `GET  /v1/apps/{app_id}/installs`               |
| Current install inputs   | `GET  /v1/installs/{install_id}/inputs/current` |
| Update install inputs    | `POST /v1/installs/{install_id}/inputs`         |

Payload/response shapes for the create + input-schema flows are in
`contracts/install-create.md` and `contracts/install-inputs-schema.md`.

## Verifying against a real org

The `nuon` CLI drives the same API and is the fastest way to confirm shapes:

```bash
nuon --output json apps list
nuon --output json apps get <app_id>          # includes input schema
nuon --read-only ...                           # guardrail: blocks all writes
```

Use `--read-only` (or `NUON_READ_ONLY=1`) while exploring so you can't mutate
state. Auth comes from `~/.nuon` (`nuon auth login`).
