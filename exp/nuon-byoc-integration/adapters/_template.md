# Adapter: <language> + <framework>

Copy this file to `adapters/<lang>-<framework>.md` to teach the skill a new
stack. The `references/` contract is reused unchanged — only fill in idioms.

## Config & secrets

How this stack reads config / secrets. Which mechanism holds `NUON_API_TOKEN`,
`NUON_ORG_ID`, `NUON_APP_ID`, `NUON_API_URL`, `NUON_CLOUD_PLATFORM`.

## Nuon client wrapper

Idiomatic HTTP client for this language that injects `Authorization: Bearer` +
`X-Nuon-Org-ID`, sets JSON content type, and normalizes errors with the upstream
status. (See `references/ctl-api.md`.)

## Routes / handlers

How routes are declared in this framework. Implement:
- `GET  /api/install-inputs`  (proxy input schema, customer-facing only:
  `source == "customer"` and not `internal`)
- `POST /api/installs`        (authorize → whitelist → inject → forward)
- `GET  /api/installs/:id`    (tenant-scoped status read)

Handler logic is identical across stacks — see
`references/contracts/install-create.md` pseudocode.

## Validation

Idiomatic validation library. Build the input validator dynamically from the app
input schema.

## Authorization

Where `authorizeCreate` / `authorizeReadInstall` hook into this framework's
session/tenant model. Mark clearly as required TODOs.

## Frontend (if applicable)

If this stack owns the UI too, describe the framework's data-fetching + form
idioms mirroring `adapters/react.md`.

## Gotchas

Anything stack-specific (async model, secret injection, CORS, etc.).
