# Native grants: operation bundles and explicit scopes

**Spec 008 supplement · Revised proposal after Oracle review · 16 November 2026**

This proposes the next RBAC foundation, not deployed enforcement. It replaces this supplement's earlier anchor/selector/path design with operation sets and a small supported-scope table. It supersedes the future-model choices in [spec 006](../006-control-plane-authorization/authorization-model.md) that distinguish org bundles from resource policies or freeze bundle membership. It leaves [008's shipped contract](prompt.md), current authorization behavior, and deferred security work unchanged until an explicitly verified cutover.

The local grant-preservation change is a prerequisite, not this model's implementation. [Spec 009](../009-service-account-permissions/prompt.md) defines the coordinated delivery plan spanning this foundation, broad operation coverage, and the granular service-account API and permission picker. The pure catalog/matcher is the first engineering milestone, not a standalone product release or a stopping point. This supplement owns model semantics; spec 009 owns end-to-end delivery. Runner-shadow work and unrelated security cleanup remain excluded.

## 1. One grant, two independent questions

- **What:** an atomic operation or a named bundle of operations.
- **Where:** one resource kind plus explicit IDs or `["*"]` for all resources of that kind within the trusted org.
- **Assignment:** existing roles contain these scoped grants; existing account-role bindings assign those roles.

Bundles are code-defined, acyclic sets. They may include other bundles, but expansion only produces operation IDs. There are no anchors, selectors, traversal expressions, or intermediate resource paths in bundle definitions.

Scopes are matched by ordinary Go code against trusted ownership facts. Supported permission/scope combinations are explicit. Unsupported combinations are rejected, not approximated by a broader grant.

Native account authority is the union of its grants. A credential ceiling, where supported, must independently allow the same operation on the same resource. Existing tenant, live-binding, listener, token, admin, workload, and operation-specific checks still apply.

## 2. One persisted shape

Use a new document format in the existing JSONB policy column, rather than changing the meaning of stored v3 names:

```json
{
  "format": 4,
  "grants": [
    {
      "scope": {"kind": "install", "ids": ["install_a", "install_b"]},
      "permissions": ["install.read"]
    },
    {
      "scope": {"kind": "runner-job", "ids": ["job_b"]},
      "permissions": ["runner_job.cancel"]
    },
    {
      "scope": {"kind": "log_stream", "ids": ["stream_c"]},
      "permissions": ["log_stream.read"]
    }
  ]
}
```

IDs here are fictional shorthand. An org reader has the same shape: `org.read` with `scope: {"kind": "org", "ids": ["org_acme"]}`. An org admin combines `org.read` and `org.write` at that org. No separate `bundle` versus `resource_rules` policy variants remain in the new format, and format 4 has no singular `id` alternative.

The trusted role and binding provide the tenant. Every selected resource must resolve within it when used; IDs never supply or override the tenant. An org scope can explicitly name only that org, or use `["*"]` to select that same org, never all orgs. The role's org does not itself grant org-wide access. Preserve `Role.Version` bookkeeping.

New v4 customer policies omit `api`: capabilities and scopes define their authority. Retain optional `api` decoding and enforcement as a compatibility constraint on existing policies and credential ceilings. Standard-role conversion preserves an existing `ctl-api` restriction or its absence; it neither adds a restriction to an unrestricted source nor removes an explicit restriction. Legacy workload writers retain their API restrictions until their endpoint authorization has capability-based equivalents. This is not a token audience system, and omitting the field does not make an unintegrated listener grant-aware. Listener compatibility remains a separate release gate; no legacy fallback or runner-shadow policy is implied.

Require non-null `grants` and `permissions` arrays; empty arrays grant nothing. Each scope requires a known `kind` and a nonempty `ids` array of valid IDs, or exactly `["*"]`. Reject missing/null/empty selections, scalar strings, wildcard-plus-ID mixtures, repeated wildcard entries, and partial globs such as `install_*`. Duplicate explicit IDs may be deduplicated; order has no meaning. Reject malformed documents, duplicate JSON keys, unknown fields/permission IDs/formats, and unsupported scope combinations. Equivalent duplicate grants are harmless. Keep scope and permissions paired throughout evaluation.

Preserve authored permission references and their scopes through persistence, loading, and editing. Compiled operation sets are derived data, not a replacement for the document: saving `install.read` must not replace it with today's leaf operations. Canonicalization may normalize order and duplicates, but must not exchange a bundle for an equivalent current leaf set or infer a bundle from explicit leaves. Removing a member of a selected bundle requires an explicit conversion to its current leaf operations, followed by editing that set. That selection no longer receives future bundle members; other overlapping grants still apply. No hidden bundle-minus-deny representation is introduced.

### One, several, or all resources

