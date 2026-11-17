# Granular-access delivery ledger

Implements the coordinated plan in [spec 009](prompt.md), using [spec 008's native model](../008-customer-rbac-foundation/native-grants-model.md). Status describes local work, not production activation. Current implementation owner is this thread; other workstream ownership has not been assigned.

## Active priority: spec 008 endpoint enforcement before spec 009 management

The active deliverable is [spec 008's grant-model supplement](../008-customer-rbac-foundation/native-grants-model.md): granular grants enforced by existing product endpoints, preserving existing access as the migration baseline. Spec 008's original v3 rollout is already shipped; that is distinct from this unfinished enforcement change. Service-account authoring, delegation, token lifecycle APIs and the picker are deferred until grants work on the existing API.

The pure engine and design are checkpointed in local commit `26c68cd1f2`; projection-independent loading, domain-owned install/job/log resolution, and scoped ID queries in `3b2161b8e8`; bounded request authority and Gin guards in `4f27e5983f`; persisted ceilings, bounded authentication and naming cleanup in `55398b12a6`; the first sixteen endpoint registrations and grant-aware authentication/model hooks in `024d99d320`. This foundation-closeout checkpoint adds broad endpoint, MCP, constructor/conversion and authorization API refactor work. All checkpoints remain local and unpushed. Declared routes select grant enforcement for format-4 accounts; v3 requests retain their existing checks. No deployed route or persisted production role has changed.

The subsequent uncommitted containment, owned-policy management, bounded-token lifecycle and supporting model/legacy-writer edits are preserved in the local Git stash named `Deferred spec 009: grant management and bounded-token lifecycle` (stash object `d2e13d0a948a2f79ae27599ef86e44f6f87f5e7c`). The stash includes new files and its verification record; it excludes unrelated spec 007 work. Restore it selectively after foundation integration, not as a dependency on shipping management APIs. Reuse its format-4 model support where required by the foundation.

Foundation integration order:

1. Map current endpoint decisions and policy/credential cohorts to explicit operations. Define verified conversion of existing effective access, not conversion by policy name or a guessed legacy CRUD adapter. Account for the old reducer's order-dependent mixtures.
2. Connect format-4 account grants to authentication and org admission for existing supported credentials, independently of the future service-account creation flow. The committed bounded-static-only path is not the final foundation contract. Preserve current credential restrictions; do not invent unrestricted ceilings or OR old and new evaluators.
3. Convert real install detail/list and job/log handlers together with ownership and SQL filtering before pagination/counts. Verify scoped grants, existing Reader/Admin access, credential restrictions and denial behavior through those actual handlers. Continue related endpoint families against the operation inventory; unsupported families remain explicitly unfinished, not implicitly covered by bundle names.
4. Convert the relevant policy constructors/reconciliation and verified persisted cohorts, then retire their legacy enforcement. New authoring APIs and UI must not become a prerequisite for using grants on existing endpoints. Do not run shared migrations or activate production changes without authorization.

The acceptance criterion is usable granular authorization on existing product routes with demonstrated baseline compatibility, not merely a working evaluator or the ability to mint a scoped credential. The code inventory and registration census are complete for the registered customer surface below; production cohort inventory, full execution parity and rollout rehearsal remain outstanding.

| Shared contract | Implemented behavior | Remaining integration |
| --- | --- | --- |
| Catalog | Explicit domain operations and expandable bundles, including `org.write`; validated references, cycles, scopes and matcher availability; authored references preserved | Assignment eligibility and management UI remain spec 009; a declaration is not proof of successful execution |
| Policy document | Strict format-4 decoding and projection-independent model hooks; standard Admin/Reader constructors available behind a default-off flag; reconciliation preserves grants | Persisted custom/private/deprecated cohorts require separate effective-access conversion |
| Compilation | Live org bindings and credential identity feed request-scoped `Authority`; grant alternatives union and intersect with optional ceiling alternatives | Production activation and mixed-revision rollout remain; catalog changes require recompilation |
| Point matching | Domain-owned resource resolution, explicit operations and continuation requirements across customer HTTP and MCP | Unsupported owner/provenance cases deny; notebook-linked generic action continuation remains unsupported |
| Credential limit | SQL NULL ceiling is absent, empty denies; global/mixed-org loading understands grants. Existing protocol restrictions remain, including human-only/non-static MCP admission | Issuance/delegation and rotation UI remain shelved; full historical credential and rollback rehearsal outstanding |
| Collections | Domain SQL filters before counts, limits and pagination; whole-org aggregates require org scope; remote-provider lists require their enclosing capability | Production-scale plans and successful ClickHouse/cloud payload reads unverified |
| Product API | Existing endpoints consume grants without waiting for the permission editor; old role setters preserve custom grants and permit exact standard-role replacement | Spec 009 catalog/preview/owned-policy/token lifecycle APIs and picker remain deferred |

The v3 catalog/checker remains separate because IDs such as `install.read` have different v3 meanings. Its document type and decoder are named `LegacyPolicy` and `DecodeLegacyPolicy`. Production accounts and database data have not been converted. A request with captured grant authority takes precedence over legacy authority; mixed v3/v4 bindings in that org fail rather than unioning incompatible semantics. Shared resource-kind validation preserves accepted v3 scopes.

`org.write` explicitly includes destructive writes, automation, credential issuance and access administration. It is not an Editor preset and does not imply reads. The standard Admin recipe combines it with `org.read` and the formerly admin-only service-account/token/trust-policy reads. Reader includes `org.read` and `slack.connect` to preserve the existing readable Slack installation-URL endpoint. These are compatibility recipes, not recommendations for new least-privilege roles.

## Current coverage and intentional boundaries

The disposable runtime census registered 513 customer routes: **486 org-scoped operation declarations, 19 authenticated globals and 8 public exceptions, with zero missing declarations**. This counts deprecated aliases. The census aborts before authorization/handlers; it proves registration coverage, not all-route decision or execution parity. The global VCS callback performs its selected-org operation check inside the handler, in addition to its existing provider-ownership check.

| Family | Implemented integration | Boundary / remaining proof |
| --- | --- | --- |
| Installs and workflows | Reads, lifecycle, configs/inputs, deploys, drift, stacks, approvals/retries/cancellation; actual target ancestry and originating operation checks | Unknown provenance denies; accepted work retains existing execution semantics |
| Apps/components/builds | All 90 app and 46 component/build registrations; filtered collections; pinned selected targets for multi-install writes and build-all | Dynamic automation and app deletion deliberately require org scope; worker/cloud execution unverified |
| Actions/runbooks/notebooks | All 64 registrations; definition ownership, mutation references, filtered runs; distinct configured/ad-hoc/notebook execution operations | Connected repositories and event-waiting runbooks additionally require org automation; generic notebook-linked action continuation denies |
| Jobs/logs/state/queues/reports/bundles | Actual owner resolution, scoped SQL, Terraform relationships, queue boundaries and capability-bearing downloads | Unsupported ownership fails closed; successful ClickHouse/cloud execution and production plans unverified |
| Org administration/integrations/triggers | Explicit capabilities for membership, features, telemetry, webhooks, cloud/VCS/image/Slack connections, trigger credentials/replays; onboarding session ancestry | Coarse org-scoped capabilities preserve current side effects; no separate tenant-boundary redesign |
| Accounts/tokens/OIDC | Existing management endpoints enforce operations; OIDC shares role setter; bounded credentials cannot use old delegation paths to escape ceilings | New grant-authoring/delegation APIs remain shelved |
| Customer MCP | All 51 former legacy authorization call sites use explicit operations; captured org authority, filtered queries and shared HTTP/MCP continuation validator | Existing human-only and non-static admission retained; no new service-account MCP support |
| Globals/public/other listeners | Existing global/public admission retained; mixed-org identities load without projecting format 4 into legacy maps | Runner, stack, Slack and employee listeners keep existing contracts; no workload-policy conversion |

`app.automation.manage` is an org-scoped capability for configuring or dispatching automation whose existing workers discover current/future targets. This includes config activation, connected repositories, branch runs, install sync, and event-driven runbooks. Known app-owned continuations also require their explicit approve/retry operation. This does not claim target-by-target asynchronous delegation. Immediate multi-resource handlers with concrete selections authorize and consume those selected IDs instead.

## Authorization API ergonomics

The refactor binds the Gin router and API once with `require.Routes`. Registrations take `Point`, `All`, `Collection` or `Organization` rules; `For` adapts rules to existing deprecated-route registration and supports multiple guards without nested router wrappers. Declaration metadata still precedes inherited authentication, while checks run afterward. Permission IDs, catalog membership, domain ownership queries and credential admission remain unchanged.

`authz.RequestAuthority` centralizes request identity and existing legacy selection. Transport-neutral `Requirement` / `Authority.Require` supply the shared operation checks. Domain `ResourceSet` definitions bind a resource kind to its ownership query; callers use `.In(db, orgID)` for resolution and `.Query(db, orgID)` for SQL composition. `Authority.Select` returns a `Selection` with shared query scopes and an ID-subquery escape hatch. HTTP and MCP use these primitives instead of duplicating identity checks, kind arguments and collection filters.

HTTP and MCP action execution reuse the authorized preparation for non-idempotent requests. Idempotent action requests retain their existing preparation hook, request hash and replay behavior; this refactor does not move preparation outside that workflow contract.

Verification against the exact pre-refactor worktree: **506 public registrations across 20 changed service families** have identical methods, paths, handlers, 486 pre-auth operation markers and 71 deprecated-route audit entries. This is the changed-family census, not a replacement for the earlier full 513-route census. Existing package checks passed with `INTEGRATION` unset; authz/auth-middleware/action-service race checks and targeted vet passed. A disposable PostgreSQL probe under the race detector exercised shared HTTP/MCP checks, ceiling intersections, wildcard/future resources, ownership/deletion rejection, list/count/pagination, legacy create-only behavior and point-versus-collection error translation. Evidence is under `.amp/in/authz-api-refactor/`. Full Temporal/cloud execution and environment-gated integration suites were not rerun. No permanent tests, production changes or new commits were made.

### Foundation review closeout

Oracle's source review found no confirmed blocking refactor regression, but identified that composing collection rules silently replaced the first selection. Grant requests now reject a second collection rule as a configuration error before resolving its resources, including when inherited through a parent group. The first selection is retained. Point-plus-collection composition and the existing v3 path are unchanged.

A temporary HTTP probe failed on unchanged code for direct, inherited and repeated collection rules, then passed all six scenarios under the race detector after the fix. Controls covered a single collection, point-plus-collection, legacy bypass and sequential requests through Gin's context pool. The probe checked PostgreSQL-dialect query construction without executing database queries; it does not add database or full-stack evidence. Logs are under `.amp/in/collection-rule-closeout/`; temporary probe sources were removed. No permanent tests were added.

This closes the code-review/checkpoint step, not production readiness. Next is persisted customer-policy inventory, remaining conversion/writer coverage and legacy-retirement planning, followed by complete-stack release verification. Spec 009 management work remains shelved.

## Guarded standard-role conversion

`enable_org_grant_policies` defaults off. When enabled, newly created orgs receive format-4 standard Admin/Reader roles; Support, Installer, Runner and private role families are not converted. Reconciliation preserves documents and can recreate a missing standard role in an already converted org. Both ordinary role assignment and OIDC refuse to replace custom/scoped grants through the singular-role API.

The existing internal org-role reconciliation endpoint accepts `grant_policies` and `apply`. Conversion requires exactly one org and previews by default; apply additionally requires the flag. It accepts exactly the managed standard Admin/Reader pair, validates policy ownership and actual v3 projection, rejects inconsistent/duplicate bindings and accounts mixing standard roles with other same-org roles, and is idempotent for exact converted recipes. Names alone never establish eligibility.

Apply uses a serializable transaction and locks `roles`, `policies`, and `account_roles` against concurrent writes. Those are table-wide locks, so production lock duration and operational impact must be measured before use. Dry-run is read-only. No conversion has been applied to shared data. Custom/private/deprecated policies, ambiguous old overwrite-order mixtures and historical wildcard policies require explicit disposition, not an approximate union migration.

## Combined verification and activation gates

- The final combined PostgreSQL probe passed **1,128 assertions**, including the 513-route declaration census, HTTP/query checks, conversion checks and ten MCP continuation calls with persisted mutation assertions. The MCP rerun used real registered tools over the SDK's in-memory transport with captured authority contexts; protocol authentication was exercised separately by the earlier transport probe. Evidence: `.amp/in/workflow-foundation-probe.log`.
- Disposable PostgreSQL probes exercised real HTTP bearer/org/operation middleware, model hooks and scoped queries; standard constructor selection, preview/apply/idempotence, reconciliation, role replacement, mixed-role rejection and actual Admin/Reader before/after controls passed. The Admin recipe also allows every catalog leaf at each supported target kind in the synthetic matrix.
- Domain probes verified app/component and execution-family ownership, asymmetric grants, pre-pagination filtering, mutation references and credential intersections. MCP transport verification exercised 48 tool calls and three authentication checks, including mixed-org v3 controls and persisted approval/cancel mutations. HTTP and MCP now share the domain continuation validator.
- Existing affected-package tests pass with `INTEGRATION` unset; this does not run the environment-gated integration suites. Shared authz/account/auth-middleware/MCP-server tests also passed with `-race`. Evidence: `.amp/in/foundation-combined-tests-final.log` and `.amp/in/foundation-auth-race-final.log`. Admin schema generation and formatting/diff checks passed. No permanent tests were added, following package guidance.
- The earlier Gin pooled-context race was reproduced against unchanged handlers with legacy credentials; it remains recorded below, not fixed as part of RBAC.
- **Not verified:** full FX/Temporal/runner/cloud execution, successful live ClickHouse reads, comprehensive payload snapshots, all production credential/policy cohorts, production-size query/lock performance, shared schema migration, mixed-version deployment, and rollback rehearsal. Fixtures used real models and selected real views, simplified supporting views, and no foreign keys; this is not a production migration rehearsal.

Before activation, inventory persisted cohorts, run representative existing-account journeys against the complete stack, verify a reader-capable rollback release, and preview the selected org's conversion. Production writes and deployment require separate authorization. Service-account authoring/UI remains deferred; no requirement here authorizes restoring its stash wholesale.

## Compatibility retirement queue

| Path | Why it remains | Removal gate |
| --- | --- | --- |
| V3 decoding and legacy projection in policy/account/role hooks | Persisted customer/workload data and private/deprecated writers still use it | Current cohort conversion evidence and supported upgrade verification; standard conversion tooling alone is insufficient |
| Old object/CRUD request checks and broad org middleware | V3 requests still use the original behavior | Verified cohort conversion, route decision parity or approved differences; declaration coverage alone is insufficient |
| Singular-role service-account replacement | Existing API/UI still use standard roles; custom grants are protected | Owned-policy API and clients replace this writer without erasing unrelated bindings |
| Existing credential admission | Historic protocols and listener restrictions remain intentional | Issuance/revocation/rotation checks and grant-capable rollback verified; do not broaden protocols merely to remove code |

Unconverted paths are unfinished work. Native grants must never fall through to broad legacy authorization. No compatibility path is removed merely because current role policies are believed to be v3.

## Historical: existing endpoint integration (checkpoint `024d99d320`)

Sixteen existing ctl-api registrations now declare operations:

- `org.inspect`: GET `/v1/orgs/current`.
- `app.inspect`: GET `/v1/apps`, `/v1/apps/:app_id`.
- `install.inspect`: GET `/v1/installs`, `/v1/apps/:app_id/installs`, `/v1/installs/:install_id`.
- `runner_job.inspect`: GET `/v1/runner-jobs`, `/v1/runners/:runner_id/jobs`, `/v1/runner-jobs/:runner_job_id`, and its `/plan` and `/composite-plan` variants.
- `runner_job.cancel`: POST `/v1/runner-jobs/:runner_job_id/cancel`.
- `log_stream.read`: GET `/v1/log-streams/:log_stream_id`, `/logs`, `/logs/tail`, `/spans`.

Format-4 bindings select the grant loader on these routes, including existing credentials whose ceiling is SQL NULL. V3 accounts retain their original org/resource/API checks. Explicit ceilings still select the bounded path, and malformed/mixed authority cannot downgrade. App/install detail authorization resolves the same name-or-ID record returned by the handler. The old singular-role replacement helper refuses to erase format-4 bindings; broader writer conversion is still pending.

Disposable PostgreSQL 17 validation passed 118 assertions under the race detector through the actual domain route registrations and handlers, real bearer authentication, org/API/authorization/pagination middleware, GORM model tables and afterquery/view/pagination plugins. Policy and binding fixtures used ordinary save hooks. Checks covered selected/multiple/wildcard installs, future installs, filtered lists and offset boundaries, app/org isolation, app-owned job/log exclusions, inspect versus read, asymmetric cancellation with persisted status, v3 Reader/Admin controls, OAuth read-only denial, optional/empty/malformed ceilings, legacy replacement rejection, narrowing, revoked bindings, mixed/malformed policies, plan redaction, name/ID collisions, runner-scoped lists and rejection of an unconverted mutation. Denied log/tail/span and composite-plan requests were exercised, but successful ClickHouse payload reads and composite-plan reconstruction were not. The actual install view was used; other required views were simplified fixtures, not a full service migration rehearsal. Auth and OAuth token types were exercised on service-account fixtures, plus a human-account static-token read; these were not complete login flows.

Existing authz, account, app-model, all middleware, affected domain-service, account/token-service and MCP tests passed with `-race`; targeted vet, affected-package generation and diff checks passed. This does not imply their environment-gated integration suites ran. No shared data, existing roles, migration, deployment or UI changed. The authoring/token work remains shelved. This is the first real integration path, not completion of spec 008 or permission to migrate current accounts. Install lifecycle writes, remaining route coverage, policy writers/reconciliation and verified cohort conversion remain foundation work before the service-account product resumes.

## Historical: initial install mutations (uncommitted)

Four more existing registrations adopt explicit operations, bringing the total to twenty:

- `install.update`: PATCH `/v1/installs/:install_id`, including its existing metadata/config updates and update/reconcile signals. Its target resolver accepts only an exact ID, matching the existing mutation query.
- `install.delete`: DELETE `/v1/installs/:install_id`, POST `/v1/installs/:install_id/deprovision`, and POST `/v1/installs/:install_id/forget`. This operation includes the existing teardown and cleanup side effects, not just deleting a row. It remains in the live `install.write` bundle; `install.update` alone does not permit these destructive routes.

For grant requests, the destructive routes consume the same name-or-ID record resolved during authorization, rather than resolving an ambiguous path twice. Forget uses that record's canonical ID for labels, synchronous credential cleanup, soft deletion, and the queued org cleanup. Legacy requests retain their original loaders, path handling and authorization, including their existing POST-versus-DELETE distinctions. No legacy permission mapping has been broadened or tightened.

Workflow admission retains the existing runner-state, approval, plan-only, requested execution role and idempotency behavior. Workflows and queue signals retain their existing concrete install/workflow IDs and initiating account/org identity. This increment does not serialize grant authority into work, reauthorize accepted jobs, or change worker execution. Revocation/narrowing is enforced on subsequent HTTP requests, including retries; already accepted work retains the existing execution semantics. A later continuation design must not silently reinterpret this as arbitrary delegation.

Validation:

- A disposable PostgreSQL 17 probe passed 70 assertions through actual bearer authentication, org/API/operation/patch middleware, domain registrations, handlers, GORM hooks/plugins, workflow creation and queue-signal persistence. It checked update/delete/read asymmetry, selected/multiple/wildcard/future installs, excluded/foreign/missing targets, OAuth read-only restrictions, ceiling intersections, name/ID collisions, config upsert, request role/plan-only propagation, signal identity, idempotent deprovision/delete and conflicting retries, disabled-runner rejection, permission narrowing, and v3 Reader/Admin controls for all four routes. Forget removed the selected install's managed credentials while retaining another install's credentials.
- The race-enabled HTTP probe completed its functional assertions but **failed** with Gin pooled-context reuse racing `database/sql.Rows.awaitDone` after writes. An overlay restoring all changed handlers/registrations to checkpoint `024d99d320`, using only legacy Admin credentials, reproduced the same race while passing 40 baseline requests. This pre-existing race is recorded, not fixed in the authorization slice. Logs: `.amp/in/install-writes-race.log` and `.amp/in/install-writes-baseline-race.log`; the successful functional run is `.amp/in/install-writes-http.log`.
- Existing authz, install service/helpers and all middleware tests passed with `-race`; environment-gated integration suites were not enabled. Targeted vet, affected-package generation, formatting and diff checks passed.
- Tables came from the real PostgreSQL model registry with foreign keys disabled for fixtures, the real current install view, simplified supporting views, and the relevant config/workflow/queue uniqueness indexes. This was not a full migration rehearsal, Temporal worker execution, cloud teardown, full historical-cohort conversion, or a production-data test. The disposable probe/database were removed; no permanent tests or shared data changed.

Next foundation work remains the other install reads/lifecycle operations, workflow inspection/control and their continuation contracts, followed by broader route/writer/cohort conversion. Service-account authoring and UI remain deferred.

## Historical checkpoints

The entries below record what was verified at each earlier checkpoint. Their then-current limitations are superseded by the current contract and endpoint-integration sections above.

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
