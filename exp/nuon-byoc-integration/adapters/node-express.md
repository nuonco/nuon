# Adapter: Node + Express (server)

Reference implementation of the proxy for a Node/Express server. Idioms only —
the contract lives in `references/`. Copyable snippets are in `assets/server/`.

## Config & secrets

Read from env (or the repo's existing config module — match it):

```
NUON_API_URL   (default https://ctl.prod.nuon.co)
NUON_API_TOKEN (secret — never sent to client)
NUON_ORG_ID
NUON_APP_ID
NUON_CLOUD_PLATFORM  (aws | azure | gcp)
```

If the repo uses `dotenv` / a config loader / a secrets manager, wire into that
instead of raw `process.env`.

## Nuon client wrapper

See `assets/server/nuonClient.js`. A `fetch` wrapper that:

- injects `Authorization: Bearer ${NUON_API_TOKEN}` and `X-Nuon-Org-ID`,
- sets JSON content type,
- throws a normalized error carrying the upstream status for the route to map.

## Routes

See `assets/server/installsRouter.js`. Mount under `/api`:

- `GET  /api/install-inputs` → `GET /v1/apps/{app_id}/input-latest-config`,
  returning only customer-facing inputs (`source == "customer"`, and not
  `internal`) — see `references/contracts/install-inputs-schema.md`.
- `POST /api/installs` → validate, `authorizeCreate` (TODO stub), whitelist
  inputs against the schema, inject cloud account + tenant label, forward.
- `GET  /api/installs/:id` → tenant-scoped read of `GET /v1/installs/{id}`.

## Validation

Use the repo's existing validator (`zod`, `joi`, `express-validator`). If none
exists and it's a TS project, `zod` is a good default. Build the input schema for
validation dynamically from the fetched app input schema, not a static object.

## Notes

- Keep `authorizeCreate` / `authorizeReadInstall` as clearly-marked TODO
  functions wired to the repo's session/tenant model — do not ship them as
  no-ops silently.
- If the project is Fastify/Nest/Next instead, keep this structure and translate
  route registration + middleware idioms; the handler logic is identical.
