# Auth Conventions

This document defines how authentication and authorization work in `services/ctl-api`
and its SDKs, and what to do when adding endpoints, roles, or credentials.
**Authorization is additive grants over object keys; enforcement is per-surface
middleware plus, where declared, per-route `require.Route`.**

## Quick Reference

| I am adding… | Do this |
|---|---|
| A public-API endpoint | Register in `RegisterPublicRoutes`; the `org` middleware enforces `CanPerform(orgID, verbFromMethod)`. Add `@Security APIKey` + `@Security OrgID` swagger annotations. |
| A runner-API endpoint | Register in `RegisterRunnerRoutes`; the surface checks identity + org membership only. If the route needs a permission check, declare it: `group.GET(..., require.Route(kind, verb, param), handler)`. |
| A resource-scoped role | Grant the key `permissions.Object(orgID, kind, id)`. Never hand-build key strings. |
| A service account | `account.EnsureServiceAccount` + a role binding; reap both in the owning entity's delete path (`DeleteServiceAccount` handles bindings, tokens, and stack roles). |
| A new permission verb | Extend `permissions.Permission` + `NewPermission`. Verbs beyond the HTTP-method mapping must be assigned via `require.Route`, never inferred. |
| A credential for an SDK | Resolve through `sdks/auth` (explicit token → `NUON_API_TOKEN` → ambient OIDC exchange). Do not re-implement precedence. |

---

## Surfaces and their middleware chains

Each API surface has its own gin engine and config-driven middleware list
(`internal/pkg/api/`, `Cfg.*Middlewares`). Auth differs by surface — never assume a
check from one surface exists on another.

- **Public API (8081)** — token auth resolves the account, then the `org` middleware
  (`internal/middlewares/org/org.go`) requires the `X-Nuon-Org-ID` header (or
  `org_id` query param), verifies the org exists, and enforces
  `acct.AllPermissions.CanPerform(orgID, permissions.FromRequest(ctx))`. Every public
  route gets an org-wide, verb-level check for free.
- **Runner API (8083)** — token auth, then `runner_org`
  (`internal/middlewares/org/runner.go`): the account must belong to exactly one
  org, which is put in context. **No permission check.** Routes that need one
  declare it with `require.Route`.
- **Admin / internal / auth surfaces** — separate chains; consult their configs
  before assuming anything.

Token resolution (`internal/middlewares/auth/auth.go`) accepts, in order: the
`Authorization: Bearer` header, the `token` query param, then the `X-Nuon-Auth`
cookie. The token maps to an account; the account's roles/policies are aggregated
into `AllPermissions` by `Account.AfterQuery`.

## The permission model

`Account —< AccountRole >— Role —< Policy`. Each policy holds an HSTORE map of
**object key → verb** (`internal/pkg/authz/permissions/`).

- **Verbs**: `all`, `create`, `read`, `update`, `delete`. `all` matches any verb.
  `permissions.FromRequest` infers the verb from the HTTP method (GET→read,
  POST→create, PUT/PATCH→update, DELETE→delete); a route whose real verb differs
  from its method must declare the verb explicitly via `require.Route`.
- **Object keys** are strings with at most two tiers, split on the **first** `:`:
  - org-wide: the bare org ID — `org_admin` is `{orgID: all}`.
  - resource-scoped: `permissions.Object(orgID, kind, id)` →
    `"<orgID>:<kind>/<id>"`, kinds in `permissions.ResourceKind`
    (`app`, `install`, `stack`, …).
- **`Set.CanPerform(obj, verb)`** checks the exact key, then the parent (the part
  before the first `:`), then `*`. Consequence: **org-wide grants always cover
  scoped objects** — scoping constrains new narrow roles without breaking existing
  broad ones. Do not add deeper nesting; the fallback only resolves one parent tier.
- Grants are **additive only**. There is no deny rule; the absence of a grant is the
  denial. Merging across roles keeps the stronger verb.

## Declarative route authorization: `require.Route`

`internal/pkg/authz/require` is how a route states what it operates on:

```go
stacks := ge.Group("/v1/stacks/:install_id",
    require.Route(permissions.KindStack, permissions.PermissionRead, "install_id"))
```

Rules:

- The **declared verb is authoritative** — `FromRequest` plays no role once a route
  declares. This is the only way to express verbs the method mapping cannot.
- It assumes account and org are already in context, so today it is valid only on
  surfaces whose engine-wide chain provides them (the runner API). Rolling it out to
  the public API requires the org middleware to delegate for declared routes —
  planned, not yet built. Do not annotate public routes yet.
- **Denials are `stderr.ErrNotFound`, never 403.** A forbidden response confirms the
  resource exists in another org; not-found does not. Handlers behind the middleware
  keep their own org-scoped lookups as defense in depth.

## Roles

- **Managed org roles** (`standardOrgRoles`, `internal/pkg/authz/create_org_roles.go`)
  are the single source of truth for their permissions and metadata; the reconciler
  keeps existing orgs in sync. Never edit managed role rows directly.
- `Role.Contexts` controls which pickers offer a role (`team`, `service_account`,
  `api_token`, `oidc_trust_policy`). A role with no contexts exists and works but is
  never offered — use that for machine-only roles.
- **Per-resource roles** (e.g. the `stack` role type) are created idempotently by
  the code that owns the resource (`EnsureStackInstallRole`), bound with
  conflict-safe `AccountRole` inserts, and hard-deleted when the owning service
  account is deleted. When swapping a broad grant for a narrow one, **grant first,
  then revoke** (`RemoveAccountOrgRoleByType`): a crash in between leaves the
  account over-privileged and convergent, never locked out.
- `Role.CreatedByID`/`Policy.CreatedByID` are notnull and filled from context. In
  temporal activities there is no requester; set the acting account explicitly
  (a service account may be its own creator).

## Credentials and SDKs

- **Tokens are shown exactly once**, at mint time
  (`POST /v1/service-accounts/{account_id}/tokens`). No endpoint may echo a live
  token afterward.
- SDK credential precedence lives in `sdks/auth` and only there: explicit token →
  `NUON_API_TOKEN` → ambient OIDC token exchanged for a short-lived one. SDKs supply
  only the transport-specific exchange (`Exchanger`).
- The OIDC exchange audience is the URL of the API surface the SDK calls (the
  runner API for `sdks/stack`) — it must match what trust policies name.

## When touching auth, always

1. State which surface(s) a route lives on and what its chain already enforces.
2. Prefer not-found over forbidden anywhere a denial would confirm existence.
3. Build object keys through `permissions.Object`; never concatenate by hand.
4. Cover new checks with the table-test shape in
   `internal/pkg/authz/require/require_test.go`: right resource passes, wrong
   resource 404s, org-wide grant passes, no grant 404s, cross-org 404s.
5. Keep authorization changes surface-local: adding a check to one route group must
   not alter behavior on any other route or surface.
