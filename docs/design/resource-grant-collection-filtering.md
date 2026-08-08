# Resource grants: collection filtering for top-level list endpoints

**Status**: deferred — designed 2026-08-08, not yet scheduled. Tiers 1–2 are
ready to implement; Tier 3 needs a denormalization decision.

## Context

The resource-grant middleware classifies every org-scoped route (enforced at
startup by `middlewares/org/coverage.go`) into:

- **Resolvable** — the route names a resource (param, owner-resolver registry,
  or type-only create); the walk-up primitive authorizes it per request.
- **Org-level by design** (`orgLevelRoutes`) — org settings, team/credential
  surfaces, and all org-wide collection/aggregate endpoints. These enumerate
  resources across the whole org, so an org-level permission is required.
- **Uncovered** (`uncoveredRoutes`) — accepted gaps (body-carried IDs, the
  trigger family pending a grantable org-owned type).

The exception to "collections are org-level" is `filteredCollections`
(`middlewares/org/collections.go`): four core navigation lists that authorize
deferred grantees by **filtering results to entitled rows** in the handler
instead of gating the endpoint:

- `GET /v1/installs`
- `GET /v1/apps`
- `GET /v1/apps/:app_id/installs`
- `GET /v1/components`

These are what make a grants-only account (e.g. `org_read_only`-less install
maintainer, or a future Install Manager grantee) able to see their resources
at all.

This document designs the extension of that pattern to every top-level
collection, so e.g. `GET /v1/workflows` returns a single-install grantee only
their install's workflows.

## The mechanism (per endpoint)

The existing pattern (`installs`, `apps`, `components` services) has four
parts; 2–4 are mechanical:

1. **A scope builder** on `scope.IDSets`
   (`internal/pkg/authz/scope/scope.go`) — a GORM scope mapping the account's
   granted ID sets (install IDs, app IDs, org IDs, per-org type wildcards)
   onto the resource table's columns. See `IDSets.Installs` for the canonical
   shape, including wildcard-org cascades and upward visibility.
2. **A per-service helper** (see `installs/service/grant_scope.go`) — returns
   a no-op scope when `cctx.OrgAuthorized` (org-wide fast path), the real
   filter for deferred grantees, and `WHERE 1 = 0` if no account is resolvable
   (fail closed).
3. **`.Scopes(...)` on the handler's list query** — composes with pagination.
4. **Route classification** — move the route from `orgLevelRoutes` into
   `filteredCollections`, and add a test asserting a single-install grantee
   sees only their rows.

## Endpoint tiers (by schema shape)

### Tier 1 — direct owner columns (~1 day, no schema changes)

| Endpoint | Approach |
|---|---|
| `GET /v1/policy-reports` | `policy_reports` has `app_id` + `install_id` columns; clone of the installs scope |
| `GET /v1/installs/health` | aggregates over `installs`; reuse `IDSets.Installs` verbatim |
| `GET /v1/installs/label-keys` | same |

### Tier 2 — one join away (~1 day, no schema changes)

| Endpoint | Approach |
|---|---|
| `GET /v1/builds` | `component_builds` reaches its app via `component_config_connections → components.app_id`; filter via subquery (same join the owner resolver uses) |
| `GET /v1/component-builds` | same |

### Tier 3 — polymorphic owners (the real decision, ~2–4 days)

`GET /v1/workflows`, `GET /v1/workflows/pending-approvals`,
`GET /v1/runner-jobs`, `GET /v1/queues`, `GET /v1/terraform-workspaces`.

These tables carry only `owner_id`/`owner_type`, and owners are sometimes
themselves children (runner job → deploy → install). The per-row owner
resolver (`middlewares/org/owners.go` `resolveOwnerRef`) walks that at
request time; a set filter cannot walk per row. Two options:

**(a) SQL unions per owner type.** One `OR` branch per owner type, each with
its own subquery down to the granted tier. Works without schema changes, but:
queries get slow and fragile, and every future owner type silently escapes
the filter — the same drift problem the coverage validator exists to prevent,
reintroduced in SQL.

**(b) Denormalize `install_id` / `app_id` columns onto the four tables.**
Backfill migration per table (batched, resolving each row's owner chain once),
populated at write time going forward. Every Tier 3 endpoint then collapses
into the trivial Tier 1 case, and the per-row owner resolvers get cheaper as
a side effect (column read instead of a multi-hop walk).

**Recommendation**: option (b), as its own PR (migrations + write-path changes
deserve isolated review), scheduled when there is a concrete grant-user need
for these lists. The detail routes already resolve per-row, so the lists are
a navigation nicety rather than a capability gap.

### Excluded

`GET /v1/triggers` and `GET /v1/triggers/dispatches` — triggers are org-owned
entities with no install/app tier; there is nothing to filter below org.
They stay org-level until triggers become a grantable org-owned type (like
webhooks), which is a separate design.

## Sequencing

1. Tiers 1–2 whenever convenient (~2 days incl. tests, no schema changes).
2. Tier 3 denormalization as its own PR, gated on demand.
3. Route entries migrate `orgLevelRoutes` → `filteredCollections` per endpoint
   as each handler adopts filtering; the startup coverage validator keeps the
   contract honest throughout.

## Related

- `docs/design/resource-grants.html` — the original hierarchical grants design
- `middlewares/org/coverage.go` — route classification contract + validator
- `middlewares/org/owners.go` — per-row owner resolvers (incl. the polymorphic
  owner walk that option (b) would partially obsolete)
- `internal/pkg/authz/scope/scope.go` — the list-filtering primitive
