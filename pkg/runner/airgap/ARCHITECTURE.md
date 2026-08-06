# Air-gap plan compilation and late binding

How an air-gap bundle gets concrete, correct Terraform/OCI plans in a customer
account that the control plane has never seen and can never reach.

Two halves, one contract:

- **Compile time** (vendor side): `services/ctl-api/internal/app/airgap/plan_compile.go`
  renders every job's composite plan from the app config alone — no install
  required — against a synthetic state where every unknowable value is a
  unique placeholder token.
- **Run time** (customer side): `pkg/runner/airgap/latebind.go` re-renders each
  plan just before execution, substituting tokens with ground truth the runner
  observes locally (S3 stack outputs, its own sandbox/component applies).

## How this works in connected BYOC (for contrast)

In the normal connected flow there are no placeholders at all, because
rendering happens **at dispatch time**, when the control plane already knows
every value:

```
VENDOR / NUON CONTROL PLANE                     CUSTOMER ACCOUNT
┌─────────────────────────────────┐
│ ctl-api + Temporal              │
│                                 │            ┌──────────────────┐
│ install state DB:               │  outbound  │ CloudFormation   │
│  • stack outputs   ◀────────────┼────────────│ stack phone-home │
│  • sandbox outputs ◀──────────┐ │            └──────────────────┘
│  • component outputs ◀──────┐ │ │
│                             │ │ │            ┌──────────────────┐
│ per job, at dispatch:       │ │ │   poll     │ runner (egress   │
│ ┌─────────────────────────┐ │ │ │◀───────────│ to ctl-api)      │
│ │ render plan from LIVE   │ │ │ │   jobs     │                  │
│ │ state → concrete values │─┼─┼─┼───────────▶│ execute plan     │
│ └─────────────────────────┘ │ │ │            │                  │
│                             │ │ │  outputs   │                  │
│  workflow engine advances ◀─┴─┴─┼────────────│ report back      │
│  DAG, dispatches next job      │            └──────────────────┘
└─────────────────────────────────┘
        render N times, one per job,
        each time with fresh truth
```

The state lives in ctl-api's Postgres; Temporal owns ordering and retries; the
runner is thin — it never renders, it just executes what it is handed. This
requires egress from the customer account (runner polling, stack phone-home)
and the control plane in the loop for every job.

Air-gapped, every one of those responsibilities moves:

| Connected BYOC | Air-gapped |
|---|---|
| Render at dispatch, per job, from live DB | Render at compile, once, with tokens |
| Postgres install state | S3 state prefix (`steps/*/outputs.json`, tfstate) |
| Temporal workflow advances the DAG | Envelope `DependsOn` DAG, runner walks it |
| Stack outputs phone home to ctl-api | CloudFormation writes outputs to S3, runner reads |
| Component outputs reported to API, injected into next render | Stored in S3 by producer step, bound via `OutputBindings` |
| Runner polls ctl-api for jobs | Runner resumes from `status.json`, retries locally |
| Dashboard/CLI hit ctl-api | `nuon-bundle status/logs/results` + portal read S3 directly |

Late binding is the connected renderer, relocated: `renderPlan` reproduces
control-plane plan chaining and environment rebinding, fed from S3 instead of
Postgres, triggered by the runner instead of Temporal.

## Compile time: synthetic state and placeholder tokens

Since there is no install, `compileState` renders every composite plan against
a state where each unknowable value is a marker:

- **Stack outputs** → `__NUON_AIRGAP_STACK_vpc_id__`,
  `__NUON_AIRGAP_STACK_region__`, ... Builtin keys the planner always injects
  are seeded unconditionally; any `.nuon.install.stack.outputs.*` reference
  found in the config is seeded too.
- **Sandbox outputs** → a config reference like
  `.nuon.install.sandbox.outputs.nuon_dns.public_domain.name` is seeded at
  that **nested path** in the state snapshot, holding the token
  `__NUON_AIRGAP_SANDBOX_nuon_dns_public_domain_name__` (dots flattened only
  in the token name).
- **Install inputs** → `__NUON_INPUT_<name>__`, resolved from the customer's
  inputs file.
- **Cross-component outputs** → tokens plus an `OutputBindings` table in the
  envelope recording which step produces which output path for each token.
- **Cluster outputs** are seeded only when something consumes them, so
  Terraform-only apps (Lambda, ECS) resolve to no ClusterInfo and never demand
  kube auth.

The bundle therefore ships plans that are structurally complete but full of
markers, and each plan carries its placeholder-filled state snapshot under
`state.sandbox.outputs` / `state.install_stack.outputs`.

