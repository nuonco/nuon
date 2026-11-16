# Granular-access delivery ledger

Implements the coordinated plan in [spec 009](prompt.md), using [spec 008's native model](../008-customer-rbac-foundation/native-grants-model.md). Status describes local work, not production activation. Current implementation owner is this thread; other workstream ownership has not been assigned.

## Current checkpoint: operation guards; credential integration is local

The pure engine and design are checkpointed in local commit `26c68cd1f2`; projection-independent native loading, domain-owned install/job/log resolution, and scoped ID queries in `3b2161b8e8`; bounded request authority and native Gin guards in `4f27e5983f`. Persisted ceilings and native authentication are the next uncommitted increment. None of these checkpoints has been pushed. M0's complete route/writer/cohort inventory and remaining API contracts are not finished. No customer-facing capability is newly assignable, and no production route has switched authorization engines.

| Shared contract | Implemented behavior | Remaining integration |
| --- | --- | --- |
| Catalog | `DefaultCatalog` supplies the initial org/app/install/job/log definitions. `NewCatalog` validates references, cycles, supported scopes and matcher availability, and flattens nested bundles. `Definitions` returns copied authoring metadata; `Revision` identifies normalized definitions | Add verified domain operations through this catalog; readiness/actor-specific assignment eligibility belongs to the API rollout, not catalog existence |
| Policy document | `Catalog.DecodePolicy` strictly decodes format 4; `NormalizePolicy` validates and copies grants, normalizes explicit ID/permission duplicates, and retains authored bundle references. `Policy` is the serialization DTO | Existing DB policy hooks still decode v3 and demand legacy projection; ordinary format-4 writes remain disabled. The new loader reads grant records independently of those projection hooks |
| Compilation | `Catalog.Compile` binds the document to a trusted org. `Role.CompileGrants` validates role/policy/binding ownership without legacy projection. `LoadAccountGrants` checks live bindings; `LoadAuthority` combines them with a required ceiling and authenticated request identity. Authentication loads this authority for tokens with explicit ceilings and operation-based routes | No production route or writer is activated. Recompile authored documents on catalog revision changes |
| Point matching | `Catalog.Match` evaluates resolved facts. Domain `access` helpers supply live, tenant-checked install/job/log ownership through `Resources.Resolve` and loaded job/stream helpers | Request paths have not adopted the helpers; other target families and unsupported owner cases remain unconverted |
| Credential limit | Grants union within the account and within the ceiling, then intersect on operation/resource pairs. Nullable `Token.PermissionCeiling` persists authored format-4 references. SQL NULL keeps the existing credential path; an explicit empty policy denies all operations. Bounded-token authentication supports direct static service-account tokens on declared org-scoped ctl-api routes | Issuance/delegation, inventory/rotation, other credential protocols and reader-capable rollback rehearsal remain. No shared schema has been migrated |
| Collections | Domain resource sets expose `Resources.AuthorizedIDs`, using the same ownership queries as their point resolver. Grant alternatives union, then intersect with independent ceiling alternatives; no fetch-then-filter adapter | Route queries must apply the subquery before pagination/counts/aggregates. Production-sized query-plan/index verification and HTTP coverage remain |
| Product API | Native policy and catalog data types are available | Preview/ETag/owned-policy and token DTOs, transactional writes, delegation checks and API endpoints remain unimplemented |

The v3 catalog/checker remains separate because IDs such as `install.read` have different v3 meanings. Its document type and decoder are now named `LegacyPolicy` and `DecodeLegacyPolicy`. Existing `MatchPermissions`, `CanPerform`, writers, routes and database data have not switched authority. Shared resource-kind validation was extracted without changing accepted v3 scopes. The grant-preservation changes are included in the engine checkpoint.

The initial native catalog deliberately omits `org.write`: spec 008 requires its complete compatibility mapping before defining it. `install.write` includes deletion and cancellation; no Editor preset has been invented. Metadata is not a claim that the corresponding routes are ready for customer assignment.

## Coverage queue

This is a starting queue, not a completed route census. All rows still use existing request authorization. Each needs concrete route/helper/continuation entries and evidence before its native capability can activate.

| Family | Source entry points to inventory | Current native coverage | Missing proof / contract |
| --- | --- | --- | --- |
| Org/app reads | `apps/service/service.go`, org middleware, org services | Pure read bundle/point semantics only | Response classification, concrete ownership, all routes and list/count filtering |
| Installs and lifecycle | `installs/service/service.go`, `get_install.go`, mutation helpers/workflows | Native point resolution and scoped ID queries at org/app/install scopes | HTTP integration, full alternate-path coverage, embedded data contract, referenced-target and destructive-cascade checks |
| Jobs and logs | `runners/service/service.go`, `get_runner_jobs_ctl_api.go`, stream handlers and owner constructors | Explicit polymorphic ownership, same-operation/build job-log relations, scoped ID queries and PostgreSQL point/list parity controls | Remaining owner census, production query plans, and stream/download/SSE/HTTP integration |
| Builds, actions, workflows | App/component/action/workflow registrations and helpers | Not integrated | Operation inventory, admission, approval/retry/cancel checks and bounded continuation authority |
| Service accounts and access administration | `accounts/service/service.go`, `service_account.go`, `authz/set_account_role.go` | DTO/model primitives only | Owned-policy writes preserving unrelated bindings, delegation, preview, concurrency and native-aware old-client behavior |
| Credentials | `tokens/service/service.go`, token model, account token helpers/auth loaders | Persisted ceilings and real native authentication for direct static service-account tokens; existing expiry/soft-delete checks retained | Issuance validation, inventory/rotation/source-lineage integration, full historical-cohort and rollback verification |
| Remaining customer families | Section 3 of spec 009 and current registrations, not historical spec inventory alone | Not integrated | Telemetry, state/config disclosure, stacks/reporting, integrations/trust, triggers/delivery, notebooks/reports and explicit transport exclusions |

## Compatibility retirement queue

| Path | Why it remains | Removal gate |
| --- | --- | --- |
| V3 decoding and legacy projection in policy/account/role hooks | Existing persisted data and writers still use it | Current cohort conversion evidence, projection-independent native loading, converted writers/reconciliation and supported upgrade verification |
| Old object/CRUD request checks and broad org middleware | Domain routes have not adopted atomic operations/scoped predicates | Per-route/cohort authority recorded, complete entry-point mapping, baseline decision/response parity or explicit approved differences |
| Singular-role service-account replacement | Existing API/UI still depend on it | Native owned-policy API and clients work; old writers cannot erase native access or unrelated bindings |
| Existing credential admission | Existing tokens have no native ceiling; other credential protocols and production routes remain unconverted | Historical restrictions preserved, required ceiling failures deny, issuance/revocation/rotation checks and native-capable rollback verified |

Unconverted paths are unfinished work. Native grants must never fall through to broad legacy authorization. No compatibility path is removed merely because current role policies are believed to be v3.

## Verification recorded for the engine checkpoint

- Existing `go test ./services/ctl-api/internal/pkg/authz/permissions` passed before implementation.
- Existing `go test ./services/ctl-api/internal/pkg/authz/...` and the same command with `-race` passed after implementation.
- A disposable standalone Go probe executed 3,160 assertions through the new exported APIs: positive/negative operation-target matrices; all/selected installs; app-build/org/foreign exclusions; asymmetric read/write grants; independent shared-stream ceilings; empty versus absent ceilings; strict JSON; unchanged authored references; catalog cycles/unknown references/scope rejection; stable revisions; live expansion versus atomic-only grants; input-copy isolation; and v3 semantic separation.
- The probe is validation, not a retained test suite. No test files were added, following ctl-api guidance. Formatting uses `gofmt` and `goimports`.
- No PostgreSQL ownership/query, HTTP, UI, production inventory, migration or deployment validation has been claimed. The new point checker trusts resolved input; database-backed and request-level checks remain required.

## Native-loading increment

`LoadNativeAccountPermissions(ctx, accountID, orgID, catalog)` reads live account/org identity and bindings from PostgreSQL, then batch-loads bound roles and policies with GORM `Scan`. These are authority records, not legacy response projections: `Scan` deliberately avoids the custom `AfterQuery` projection hook while retaining model soft-delete predicates. The role compiler applies the same live binding and tenant validation as v3 before decoding and compiling format 4. No hook is globally disabled or changed.

Only bindings for the requested account/org contribute. Unbound roles and revoked bindings contribute nothing. A missing/deleted role or policy behind a live binding, duplicate binding/policy, invalid document, mixed document/HSTORE authority, or non-native policy in that account/org fails the entire native load; it never returns partial grants or falls back to v3. No bindings returns an explicit empty slice. Authored documents are not rewritten. Role names and version bookkeeping do not select authority.

This is an account-authority reader, not credential authorization. Existing authentication paths still use `LoadAccountPermissions`; existing policy-write hooks still reject format 4. Enabling native writers alone would therefore break existing readers. No migration, activation, token ceiling, query filter, or HTTP journey is included in this increment.

Validation completed:

- Existing authz and root app tests passed, including `-race`; `go vet` passed for those packages. Account loading, auth middleware, and MCP server tests also passed.
- A disposable PostgreSQL 17 container with minimal authority tables and the real `afterquery` plugin passed 57 executable assertions. Checks covered additive asymmetric read/write access, live binding revocation and narrowing on reload, unbound and foreign-org isolation, soft-deleted/missing identities and authority rows, malformed/mixed/v3 documents, duplicate records, explicit empty grants, query failure, preservation of authored bundle references, and v3 reader/admin/read-only baseline controls.
- The database probe bypassed normal writes only to seed its disposable fixtures. It confirmed that ordinary policy hooks and the legacy loader still reject native documents. It did not exercise the full migration schema, HTTP requests, token issuance, resource ownership resolution, or concurrent changes across the loader's queries.
- The disposable probe and container were removed after validation; no permanent tests or shared data were added or modified.

## Ownership and scoped-query increment

Domain `access` packages own SQL definitions of live resources and their supported ownership. SQL follows stored foreign keys, checks tenant equality and soft deletion at each ownership hop, and does not depend on response projections or model views. Missing, foreign, deleted, and unsupported owners produce no native target; SQL failures remain errors. This is explicit SQL, not a configurable ownership graph or path language. The original global `ResolveNativeTarget` / `NativeResourceIDs` implementation was replaced by the domain contract below before request integration.

| Owner family | Verified ancestry |
| --- | --- |
| Installs | Install → app → org |
| Deploys | Deploy → install component → install → app → org; no inferred `install_deploys.install_id` column |
| Action and sandbox runs | Run → install → app → org |
| Component builds | Build → component config connection → component → app → org; neither the executor runner nor referenced installs establish ownership |
| App sandbox builds and branch runs | Build → app, or branch run → branch → app → org |
| Workflows and steps | Workflow → supported org/app/install/branch/branch-config owner; step → workflow via `install_workflow_id` |
| Runner operations | Operation → runner → runner group → its org/install owner. A job's executor `RunnerID` is never an ownership edge |

Direct org/app/install/branch/branch-config owners are also recognized. Other owner types, including ClickHouse-only health-check records, remain unsupported by this PostgreSQL resolver. This table is implemented ownership coverage, not a complete production cohort census.

The job-to-stream read relation requires a live same-tenant job referencing the stream and an identical verified owner pair. Its initial allowlist is `install_deploys`, `install_action_workflow_runs`, `install_sandbox_runs`, `component_builds`, and `app_sandbox_builds`, based on their job/log constructors. Other owner pairs do not gain job-to-log access. Whole-stream semantics remain: different related jobs can independently satisfy account and ceiling scopes. Parent stream IDs add no authorization edge.

`Catalog.QueryScopes` selects copied scopes for one atomic operation, target kind, org, and API from current-revision compiled grants. `NativeResources.AuthorizedIDs` applies OR within each authority set and AND between account and ceiling sets. Wildcards retain ancestry checks; explicit arrays become bound SQL membership predicates. Nil ceiling is absent; empty ceiling denies. The query returns each resource once despite overlapping grants or shared streams, and can be used as `id IN (subquery)` inside an existing model/view query without fetching IDs into memory.

Validation completed for the initial global resolver (superseded by the refactor validation below):

- Existing authz and root app tests passed normally and with `-race`; authz `go vet`, account loading, auth middleware, and MCP server tests passed. Formatting and `git diff --check` were clean.
- A disposable PostgreSQL 17 probe passed 1,850 assertions with the actual afterquery, view, and pagination plugins. Its minimal fixture columns were checked against GORM's real model schemas. Independently specified expected ownership and allow/deny sets covered every implemented owner branch, asymmetric read/write grants, selected/all/future installs, app-build exclusion despite an install executor, shared-stream account/ceiling intersection, unrelated streams, parent-copy non-inheritance, missing/deleted/foreign owners and intermediate ancestors, explicit empty ceilings, API restrictions, stale catalogs, immutable compiled scopes, and bound ID parameters.
- Executed list, count, limit/offset, cursor, ID-subquery, and view/plugin composition checks confirmed filtering before pagination and no duplicate results. Point decisions and SQL results matched those independent expected sets.
- This was not full-schema migration validation, HTTP/credential integration, a production-data census, or production-scale query-plan validation. Those checks remain before activation. No runtime route or writer changed; no permanent tests were added.

Next integration work connects request authority selection and credential limits to these readers/resolvers/queries. The broader family inventory and service-account API work remain open. Persisted native writes remain disabled until readers and enforcement are ready.

## Domain-owned resource contract

The endpoint chooses the atomic operation. Domain helpers verify the target identity, tenant, ancestry, and supported relations. The permission engine evaluates those facts against account grants and credential ceilings; it does not know application tables. An install ID supplied by the request is not a substitute for resolving the requested job or stream.

- `installs/access.Resources(db, orgID)` returns the install resource set. `Resolve(ctx, id)` returns a `NativeTarget`; `AuthorizedIDs(ctx, catalog, query)` returns the SQL subquery for collections. The set binds its kind and tenant, rejecting a mismatched permission query.
- `runners/access.Jobs` and `LogStreams` expose the same contract. When a handler already loaded a job or stream, `JobTarget(ctx, db, job)` and `LogStreamTarget(ctx, db, stream)` select only that owner family's query and revalidate it in PostgreSQL. The supplied record is not accepted as proof of ancestry. Unrecognized or deleted owners remain denied.
- Org, app, component, and install `access` packages define their own ancestry queries. Install workflows reuse app/org/install queries; runner-group ownership reuses org/install queries. These lightweight packages avoid cycles between existing FX helper packages.
- Runner access owns the explicit supported owner list and the job-to-stream relation flag together. Point and collection paths use that same list. Full job/log collections still compose all supported owner queries; resolving a loaded deploy job does not query component builds.
- `authz.NewNativeResources` wraps a **trusted domain query**, not a user-supplied query or arbitrary scope. Its input must return one row per live resource with `id`, `org_id`, `app_id`, and `install_id`; absent parents are empty strings. Stream queries additionally return `related_job_ids` as a JSON array of verified same-tenant job IDs. It owns the SQL alias and matching predicates. Domain authors remain responsible for the ancestry joins; the wrapper cannot prove arbitrary SQL correct.
- Query builders start fresh GORM statements while retaining the connection/transaction and context. Resource sets clone statements between point and collection calls, so pagination, model aliases, or prior filters cannot contaminate another check. Apply the returned ID subquery to the route's model query **before** pagination, counts, and aggregates.

The brittleness review also consolidated operation/org/API/revision grant selection into one pure function used by `Catalog.Match` and `QueryScopes`. Returned query scopes remain copies, and empty versus absent ceilings retain their distinct meanings. The native loader and shared role/binding validation were reviewed without adding compatibility fallback or changing legacy behavior.

Refactor validation:

- Existing authz, root app, account service, auth middleware, and MCP server tests passed normally and with the race detector; all new domain packages compiled. `go vet` passed for authz, root app, and the domain access packages.
- A disposable PostgreSQL 17 probe passed 6,804 assertions against independently specified ownership and authorization expectations. It exercised every supported owner branch; selected/all installs; asymmetric read/write grants; exact and wildcard jobs; shared-stream intersections; API restrictions; absent/empty ceilings; stale catalogs; foreign, missing, deleted and unsupported ownership; and parent-log non-inheritance.
- Point and SQL collection checks agreed, including the loaded-record helpers. Checks also covered count/page and model-view/pagination-plugin composition, repeated use of a resource set, caller query isolation, retained transaction/context behavior, future resources, and rejection of mismatched tenant/kind query inputs. Renaming the component-build table in the disposable database did not break resolution of a deploy job or its stream.
- No HTTP path, credential loader, writer, migration, shared database, or deployment changed. Full-schema validation, production-sized query plans, and request-level activation checks remain open. These are disposable probes, not retained tests.

## Bounded native request increment

`LoadNativeAuthority(ctx, catalog, NativeAuthorityInput)` loads live native account grants and compiles an explicit format-4 credential ceiling into one request-scoped `NativeAuthority`. It binds account ID, credential ID, org, and listener API. Missing, malformed, or non-native ceilings fail loading; an explicit empty policy loads successfully but denies every operation. There is no constructor that silently substitutes unrestricted authority when a required ceiling is absent.

This is an **authorization loader, not an authenticator**. Its input must come from trusted authentication and persisted credential state, never request policy JSON. That caller must validate token identity, expiry, revocation, audience/source restrictions, and the credential's org binding before calling it. Passing IDs alone is not proof that a token exists or belongs to the account. The current Token model has no ceiling column, and no production caller supplies this input yet. Historical credentials need an explicit compatibility decision; they are not given invented ceilings here.

`NativeAuthority.Check(operation, target)` and `AuthorizedIDs(ctx, operation, resources)` use the same captured grants, ceiling, org, API, and catalog. Callers cannot replace those values between point and collection checks. Reload on each request; do not cache an authority across requests. This does not add a transaction snapshot or promise mid-request revocation.

The optional Gin adapters live beside the existing `require` helpers:

- `NativeRoute(router, api, operation, load, resolve)` takes ordinary loader/resolver functions and an explicit atomic operation, independent of the HTTP method. It validates the loaded authority against the request's account, token ID, org, and declared listener API, resolves domain facts, and runs the point check before the handler. Missing or denied resources return the same generic 404; loader/resolver failures return a generic 500 and log details server-side.
- `NativeCollectionRoute(router, api, operation, load, resources)` validates the same identity and produces the authorized ID query before the handler. The handler obtains it with `NativeCollectionIDs(ctx)` and applies it before pagination/counts/aggregates. An explicit no-access authority produces an empty result, not an unfiltered query. Looking up the query without its guard returns an error.
- Both adapters use the existing route-declaration marker. It is not an authorization grant or a native-cohort selector. The current auth/org middleware still runs unchanged; these helpers do not bypass its broad checks or turn native denial into legacy authorization.

No application route was registered with these guards. Before adoption, add persisted ceilings and native-aware authentication/cohort selection, preserve existing credential restrictions, convert writers, and complete each route's response and collection contracts. The next increment is credential storage/loading and the actual authenticated path, not enabling native policy writes in isolation.

Validation:

- Existing authz tests passed; authz, root app, auth/org middleware, and MCP server tests passed with the race detector. Authz `go vet` passed.
- A disposable PostgreSQL 17 probe ran the real native loader, domain install resolver, Gin guards, and stderr middleware with injected authenticated identity. Its 142 assertions covered selected/wildcard reads, asymmetric read/write grants and ceilings, identity/API mismatches, missing/malformed/empty ceilings, nil authority, missing/foreign targets, list/count results, query-copy isolation, future installs, narrowing/restoration/revocation on fresh requests, and loader/database failures. Broad legacy authority deliberately remained on the fixture account to verify there was no fallback. Error bodies did not expose private failure details.
- The probe did not exercise real token authentication, issuance, OAuth/source restrictions, full-schema migrations, or production routes. No permanent tests or shared-data changes were made.

## Persisted ceiling and authentication increment

This increment supersedes the loader callbacks and unchanged-authentication limitations described in the preceding checkpoint. `Token.PermissionCeiling` is nullable JSONB, hidden from ordinary JSON responses and retained in Temporal serialization. Normal model saves strictly decode and compile non-nil format-4 documents against the token org, normalize them, and retain authored bundle references. This validates shape and scope, not the caller's authority to issue a token. GORM map/column updates can bypass this hook; the future issuance API must validate explicitly, and authentication always validates persisted data again.

SQL NULL selects the existing authentication path with no invented ceiling or policy conversion. Every non-nil ceiling selects native authorization, including invalid JSON documents or JSON `null`; invalid authority cannot downgrade to legacy. Explicit empty grants are valid but authorize no operation. JSONB itself rejects malformed JSON; the codec rejects malformed native policy shapes.

The supported native credential cohort is deliberately narrow: a live, unexpired static token for a live service account, with both source fields empty, on an explicitly native, org-scoped ctl-api route. OAuth, federated, other token types, sourced tokens, global routes, and non-ctl native routes are refused. Their existing SQL-NULL tokens keep their existing behavior. This is not a managed-account conversion or an issuance-eligibility rule; managed-account refusal and delegation belong to the forthcoming write APIs.

Authentication resolves the token through the existing lookup/expiry checks, loads account identity without legacy role projection, and loads current grants intersected with the persisted ceiling. It stores that request authority in context. Operation-based route declarations publish their API metadata before inherited middleware, and `OperationRoute(router, api, operation, resolve)` / `CollectionRoute(router, api, operation, resources)` consume the authenticated authority rather than accepting a loader callback. Org middleware verifies account/token/org/API identity without requiring a broad org grant; API-surface middleware verifies the actual listener API. Later point/collection guards still enforce the declared operation. There is no legacy OR fallback. Bounded credentials are refused on unconverted routes, and legacy credentials are refused on operation-based routes.

Migration and activation constraints:

- The existing Token GORM migration adds a nullable column, without a default ceiling or backfill. Deploy the schema before native writes. No shared migration was run here.
- All production registrations and issuance DTOs remain unchanged. Policy/binding write hooks still reject format 4. This increment is infrastructure for the first real customer journey, not its activation.
- Rollback before native data is issued can leave the unused nullable column. Once native credentials/policies exist, retain a native-capable binary that enforces ceilings. Older binaries ignore this column; clearing ceilings or converting roles back to v3 is not a safe rollback strategy. Mixed-version rollout and rollback still require rehearsal before activation.
- Authority reloads on each request. Revocation/narrowing affects subsequent requests; this adds no transaction snapshot or synchronous cancellation of accepted work.

Validation completed:

- Existing authz, account, root app, auth/org/API-surface, and MCP tests passed normally and with the race detector. All middleware, API infrastructure, account-service and token-service tests passed. `go vet` passed for the changed authorization/authentication packages and root app. Affected-package `go generate`, formatting and diff checks completed.
- A disposable PostgreSQL 17 probe passed 159 assertions under the race detector, using the actual token model, afterquery plugin, authentication, org, global, API-surface and stderr middleware, and native point/collection guards. It exercised real bearer/cookie/query credentials rather than injecting account identity.
- The probe added the ceiling column to a token table without it, reran AutoMigrate, and verified the existing token remained SQL NULL. Normal token creates/saves accepted valid ceilings and rejected invalid ones; authored references survived persistence, and ordinary JSON omitted ceilings. JSON `null` remained distinguishable from SQL NULL after a database read.
- HTTP checks covered account/ceiling intersections, selected/wildcard/future installs, read/write asymmetry, empty grants, list/count filtering before pagination, changed ceilings, narrowing/restoration and binding revocation, deleted identities/roles/policies/tokens, expiry, wrong org/listener/API, unsupported token types/sources/routes, v3/native mismatch, database failure, and sanitized denial responses. Legacy read/admin and OAuth read-only controls retained their expected behavior. A legacy admin account with a corrupt native token could not fall back to legacy authorization.
- This used a full Token table and minimal authority/resource tables, not the full service migration stack, deployed API, production data or UI. Native policies were fixture-seeded because native writers remain disabled. Source-chain lifecycle, issuance/delegation, managed-account writes, rotation and mixed-version rollback were not verified. The disposable probe and container are removed after validation; no permanent tests were added.

Next: implement the native owned-policy and bounded-token write contract with delegation checks, then exercise create → use → narrow → revoke through actual install/job/log endpoints. Broad route/writer inventory remains required before exposing bundles as customer capabilities.

## Naming cleanup

The implementation now uses domain names rather than the temporary `Native` prefix. Earlier checkpoint entries retain historical names; the current contract table above uses the replacement names.

| Responsibility | Current names |
| --- | --- |
| Grant documents and catalog | `Policy`, `Grant`, `GrantScope`, `DefaultCatalog` |
| Point and collection evaluation | `Target`, `Check`, `Query`, `Resources` |
| Authenticated request authority | `Authority`, `AuthorityInput`, `LoadAuthority`, `LoadAccountGrants`, `Role.CompileGrants` |
| Operation guards | `OperationRoute`, `CollectionRoute`, `CollectionIDs`, `OperationAPI` |
| Token limit | `Token.PermissionCeiling`, `permission_ceiling` database column and Temporal field |
| Existing v3 compatibility | `LegacyPolicy`, `DecodeLegacyPolicy`, `LegacyRolePolicy`, `LegacyResourcePolicy`, `LoadLegacyAccountPermissions`, `Role.CompileLegacyPermissions` |

This is a naming-only refactor. Permission IDs, format numbers, JSON policy shapes, catalog membership, and authority selection remain unchanged. The token column has not been deployed, so its final name replaces the provisional name without a shared-database rename migration. No compatibility alias is added for these unshipped Go interfaces.

Validation: existing authz, account, app, middleware, API, MCP, account/token/org service tests passed; targeted race checks and vet passed. A disposable PostgreSQL probe passed 25 checks under the race detector for the renamed column and model save/load, Temporal field tag, unchanged v3 JSON, point/SQL ceiling decisions, every supported scope kind's SQL predicate, and real authentication through operation/collection guards. The probe used minimal authority tables and supplied target rows; it did not repeat the earlier full domain-ownership matrix or a deployment. Formatting, affected-package generation, and stale-symbol/diff checks completed. No permanent tests were added; the probe and disposable database were removed.
