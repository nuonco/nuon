# Native-access delivery ledger

Implements the coordinated plan in [spec 009](prompt.md), using [spec 008's native model](../008-customer-rbac-foundation/native-grants-model.md). Status describes local work, not production activation. Current implementation owner is this thread; other workstream ownership has not been assigned.

## Current checkpoint: pure engine implemented, integration pending

The first part of M0 is implemented locally in `services/ctl-api/internal/pkg/authz/permissions/`. M0's complete route/writer/cohort inventory and remaining API/query contracts are not finished. No customer-facing capability is newly assignable, and no request path has switched authorization engines.

| Shared contract | Implemented behavior | Remaining integration |
| --- | --- | --- |
| Catalog | `NewNativeCatalog` supplies the initial org/app/install/job/log definitions. `NewCatalog` validates references, cycles, supported scopes and matcher availability, and flattens nested bundles. `Definitions` returns copied authoring metadata; `Revision` identifies normalized definitions | Add verified domain operations through this catalog; readiness/actor-specific assignment eligibility belongs to the API rollout, not catalog existence |
| Policy document | `Catalog.DecodePolicy` strictly decodes format 4; `NormalizePolicy` validates and copies grants, normalizes explicit ID/permission duplicates, and retains authored bundle references. `NativePolicy` is the serialization DTO | DB policy hooks still decode v3 and demand legacy projection; native persistence/loading is not enabled |
| Compilation | `Catalog.Compile` binds the document to a trusted org, validates org selections and creates immutable compiled grants. Compilation uses the current catalog; stale compiled revisions are rejected | Account/binding loader must validate live bindings and provide the trusted org. Recompile authored documents on catalog revision changes |
| Point matching | `Catalog.Match` takes a server-selected atomic operation, listener API and `NativeTarget`; matches same-resource, verified app/install ownership, or the named job-to-log read relation | Domain resolvers must supply trusted facts. Shape validation does not establish existence, tenancy of ancestry, or supported owner pairs |
| Credential limit | Grants union within the account and within the ceiling, then intersect on operation/resource pairs. A nil ceiling means the caller supplied no native ceiling; a non-nil empty ceiling denies everything. Live bundle changes apply after recompilation | Credential loading must preserve existing restrictions and must never turn a missing/failed required ceiling load into nil. No token-ceiling column or issuance path exists yet |
| Collections | Not implemented | SQL predicates must match the point checker before pagination, counts, cursors and aggregates; no fetch-then-filter adapter |
| Product API | Native policy and catalog data types are available | Preview/ETag/owned-policy and token DTOs, transactional writes, delegation checks and API endpoints remain unimplemented |

The v3 catalog/checker remains separate because IDs such as `install.read` have different v3 meanings. Existing `Policy`, `DecodePolicy`, `MatchPermissions`, `CanPerform`, writers, routes and database data have not switched to native authority. Shared resource-kind validation was extracted without changing accepted v3 scopes. The earlier uncommitted native-grant preservation changes remain untouched.

The initial native catalog deliberately omits `org.write`: spec 008 requires its complete compatibility mapping before defining it. `install.write` includes deletion and cancellation; no Editor preset has been invented. Metadata is not a claim that the corresponding routes are ready for customer assignment.

## Coverage queue

This is a starting queue, not a completed route census. All rows still use existing request authorization. Each needs concrete route/helper/continuation entries and evidence before its native capability can activate.

| Family | Source entry points to inventory | Current native coverage | Missing proof / contract |
| --- | --- | --- | --- |
| Org/app reads | `apps/service/service.go`, org middleware, org services | Pure read bundle/point semantics only | Response classification, concrete ownership, all routes and list/count filtering |
| Installs and lifecycle | `installs/service/service.go`, `get_install.go`, mutation helpers/workflows | Pure inspection/update/delete semantics at org/app/install scopes | Scoped SQL, full alternate-path coverage, embedded data contract, referenced-target and destructive-cascade checks |
| Jobs and logs | `runners/service/service.go`, `get_runner_jobs_ctl_api.go`, stream handlers and owner constructors | Pure inspect/cancel/read/append and validated-related-job input contract | Actual owner census, shared streams, stream/download/SSE surfaces and query equivalence |
| Builds, actions, workflows | App/component/action/workflow registrations and helpers | Not integrated | Operation inventory, admission, approval/retry/cancel checks and bounded continuation authority |
| Service accounts and access administration | `accounts/service/service.go`, `service_account.go`, `authz/set_account_role.go` | DTO/model primitives only | Owned-policy writes preserving unrelated bindings, delegation, preview, concurrency and native-aware old-client behavior |
| Credentials | `tokens/service/service.go`, token model, account token helpers/auth loaders | Pure ceiling intersection only | Persisted ceilings, issuance validation, rotation/revocation/lineage, historical-token compatibility |
| Remaining customer families | Section 3 of spec 009 and current registrations, not historical spec inventory alone | Not integrated | Telemetry, state/config disclosure, stacks/reporting, integrations/trust, triggers/delivery, notebooks/reports and explicit transport exclusions |

## Compatibility retirement queue

| Path | Why it remains | Removal gate |
| --- | --- | --- |
| V3 decoding and legacy projection in policy/account/role hooks | Existing persisted data and writers still use it | Current cohort conversion evidence, projection-independent native loading, converted writers/reconciliation and supported upgrade verification |
| Old object/CRUD request checks and broad org middleware | Domain routes have not adopted atomic operations/scoped predicates | Per-route/cohort authority recorded, complete entry-point mapping, baseline decision/response parity or explicit approved differences |
| Singular-role service-account replacement | Existing API/UI still depend on it | Native owned-policy API and clients work; old writers cannot erase native access or unrelated bindings |
| Existing credential admission | Native ceilings are not stored/loaded | Historical restrictions preserved, required ceiling failures deny, issuance/revocation/rotation checks and native-capable rollback verified |

Unconverted paths are unfinished work. Native grants must never fall through to broad legacy authorization. No compatibility path is removed merely because current role policies are believed to be v3.

## Verification recorded for this checkpoint

- Existing `go test ./services/ctl-api/internal/pkg/authz/permissions` passed before implementation.
- Existing `go test ./services/ctl-api/internal/pkg/authz/...` and the same command with `-race` passed after implementation.
- A disposable standalone Go probe executed 3,160 assertions through the new exported APIs: positive/negative operation-target matrices; all/selected installs; app-build/org/foreign exclusions; asymmetric read/write grants; independent shared-stream ceilings; empty versus absent ceilings; strict JSON; unchanged authored references; catalog cycles/unknown references/scope rejection; stable revisions; live expansion versus atomic-only grants; input-copy isolation; and v3 semantic separation.
- The probe is validation, not a retained test suite. No test files were added, following ctl-api guidance. Formatting uses `gofmt` and `goimports`.
- No PostgreSQL ownership/query, HTTP, UI, production inventory, migration or deployment validation has been claimed. The new point checker trusts resolved input; database-backed and request-level checks remain required.

Next integration work connects projection-independent native loading and a concrete install/job/log resolver/query contract while the broader family inventory and service-account API work proceed. Persisted native writes remain disabled until readers and enforcement are ready.
