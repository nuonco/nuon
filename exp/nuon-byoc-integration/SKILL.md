---
name: nuon-byoc-integration
description: >-
  Integrate a vendor application with the Nuon BYOC ctl-api so the vendor's own
  customers can create and manage Nuon installs from the vendor-owned UI. Adds a
  server-side proxy (any language/framework) that holds the Nuon API token and
  forwards requests to ctl-api, then wires up frontend components that call the
  new proxy endpoints. Use when the user wants their app/customers to create
  installs, drive installs, or otherwise talk to a Nuon BYOC control plane from
  their own product. Trigger phrases: "integrate my app with Nuon", "let
  customers create installs from my UI", "proxy the ctl-api", "add a create
  install button".
---

# Nuon BYOC Integration

Build an integration that lets a vendor's customers create and manage Nuon
installs from the vendor's own product, without ever exposing Nuon credentials
to the browser.

## The architecture (always)

```
┌────────────┐   /api/... (vendor)   ┌─────────────────┐   /v1/... (ctl-api)   ┌──────────┐
│ Vendor UI  │ ────────────────────▶ │  Vendor server  │ ────────────────────▶ │  Nuon    │
│ (customer) │                       │  (proxy)        │  Bearer + X-Nuon-Org  │  ctl-api │
└────────────┘                       └─────────────────┘                       └──────────┘
```

Three tiers, three properties that MUST hold:

1. The Nuon API token and org ID live **only** on the vendor server. Never in
   the browser bundle, never returned to the client.