| Scope | Selection within the trusted org |
| --- | --- |
| `{"kind":"install","ids":["install_a"]}` | Install A |
| `{"kind":"install","ids":["install_a","install_b"]}` | Installs A and B |
| `{"kind":"install","ids":["*"]}` | All current and future installs |

An all-installs reader uses this grant:

```json
{
  "scope": {"kind": "install", "ids": ["*"]},
  "permissions": ["install.read"]
}
```

Several IDs are equivalent to separate grants with identical permissions, one per ID. `["*"]` selects the current set at evaluation time, not a snapshot of today's IDs. It selects resources, not operations: no wildcard permission names or resource kinds are introduced. Unknown, deleted, wrong-kind, or foreign-org resources cannot match an explicit ID or a wildcard.

This intentionally differs from [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#referring-to-resources): Kubernetes uses `resourceNames` for exact names and omission/empty names for unrestricted selection; it does not treat `resourceNames: ["*"]` as all names. Here, all-resource selection must be explicit. Kubernetes also does not automatically grant `pods/log` access through `pods`; our bundles and named job-log rule provide that relationship deliberately.

Policy names are optional classification hints, never authority. Verify contents, scopes, role type, and listener restrictions before migration; a policy named `admin` is not sufficient evidence of org-admin authority.

## 3. Bundles expand operations, not resource paths

An atomic permission names an existing operation/response contract, not an isolated database row. `.inspect` distinguishes a direct read operation from an expandable `.read` bundle. It does not require one permission per endpoint.

Concrete read catalog for this slice:

```yaml
org.read:
  - org.inspect
  - app.read
  - install.read
  - runner_job.read
  - log_stream.read
app.read:
  - app.inspect
  - install.read
  - runner_job.read
  - log_stream.read
install.read:
  - install.inspect
  - runner_job.read
  - log_stream.read
runner_job.read:
  - runner_job.inspect
  - log_stream.read
```

For example, `install.read` expands to `{install.inspect, runner_job.inspect, log_stream.read}`. The grant's install scope limits those operations to the selected installs and their supported owned resources. `org.read` expands further, and its org scope includes supported app-build and org-owned resources too. An install wildcard never becomes an org scope during expansion.

The same mechanism supports write bundles. For this slice, `install.write` includes `install.update`, `install.delete`, and `runner_job.cancel`. `org.write` must explicitly include the operations allowed by current broad org-write authority, including destructive operations. Writes do not imply reads; an admin gets both bundles. No suffix wildcard enrolls future operations automatically.

Broad `.write` is not an Editor preset. Preserve its destructive compatibility meaning; the later routine-write catalog introduces `.edit` bundles excluding destructive lifecycle and access-administration operations. Reader/Editor/Admin presets are recipes for ordinary grants, not special evaluator behavior or a new policy shape. Their exact memberships and supported scopes are defined when those operations are ready; they do not enlarge this slice's scope table.

This worked example is not the complete catalog. The local implementation now adds explicit state, stack, runner, telemetry, credential-bearing and other domain operations. See the [delivery ledger](../009-service-account-permissions/delivery-ledger.md) for implemented coverage, verification and outstanding activation gates; the example above does not freeze bundle membership at its original three leaves.

The code-owned catalog is also the source for future authoring metadata: stable operation/bundle IDs, titles, descriptions, supported scopes, and explicit bundle membership. Distinguish definitions from combinations ready for customer assignment and from the current actor's delegation authority. Readiness and display metadata do not themselves grant access. A later catalog API and picker consume these definitions rather than maintaining a second permission registry; neither is part of the first implementation increment.

Validate references and cycles when building the catalog. Flatten membership once per catalog revision. Bundle names remain useful for authoring and explanations but are not runtime authorization targets.

## 4. Deliberately limited scope combinations

These combinations define the initial worked slice, subject to implementing its owner types and route/query contracts:

| Permission reference in a grant | Supported scope kinds |
| --- | --- |
| `org.read`, `org.write`, `org.inspect` | org |
| `app.read` | app |
| `install.read`, `install.write` | install |
| `runner_job.read` | runner-job |
| `app.inspect` | org, app |
| `install.inspect`, `install.update`, `install.delete` | org, app, install |
| `runner_job.inspect`, `runner_job.cancel` | org, app, install, runner-job |
| `log_stream.read` | org, app, install, runner-job, log_stream |
| `log_stream.append` | log_stream initially |

This table validates references as authored in grants. Each allowed scope kind supports one ID, several IDs, or `["*"]`; cardinality does not change which permission/kind combinations are allowed. Nested bundle references only reuse operation sets: they are not separately scoped grants and do not re-anchor anything. Thus `org.read@org` can contain `install.read` even though a directly authored `install.read@org` is rejected. Validate that every flattened operation has a matcher for the containing bundle's allowed scopes. Broader write matching must be defined alongside the complete `org.write` catalog before enabling it; it is not inferred from a name.

For each allowed operation/scope pair, use either same-resource matching, verified owner containment, or the single job-to-stream read rule below. There is no upward access and no generic traversal of arbitrary relationships.

**All installs without app-level access:** use `install.read` with `kind: "install", ids: ["*"]`, not an org scope. For a target to match, it must be a selected install or have a verified selected install in its ownership ancestry. App-build logs have no such install ancestor and therefore do not match. This includes future installs without enumerating them or introducing a special `org_installs` scope kind.

**Remaining limits:** `install.read@org`, `runner_job.read@install`, and `org.read@install` are unsupported. Wildcard selection does not add arbitrary combinations or paths. `runner_job.read` at `runner-job:["*"]` selects all org jobs and their validated associated streams, including app-build jobs; it does not mean only jobs beneath installs.

Atomic `log_stream.read@install` does work: it reads install-owned streams, including streams with no associated job. Combining atomic job inspection and log reads at install scope does not mean "only streams associated with jobs."

## 5. Ownership and one related-resource rule

Resolve ownership from server records, not request-supplied ancestry or the runner that happened to execute a job:

```text
org
  app
    install A
      deploy/action operation
        job A
        stream A
    install B
      deploy operation
        job B
        stream B
    component build
      build job
      build stream
  org-owned operation
    org job / stream
```

Installs use their app/org ownership. Jobs and streams use their actual polymorphic owners: deploy, action, sandbox, build, workflow, or runner operation as supported by the relevant constructor. Preserve supported org-owned operations. A shared runner or referenced component build does not make a resource an install descendant. All edges must stay within the trusted tenant.

Jobs and streams often share an owner; a stream does not universally belong to one job. For `log_stream.read` at job scope, use one explicit read-only rule: the job's `LogStreamID` references the stream, both belong to the tenant, and their owner pair is supported and verified. Initially support same-operation/build owner pairs proven by constructors; inventory others before migrating their callers. No relationship DSL or generic graph walker is needed.

That rule also makes `runner_job.read@job` include the associated stream. With multiple job IDs or a job wildcard, at least one selected job must establish the validated relation to the requested stream. A job wildcard does not grant streams without a related job. If several jobs share a stream, access covers the **whole stream**, not one job's lines. Reading a shared stream does not itself grant sibling-job endpoints, the owner's endpoint, append, or a writer credential.

`ParentLogStreamID` is not authorization inheritance. Existing ingestion copies child records into parent streams: parent readers can see those records, and child appends can affect the parent. Preserve those payload/side-effect contracts rather than adding parent grants or promising record-level isolation.

Missing/foreign/unsupported ownership cannot establish native scoped access. Database failures are operational errors, not a reason to try broader authority. Leave affected routes on their existing checks until their owner cases are supported.

## 6. Observable examples and response boundaries

| Grant | Allows | Does not grant |
| --- | --- | --- |
| `org.read@org_acme` | Supported org, app, install, job and log read operations, including app-build logs | Other-org reads or mutations |
| `app.read@app_shop` | App reads and supported install/build/job/log reads owned by that app | Other apps or org-only operations |
| `install.read@install_a` | Install A reads, its jobs and owned logs | Install B, app-build logs, parent app endpoint, mutations |
| `install.read` at `install:["install_a","install_b"]` | Reads of A and B and their supported jobs/logs | Install C or app-build logs |
| `install.read` at `install:["*"]` | List/read all current and future org installs and their supported jobs/logs | App endpoints, app-build logs, org-only operations, other-org installs, mutations |
| `install.inspect@install_a` | The existing install inspection response | Separate job/log endpoints solely because they belong to A |
| `runner_job.read@job_a` | Job A and its validated associated whole stream | Sibling-job endpoints, unrelated streams, append |
| `runner_job.read` at `runner-job:["*"]` | All org jobs and their validated associated streams | Streams without a related job, install/app endpoints, mutations |
| `log_stream.read@install_a` | All supported install A streams, with or without jobs | Install/job detail endpoints or app-build logs |
| `log_stream.read` at `install:["*"]` | Streams owned beneath any org install, including streams without jobs | App-build logs or org-only streams |
| `log_stream.read` at `log_stream:["*"]` | All supported org streams, including app-build logs | Job/install/app endpoints or other-org streams |
| `log_stream.read@stream_c` | Stream C's existing payload, including copied records | Other stream endpoints or job details |
| `install.read@install_a` plus `runner_job.cancel@job_b` | Read A; cancel B subject to current state constraints | Read B or cancel A |
| `log_stream.append@stream_c` | Existing append operation and its parent-copy side effects | Read, credential minting, or direct append to other streams |

**Endpoint access is not field-level isolation.** Install responses currently embed app data; deploy responses can embed job, plan, build, and log metadata. `install.inspect` preserves that representation. Denying the separate app/job endpoint does not remove embedded information. Narrower responses are a separately reviewed behavior change, not part of this foundation increment.

## 7. Evaluation and collection queries

```text
For a server-selected atomic operation and target:
  retain existing authentication, tenant and credential restrictions
  resolve the ownership/relation facts needed by supported grants
  require an account grant containing the operation and matching a selected root
    root kind must match scope.kind and root must belong to the trusted org
    explicit IDs match a member; ["*"] matches any verified root of that kind
    target must be the root, its supported owned resource, or its named job-log relation
  if a native ceiling applies, require it to allow the same operation/target
  retain existing operation-specific restrictions
```

The matcher consumes verified facts; domain code resolves them. Keep provenance for explanations, not permission intersection. Account access through job A and a ceiling through job B can both permit their shared stream S even when no job itself is allowed by both. An absent ceiling follows the existing credential contract; an explicitly empty ceiling grants nothing. Do not redefine existing OAuth restrictions during this change.

Each supported scope/operation pair needs a point matcher and, before activating a collection surface, an equivalent SQL predicate. Apply account-grant OR predicates intersected with ceiling-grant OR predicates **before** pagination, counts, cursors, or aggregates. Use distinct-resource semantics for shared streams. Do not fetch a page then filter it.

Explicit ID arrays become membership predicates on the verified root IDs. Wildcards remove only that ID predicate: retain tenant, root kind, ownership, and any required job-log relation. For install-scoped logs, even `["*"]` requires a verified install ancestor; filtering only by org is incorrect. Do not expand wildcard scopes into an enumerated list during authentication. Overlapping IDs, grants, and shared streams must not duplicate results or counts.

Keep unconverted collections on existing authorization. Prove install listing, one polymorphic job-list query, and one related-stream query before expanding native route adoption. Creation checks an existing parent; wildcard selection does not itself grant creation. Moves/multi-target operations need their own explicit contracts. Do not invent those checks from HTTP verbs or globally bypass org middleware.

**Access matching is not delegation validation.** A point check answers whether an operation on a resource is allowed now; it cannot establish whether an actor may assign a policy or issue a credential with that policy as its ceiling. Such writes also require management authority and containment of the requested evolving capabilities and scopes. An explicit list covering every current install does not justify granting `["*"]`; today's explicit leaf set does not justify delegating a live bundle. Future delegation checks must reason about those references and scope relationships, not enumerate today's resources or reuse point matching as a general containment proof. Implement that validation with the later management/issuance APIs, not in the first catalog/matcher increment.

## 8. Expansion, refinement, and role reuse

Adding `runner_job.metrics.read` to `runner_job.read` gives it to that bundle's holders, including holders of enclosing install/app/org read bundles, within their existing scopes. Atomic-only holders receive no new operation. Review catalog changes as access changes, including any changes to supported scopes or ownership rules.

Atomic IDs preserve their operation/response contracts. If metrics were already part of `runner_job.inspect`, extracting them cannot silently revoke existing inspect access: keep the old contract or plan an explicit breaking migration. Adding a leaf to a broad bundle alone does not preserve atomic-only holders.

**Live expansion is the intended contract for both account grants and credential ceilings.** Selecting `install.read` means the capabilities defined as install reads as that bundle evolves, not a frozen issuance-time operation list. A new member is usable only where both the account grants and the ceiling allow that operation/resource pair; bundle expansion does not broaden the selected resource scope. Adding an unrelated account grant does not add it to an existing ceiling. Review new members against the bundle's advertised purpose: a diagnostic read may fit `install.read`, deletion does not.

Deploy readers/catalog support before enabling new IDs or writers, account for mixed revisions during rollout, and include the catalog revision in compiled-cache keys. Start with request-scoped ownership facts rather than a new cross-request authorization cache.

Wildcard scopes also follow new resources of their selected kind in the trusted org. A wildcard ceiling is not a snapshot of resources at credential issuance. An explicit array fixes the selected IDs, but existing descendant-scope semantics still include future owned children of those resources. A wildcard account grant intersected with an explicit install-A ceiling must remain limited to A and its permitted children.

Scopes stay inside policy grants. Alice reading install A and Bob reading install B therefore need separate concrete scoped role instances, potentially produced by the same template. Bundles are reusable definitions; scoped role instances are not binding-parameterized definitions. Moving scope onto bindings would be a separate schema/product change, not a prerequisite.

## 9. Native-first migration without a false parity promise

New native policies never need an exact legacy response-map projection to be valid. However, generic per-policy conversion followed by union cannot preserve every old account decision: the old reducer retains existing `all` but otherwise overwrites a source with a later contribution. Native union can restore permissions that were previously overwritten. Whether affected mixtures exist in production needs inventory.

Old object permissions also have route-specific meaning. An exact install grant can satisfy some existing log/state guards; mapping it only to install operations loses access, while mapping it blindly to the new install bundle can widen access. Migration must account for actual route decisions, not names alone.

Implement these dependencies within spec 009's coordinated effort, not as serial product releases:

1. **Establish shared contracts and the pure model.** Implement the catalog, supported-scope validation, and matcher without an authorization cutover. In parallel, inventory customer entry points and policy cohorts, and settle the authoring/token API contracts so domain and UI work can proceed.
2. **Prove one real integration path early.** Use the install/job/log journey against disposable PostgreSQL to prove authentication, point/list/count parity, native loading, and credential intersection. Use its evidence to correct shared assumptions before activating broader enforcement; do not wait for a separate read-only product launch before developing other families.
3. **Replace legacy dependencies at their sources.** Add format-4 decoding and projection-independent validation/loading, and update policy writers and reconciliation. Reader support must reach all serving instances before native writes are enabled. Build the APIs, token storage, and picker against these contracts as part of the same effort, rather than extending legacy adapters for them.
4. **Integrate broad coverage, activate only verified behavior.** Compare existing effective account decisions, both relevant role orders, credential restrictions, status codes, and response contracts. Record the authoritative path for each route/cohort. Native grants must not acquire guessed behavior through an old object/CRUD adapter or fall through to unmigrated checks. Each advertised capability needs complete entry-point coverage before exposure.
5. **Remove transitional paths as part of completion.** Every remaining compatibility path needs a named owner, its routes/cohorts, and removal evidence in the delivery ledger. Retire it when readers/writers are converted, decision parity is established or differences explicitly approved, and upgrade/rollback checks pass. Do not OR old and new evaluators. Ambiguous mixtures need verified account-level conversion; unsupported policies need explicit disposition. An unresolved path is tracked unfinished work, not an indefinite extension point.

Migration 156 skipped unsupported data; spec 008's historical census recorded three wildcard hosted-installer policies. That is not a fresh production inventory. Remove old readers based on current evidence, and retain migration utilities needed for supported skip-version upgrades.

The new typed `["*"]` selection does not establish equivalence with historical wildcard policies. Verify their old object/verb/route semantics independently; never convert a legacy wildcard merely because the new syntax also uses `*`.

Adding the document format alone needs no new SQL column or bulk rewrite. Once native documents are written, rollback must retain a reader-capable binary; an old binary or approximate reverse conversion is not a safe rollback.

### Local implementation decisions for broad coverage

The current implementation integrates existing customer HTTP and MCP endpoints before resuming the service-account editor. Format-4 requests use explicit operations and never fall back to v3. Existing v3 requests keep their original route checks until their cohorts are verified and converted. The runtime registration census covers all registered org-scoped customer routes; it does not establish every handler's compatibility or asynchronous execution parity.

**Dynamic automation remains deliberately coarse.** `app.automation.manage` is assignable only at org scope. It covers config activation, connected repositories, branch runs, install synchronization and event-driven execution whose workers discover current/future targets. Known app-owned workflow approvals/retries require both this capability and the continuation operation. Connected-repository action configuration and event-waiting runbooks require it in addition to their ordinary operations. `app.delete` is also org-scoped because its existing cleanup affects descendants. These capabilities describe existing side effects; they do not promise isolated app/install automation. Immediate handlers with a concrete multi-resource selection instead authorize and consume the selected IDs.

**Accepted work keeps its existing execution contract.** HTTP and MCP share domain continuation validation, checking actual workflow ancestry and the underlying execution operation. This does not serialize a grant snapshot into every worker or promise revocation of already accepted work. Unknown or contradictory provenance, including notebook-linked action runs presented through the generic action continuation path, denies rather than guessing an operation. Notebook execution still uses its own declared operation and existing warm-workflow path.

**Standard conversion is narrow and opt-in.** `enable_org_grant_policies` defaults off. It enables format-4 standard Admin/Reader creation and explicit application through the internal reconciliation endpoint; preview is read-only and requires exactly one org. The converter verifies the standard managed role pair, actual old authority, ownership and bindings. It refuses same-org mixtures with custom/private roles instead of unioning their potentially order-sensitive legacy decisions. Reconciliation preserves authored grants; standard-role replacement cannot erase custom grants. Support/Installer/Runner and private policies remain outside this converter.

The standard Admin recipe explicitly includes org reads/writes plus the existing admin-only access-management reads. Reader additionally retains the readable Slack installation-URL capability. No Editor preset or blanket permission-name translation is introduced. Apply serializes conversion with table-wide authority-write locks, whose production impact must be measured. Production inventory, complete-stack journeys, cohort parity, deployment ordering and a reader-capable rollback are activation gates, not completed claims.

## 10. Implementation boundaries and verification gates

Reuse the current owners under `services/ctl-api/internal/`: `pkg/authz/permissions/{catalog,policy,scope,evaluate}.go` for definitions, codec and matching; `pkg/authz/account_permissions.go` and account/role/policy models for loading; existing role constructors/reconciliation for writing; `pkg/authz/require` and domain helpers for explicit enforcement and queries. Remove legacy projection dependencies at their sources, not through a new framework.

Before activating the native slice, validate:

- All section 6 decisions with two orgs, two apps, sibling installs, app-build jobs, org-owned jobs, shared streams, and streams without jobs.
- One/multiple/all-ID selections, duplicate-ID equivalence, and rejection of missing/null/empty IDs, scalar IDs, mixed/repeated wildcards, partial globs, and the old singular `id` field in format 4.
- All-install access includes an install created after the grant, but excludes app-build logs, org-only resources, foreign-org installs, and unsupported/wrong-kind roots. Explicit install lists exclude a newly created third install.
- Rejection of unsupported combinations such as `install.read@org`, including when mixed with valid references.
- Nested expansion, cycles/unknown IDs, scope pairing, and intentional live expansion for bundles versus atomic-only grants; authored bundle references and explicit leaf selections survive round-trips without conversion into each other.
- Ceiling intersection on operation/resource pairs, including shared streams, wildcard/explicit-ID intersections, and asymmetric read-A/write-B grants. A new bundle member works only where both account and ceiling permit it; unrelated account grants do not enlarge the ceiling.
- Verified ownership, unsupported owner types, foreign IDs, deleted bindings, and failures without broadening.
- Point/list/count/pagination agreement using PostgreSQL, with authorized rows beyond the first unfiltered page and overlapping ID sets/grants that must not duplicate rows or counts.
- Actual policy/account hook round-trips and managed-role reconciliation for grants with no legacy projection.
- Differential checks against the existing checker for every migrated cohort and route, including embedded payloads and log fan-out.

These are implementation acceptance gates, not claims of completed runtime verification. Current evidence and remaining gaps are recorded in the delivery ledger. The earlier path-based reference probe does not validate this simpler model.

The [Oracle review](https://ampcode.com/artifacts/spec-008-oracle-critique-of-permission-representation-eb58f52f-f513-4c5d-b2b8-bc817a975609) records why the earlier design changed. Keep the initial catalog/matcher change independently reviewable and free of policy rewrites or authorization cutover, then continue through spec 009's coordinated plan. Catalog APIs, permission-editor UI, delegation validation, persisted token ceilings, audit storage, and required workflow changes belong to that same delivery effort, not a sequence of separately postponed projects. The foundation acceptance gates above remain prerequisites for exposing native access; unrelated security cleanup is not added to them.

### Credential issuance: targeted fixes now, shared contract later

The release-candidate review reproduced a bounded runner-job credential becoming an unbounded Nuon credential through device approval and redemption. Existing route-level delegation guards do not cover every issuer. Deferring an issuance abstraction does not defer this ceiling escape: close confirmed escapes before activating affected bounded credentials. Disabling new issuance alone does not protect credentials already issued.

**Foundation scope: manually audit and fix existing paths.** Inventory credential issuance, exchange, refresh, rotation and expiry extension, including alternate handlers and background producers. Cover browser login, device authorization, OAuth code/CLI exchanges and refresh, ordinary and deprecated service/personal token APIs, OIDC federation, runner bootstrap/cloud exchange/job claims, stack/phone-home, log-stream/action credentials, telemetry exchanges, and internal/admin/seed issuers. The older [audience inventory](issuance-inventory.md) is a starting list of paths, not current coverage evidence or an audience requirement for this release.

For each path, record its trusted authentication source, target account/org/workload, allowed credential types, ceiling handling, lifecycle restrictions, and actual consumers. Distinguish fresh authentication or trusted provisioning from exchanges of existing credentials; an account ID or token type alone does not establish permission to issue a replacement.

- Add targeted checks in existing owners rather than making a new issuance framework a release dependency. Reject bounded credentials in session approval and redemption, including OAuth refresh; approval-time checks alone do not protect existing grants.
- Preserve ceilings and source/org/workload constraints on explicitly supported workload delegation. Copying ceiling JSON alone is not proof that a different credential type or receiver preserves authority. Unsupported bounded exchanges must deny rather than mint a SQL-NULL ceiling.
- Preserve existing unbounded login, refresh and provisioning behavior. Review historical source-token retention and revocation semantics before changing them; do not impose a universal live-parent requirement as incidental cleanup.
- Reproduce confirmed escapes against unchanged code, then verify denial and persisted credential state after each fix. Include unbounded compatibility controls, already-approved code/refresh cases, and supported bounded workload claims. Record audited paths, evidence and unresolved gaps in the delivery ledger; do not label a source census as a completed behavioral audit.

**Follow-up: a shared issuance contract.** After the manual audit establishes real contracts, consolidate common eligibility and restriction-preservation checks below transport/router wrappers so alternate callers cannot omit them. Keep fresh authentication, administrative issuance and workload delegation explicit, without assuming one generic issuer can safely infer purpose. Cover intermediate authorization grants and refresh credentials as well as final token writes. Reuse the audit's behavior checks to demonstrate compatibility during consolidation.

The shared issuance API and its migration are deferred; their exact shape remains a follow-up design decision. A new audience system, the service-account picker and unrelated legacy hardening are not prerequisites for the targeted fixes. This sequencing decision does not defer the separate authorized-resource handoff work or mark any review finding resolved.

### Authorized request handoff: small shared context, domain-owned preparation

**Shared context and initial consumers implemented locally.** Successful authorization results are retained in a transport-neutral `authz.AccessContext`, reusing `Authority`, `Requirement`, `permissions.Target` and `Selection`. HTTP rules publish it through `require.Access(ctx)`; MCP point and multi-operation checks return it explicitly. The runbook execution and README HTTP/MCP paths consume verified install identity. Broader handler adoption remains incremental; there is no new ownership graph, permission evaluator, generic model cache or endpoint-registration framework.

The shared context has three responsibilities:

1. **Authenticated identity:** account and credential IDs, listener API, selected/validated org, and an optional enforced credential-org binding. Never expose the bearer secret. Selected org is not the same as a credential binding; historical token org metadata must not acquire new enforcement semantics through this refactor. Global discovery can have authenticated identity without a selected org or org authority.
2. **Successful operation checks:** retain operation/target pairs, including each target's canonical ID, kind, org, app/install ancestry and verified related jobs. App/install IDs belong to each target, not to singleton fields on the request: one request can touch several installs or apps. Ancestor facts do not authorize operations on those ancestors.
3. **Authorized collection selection:** retain the existing SQL-backed `Selection`, applied before pagination, counts and aggregates. Preserve the current one-collection-rule-per-HTTP-route contract and rejection of conflicting rules. MCP's explicit query compositions remain explicit; do not cache selections by permission or kind alone.

Read surface, with private storage and copied value/slice results:

```go
access.Identity()              // authenticated request metadata
access.Target(operation)       // exactly one distinct checked target, or an error
access.Requirements()          // all successful operation/target pairs
access.Collection()            // the declared collection selection, or an error
```

`Target` is a convenience for single-resource handlers, not a lookup that silently chooses the first matching resource. Identical checks can be deduplicated; conflicting facts for the same resource must not overwrite an earlier result. Missing required authorization results are configuration errors on grant paths, not permission to re-resolve a user reference. Return snapshots; a handler cannot change recorded org/ownership facts by modifying an accessor result. A successful collection selection is not a proof that an arbitrary individual ID passed a point check.

The adapters publish results only after the corresponding checks succeed. `All` publishes its requirements atomically, and inherited rules cannot silently overwrite earlier results. HTTP keeps its current route declarations and exposes the completed result to the handler. MCP receives the same result explicitly from its authorization helper; it does not need a mutable ambient check log shared across tool calls. Request identity and existing authority must agree. No result is reusable across requests or serialized into background work, and database operations must not retain pooled Gin contexts beyond the request lifetime.

**Permission targets are not always the records being operated on.** Notebook execution authorizes an install after verifying a notebook/cell; process shutdown authorizes a runner after verifying process/shutdown records; onboarding authorizes an org after loading the active onboarding record. Retaining only the permission target loses those identities. Keep small typed prepared values in the owning domain for these cases, just as action execution already uses `preparedInstallActionWorkflowRun`. Pass them directly between shared HTTP/MCP domain methods; HTTP adapters may use a domain-owned typed accessor where middleware and handler signatures require it. Do not put arbitrary structs into `AccessContext` or add a shared field for every domain ID.

Resolve name-or-ID inputs once at the boundary. Subsequent domain calls consume that exact record or perform ID-only loads constrained by its verified org/parents. Body references and secondary resources still need their domain relationship checks. Preserve pinned config versions, request hashes and idempotency behavior. Where ownership can change during a write, retain transaction-time checks; the context is not a database snapshot or a permanent authorization capability.

#### Existing HTTP/MCP patterns checked against the design

This is a source-level API-shape audit of registration/helper calls and the representative implementations below, not an all-endpoint security audit or runtime verification. Counts of helper call sites are not route counts: wrappers and deprecated aliases expand differently.

| Existing use case | Handoff needed / design consequence | Source under `services/ctl-api/internal/app/` |
| --- | --- | --- |
| Named install/app/component and MCP runbook/README | Single checked target with org/app/install ancestry; no second name-or-ID resolution after checking | `installs/service/mcp_get_install_readme.go`, `runbooks/service/mcp_run_runbook.go`, `apps/service/access.go`, `components/service/access.go` |
| Org management and creation under a container | Identity plus the checked org/app/install parent; no invented ID for a not-yet-created child | `orgs/service/service.go`, `components/service/service.go`, `runners/service/access.go` |
| Collections and point-plus-list routes | Existing selection and SQL composition; preserve a point check alongside the collection and filter before paging | `actions/service/service.go`, `installs/service/mcp_access.go`, `components/service/mcp_list_builds.go` |
| Bulk config application and build-all | Multiple requirements plus the exact prepared batch. A successfully admitted empty batch stays empty, never falls back to re-enumeration; absence of child requirements alone cannot prove an empty batch was prepared | `apps/service/access.go`, `apps/service/update_app_config_installs.go`, `components/service/build_all_components.go` |
| Branch deletion | Checked app/install targets plus pinned connection IDs in domain preparation; install IDs alone cannot identify which connections may be changed | `apps/service/access.go`, `apps/service/delete_app_branch.go` |
| Workflow approval/retry and underlying execution | Preserve all operation/target pairs, using the shared continuation resolver; one successful check does not replace the other requirements | `installs/service/workflow_access.go`, `installs/service/mcp_workflow_access.go` |
| Notebook cells, action runs and runner protocol children | Domain preparation pins actual cell/config/process/execution/shutdown records even when authorization targets their parent. Runner execution identity is not a general job-ownership edge | `notebooks/service/authorization.go`, `actions/service/authorization.go`, `actions/service/mcp_run_action.go`, `runners/service/access.go` |
| Logs and shared streams | Reuse `Target.RelatedJobs`; do not invent one universal parent job or require all logs to belong to an install | `runners/access/queries.go`, `runners/service/access.go` |
| Onboarding, aggregates and coarse automation | Keep the prepared onboarding record or existing enclosing-capability contract; an org check does not manufacture a list of individually checked children | `onboarding/service/access.go`, `queues/service/access.go`, `policy_reports/service/access.go` |
| Global org discovery and legacy requests | Identity can exist without selected-org grant checks. Preserve existing legacy/global admission; a legacy bypass must not manufacture a successful grant result | `orgs/service/mcp_list_orgs.go`, `mcp/server/auth.go`, `pkg/authz/require/operation.go` (the last path is relative to `internal/`) |

The runbook/README handoffs are implemented. MCP resolves the install reference once, cross-checks its ID/org/app against the checked target, and passes the same record into the domain method. HTTP consumes the checked target; runbook execution loads its exact ID constrained by org/app. Neither domain method reinterprets that ID as a name. Other MCP call sites accept the new helper result but retain their existing domain contracts; this does not claim every handler consumes `AccessContext` yet.

Continue adoption through named-resource handlers, prepared parent/child operations and batches. Reuse domain-owned preparation for HTTP and MCP instead of adding transport-specific copies. On legacy paths retain current authorization and domain validation explicitly; on grant paths a missing handoff must never trigger the legacy resolver as a fallback. This change does not remove legacy enforcement, tighten credential eligibility, grant new parent access, or fix the separately missing connected-repository operation check. Global requests retain existing identity accessors and do not manufacture an org-scoped `AccessContext`.

Verification of this increment used disposable PostgreSQL on port 55457 and temporary Go overlays:

- Live grant loading, immutable identity/target snapshots, related-job slice copying and order-independent deduplication, conflicting ancestry rejection, missing/ambiguous targets, atomic multi-operation failure, and credential ceiling denial passed.
- Actual install ownership queries rejected foreign-org and corrupt app/org ancestry. Point-plus-collection composition retained the point target and filtered SQL before pagination. Repeated collection declarations failed.
- HTTP inherited-rule composition, missing handoff, sequential pooled-request isolation and legacy bypass controls passed. MCP point/multi-operation handoffs, ceilings and legacy bypass controls passed. This was adapter verification, not a full authenticated network deployment.
- The same MCP collision probes failed against the checkpoint and passed with the changes: an install named `acme` with ID `z-install`, plus another install named `z-install` with ID `a-install`, caused baseline runbook queries to use the second install's app and baseline README state reads to use `a-install`. Both now use the first install's verified ancestry/ID. The probes exercised real handlers and downstream SQL selection, stopping before runbook execution or state rendering; they do not establish full Temporal/cloud execution parity.

Full handler adoption, prepared child/batch handoffs (including empty batches), complete workflow continuation execution and authenticated global discovery still need their respective validation before declaring the entire handoff migration complete. Temporary probe sources and the disposable database are removed after validation; evidence logs are under `.amp/in/access-handoff/`.