## Run time: the late-binding pipeline

Every time a job is about to execute, `renderPlan` re-renders its composite
plan:

```
┌─────────────────────┐
│ composite plan JSON │ (from envelope, tokens baked in)
└─────────┬───────────┘
          ▼
 1. chain plan contents      apply-plan steps get the tfplan produced
    (plan_from_step)         by their create-plan step
          ▼
 2. rebind stack outputs     CloudFormation stack outputs read from S3
          ▼                  (stack-outputs/outputs.json)
 3. rebind sandbox outputs   latest finished sandbox step's recorded
          ▼                  terraform outputs
 4. rebind cluster info      only if a sandbox produced a cluster
          ▼
 5. substitute inputs        customer's inputs.json; unresolved → hard error
          ▼
 6. bind component outputs   OutputBindings: read producer step's stored
          ▼                  outputs, error if missing
┌─────────────────────┐
│ concrete plan JSON  │ → handed to the normal runner job handler
└─────────────────────┘
```

The envelope's dependency DAG guarantees a producer step applies (and stores
its outputs in S3 state) before any consumer renders, so an unresolvable token
is a hard error rather than a silently broken deploy.

## The matching trick: structural alignment

Steps 2–3 do not look for tokens explicitly. `rebindSandboxOutputs` and
`collectDeepSubstitutions` walk the old snapshot embedded in the plan and the
fresh outputs **in parallel**; wherever the same path holds two different
strings, they record an `old → new` substitution, then rewrite every string in
the plan that matches an `old` value.

For a zero-install bundle the "old" values are the tokens
(`snapshot.nuon_dns.public_domain.name = "__NUON_AIRGAP_SANDBOX_..."`), so
alignment against the real sandbox apply outputs yields
`token → actual domain`. For a reference-install bundle (see below) the "old"
values are the vendor reference install's real outputs. One code path serves
both.

Substitution safety rules:

- Values shorter than `minRebindValueLength` are never substituted, so
  `"true"`, `"1"`, or availability-zone letters cannot corrupt unrelated
  fields on a coincidental match.
- Observed-**value** substitutions are exact-match only (whole string, or a
  whole comma-separated element).
- Placeholder **tokens** (`__NUON_AIRGAP_*__`) are globally unique markers and
  are additionally substring-replaced, so tokens embedded in larger strings
  (`*.<token>`, `https://<token>/path`) still resolve.

## History: the reference-install export approach

The first implementation required a real, connected, fully deployed install
(the "reference install"). Bundle export grabbed that install's
already-rendered plans, including every concrete value the control plane had
resolved: the reference VPC ID, subnet IDs, zone IDs, ECR URLs, IAM role ARNs.
At the customer, late binding worked purely by value alignment.

```
OLD: reference-install export
┌──────────────┐    ┌─────────────────────┐   export   ┌────────────────────┐
│  app config  │──▶ │ REAL install        │ ─────────▶ │ bundle with REAL   │
└──────────────┘    │ (reference env)     │            │ values baked in    │
                    └─────────────────────┘            └─────────┬──────────┘
                                                                 ▼
                                                    rebind by VALUE match:
                                                    "vpc-0285…" → "vpc-0e76…"
                                                    (miss → silently wrong)

NEW: zero-install compile
┌──────────────┐   compile vs synthetic state   ┌──────────────────────────┐
│  app config  │ ─────────────────────────────▶ │ bundle with TOKENS + DAG │
└──────────────┘                                └─────────────┬────────────┘
                                                              ▼
                                                 bind by TOKEN (exact marker)
                                                 unresolved → HARD ERROR
```

Why it was replaced:

- A live install was a hard prerequisite for publishing a bundle.
- The reference install's environment leaked into the bundle: exported plans
  baked the vendor's stack outputs into rendered terraform vars, and the
  customer's stack is unknowable at export time, so republishing could never
  fix it.
- Value-based rebinding is heuristic; anything derived, concatenated, or
  coincidentally colliding was fragile.
- The reference install had to stay healthy and re-deployed on every config
  change or the bundle exported stale values.

| | Reference-install export | Zero-install compile |
|---|---|---|
| Prerequisite | Deployed, healthy install | App config only |
| Plan values | Reference env's real values baked in | Placeholder tokens |
| Rebinding | Old value → new value (heuristic match) | Token → value (exact, unambiguous) |
| Cross-component outputs | Baked reference values, rewritten by alignment | Explicit DAG: producer step → output path → token |
| Failure mode | Silent wrong value if alignment misses | Hard error on unresolved token |
| Bundle staleness | Drifts with reference install | Deterministic from config |