2. The proxy **authorizes the customer** (vendor's own tenant model) before
   forwarding, and owns the sensitive fields (`app_id`, cloud account block).
3. The customer supplies only a **whitelisted** subset of the ctl-api payload
   (`name` + declared inputs).

Read `references/architecture-and-security.md` before generating any code — it
is the contract the whole integration is built against and is fully
language-neutral.

## This skill is language- and framework-agnostic

Do NOT assume Node, TypeScript, Express, or React. The `references/` directory
is pure protocol (HTTP, auth, payloads, lifecycle, security) and applies to any
stack. The `adapters/` directory holds idioms for specific stacks and is
additive — Node/Express and React ship as the reference implementations, but you
can and should generate a correct integration for an unknown stack from the
references + the repo's own conventions, marking anything uncertain with
explicit `TODO` comments.

## Procedure

Follow these steps in order. Track them with tasks if the integration is large.

### 1. Detect the stack (do not assume)

Inspect the repo to determine:

- **Language**: `package.json` (Node), `go.mod` (Go), `pyproject.toml` /
  `requirements.txt` (Python), `Gemfile` (Ruby), `pom.xml` / `build.gradle`
  (Java/Kotlin), `composer.json` (PHP), etc.
- **Server framework**: e.g. Express/Fastify/Nest/Next (Node); Gin/Echo/net-http
  (Go); FastAPI/Flask/Django (Python); Rails/Sinatra (Ruby).
- **Frontend framework**: React/Vue/Svelte/Angular, or none.
- **Conventions**: how routes are declared, how config/secrets are read, what
  validation library exists, what HTTP client the frontend already uses, TS vs
  JS.

If an `adapters/<lang>-<framework>.md` exists for the detected stack, load it.
Otherwise generate from `references/` + observed conventions.

### 2. Provision a service-account token for the server

The server authenticates to ctl-api with a long-lived token. It MUST be a
**dedicated service-account token**, not a developer's personal token. Minting it
requires admin credentials and the token must never pass through the skill, so
**the skill does not create the token — it displays directions the user runs
themselves.** See `references/service-account-token.md`.

1. **Confirm the control plane first (BYOC — not always `api.nuon.co`).** Run
   `nuon --help`; it prints `✅ You are logged into <api_url>.` (also in `~/.nuon`
   as `api_url`). Show that URL to the user and confirm it is the control plane
   they want to integrate with. If not, have them `nuon auth login` against the
   correct one, or use the URL they specify. This URL becomes `NUON_API_URL` —
   never assume or hardcode it.
2. **Display the service-account-token directions.** Show the user the two-step
   admin-API flow with ready-to-paste curl commands (fill in the placeholders you
   know — `ORG_ID`, and the admin API base if the user gave it):

   ```bash
   # Set these first:
   ADMIN_API_URL=...            # BYOC admin API base (ask the user; not the public API)
   ADMIN_EMAIL=...              # an admin account email you control
   ORG_ID=<org_id>

   # 1) Create (or fetch) the org service account — idempotent, empty body:
   curl -sS -X POST "$ADMIN_API_URL/v1/orgs/$ORG_ID/admin-service-account" \
     -H "X-Nuon-Admin-Email: $ADMIN_EMAIL" -H "Content-Type: application/json" -d '{}'
   #   → note the returned "email" / "subject"

   # 2) Mint a long-lived static token for that service account:
   curl -sS -X POST "$ADMIN_API_URL/v1/general/admin-static-token" \
     -H "X-Nuon-Admin-Email: $ADMIN_EMAIL" -H "Content-Type: application/json" \
     -d '{"email_or_subject":"'"$ORG_ID"'-admin-service-account@serviceaccount.nuon.co","duration":"8760h"}'
   #   → { "api_token": "<token>" }  ← this is NUON_API_TOKEN
   ```

   Tell the user this service account is granted **org-admin** (note the blast
   radius). Do not run these commands for them and do not ask them to paste the
   token back to you.
3. **Never bake the token into the repo.** The user places the `api_token` into
   the project's secret mechanism as `NUON_API_TOKEN` themselves; keep `.env`
   gitignored. A personal `~/.nuon` token is acceptable ONLY for local
   verification, never for the committed/deployed integration — flag this
   explicitly.

### 3. Confirm what else can't be inferred

Ask ONLY for what the repo doesn't reveal:

- The target Nuon `app_id` and `org_id`. `org_id` / `app_id` are also in
  `~/.nuon` / discoverable via `nuon apps list`.
- The app's **cloud platform** (aws / azure / gcp) — determines the account
  block in the create payload.
- Where the customer's tenant→app authorization should hook into their existing
  auth/session model.

Do not ask for anything already discoverable in the repo or via the app schema.

### 4. Fetch the install-input schema

Install form fields are rendered **dynamically from the app's input schema**, not
hardcoded. Retrieve the schema (see `references/contracts/install-inputs-schema.md`)
via the `nuon` CLI / MCP, or plan the proxy's schema endpoint to fetch it live.
Use it to drive both the generated form and the server-side input whitelist.

### 5. Generate the server proxy

Per `references/contracts/install-create.md` and the relevant adapter:

- A small Nuon HTTP client that injects `Authorization: Bearer <token>` and
  `X-Nuon-Org-ID: <org>` and normalizes ctl-api errors.
- `POST /api/installs` — validate body, authorize caller, whitelist inputs,
  inject server-owned `app_id` + cloud account, call ctl-api, map errors.
- `GET /api/install-inputs` — proxy the app input schema so the UI can render
  fields (keeps the token server-side).
- `GET /api/installs/:id` — proxy install status for the async lifecycle.
- A tenant-authorization step, clearly marked where the human must enforce their
  own tenant→app mapping. Never generate a pass-through-everything proxy.

### 6. Generate the frontend

Per the frontend adapter:

- A typed API layer calling the vendor proxy (not ctl-api directly).
- A create-install form whose fields are rendered dynamically from the input
  schema (types, required, defaults, sensitive→masked).
- Create action with pending/disabled state, then success → navigate to a status
  view that polls `GET /api/installs/:id` (creation is async).
- Error surfacing from the mapped proxy errors.

### 7. Verify

Build/typecheck the project. Optionally dry-run the proxy against Nuon using the
`nuon` CLI `--read-only` guardrail or a test org before wiring the write path.

### 8. Report

List generated files, the security TODOs the human must complete (tenant auth,
secret storage), and how to extend to more operations (deploy, teardown) by
adding contract files.

## Extending

- **New language/framework**: copy `adapters/_template.md`, fill in the idioms.
  References are reused unchanged.
- **New operation** (deploy, teardown, status detail): add a file under
  `references/contracts/` following `install-create.md`, then generate against it.
