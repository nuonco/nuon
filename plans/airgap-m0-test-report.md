# Airgap M0 — end-to-end test report

Date: 2026-08-06
Workspace: `/Users/harsh/work/nuonco/ws-airgap-m0` (jj bookmark `ht/airgap-m0`, base `main@origin e957d343`)
Test artifacts: `/Users/harsh/work/nuonco/airgap-m0-test/` (outside the repo)

## Goal

Prove `runner airgap` can execute an exported plan envelope with **zero ctl-api
access**: local job queue, disk state store, loopback Terraform HTTP backend,
and offline late binding (create-plan → apply-plan chaining, cluster
rebinding, forced default cloud auth). Then evaluate whether the existing
`eks-simple` install's real plans can be replayed the same way.

## Verdict

- **Synthetic end-to-end: PASS.** A two-step `sandbox-terraform`
  create-apply-plan → apply-plan envelope deployed real AWS resources (S3
  bucket + SSM parameter) with `NUON_API_URL`/`NUON_API_TOKEN` unset and
  vendored Terraform binary + provider mirror (no registry egress).
- **Kill/resume: PASS.** `kill -9` between create and apply, then re-invoking
  with the same `--state` dir, resumed and finished cleanly.
- **Real `eks-simple` replay: NO-GO (blocked, by design of the fixture).**
  See "Real install evaluation" below.

## How it was run

```bash
go build -o /Users/harsh/work/nuonco/airgap-m0-test/runner ./bins/runner

env -u NUON_API_URL -u NUON_API_TOKEN \
  AWS_PROFILE=stage.NuonAdmin \
  /Users/harsh/work/nuonco/airgap-m0-test/runner airgap \
  --plan /Users/harsh/work/nuonco/airgap-m0-test/envelope-run4.json \
  --state /Users/harsh/work/nuonco/airgap-m0-test/state-run4
```

The fixture (`tf-min/`) uses a `local_archive` Terraform source (no Git/OCI
fetch), vendored Terraform `1.11.0` (`darwin_arm64`), and a filesystem mirror
of `hashicorp/aws` `5.100.0`. AWS account `676549690856`, region `us-west-2`.

Caveat: this proves ctl-api independence, not full network isolation — the
AWS provider used normal connectivity to reach AWS APIs. A point-in-time
`lsof` check showed no unexpected connections, but that is not rigorous
egress enforcement.

### Successful runs

| Run | What it proved | State dir |
| --- | --- | --- |
| run2 | Clean two-step create→apply, valid backend state (v4, 3 resources) | `state-run2/` |
| run3 | Exposed backend-port resume bug; succeeded after manual port restore | `state-run3/` |
| run4 | SIGKILL after create-plan finished; automatic resume applied the saved plan | `state-run4/` |

run4 outputs:

```json
{
  "account_id": "676549690856",
  "proof_bucket": "nuon-airgap-m0-proof-run4-676549690856",
  "proof_parameter": "/nuon/airgap-m0/proof-run4"
}
```

## AWS resources left running (intentionally — do not delete)

S3 buckets (account `676549690856`, `us-west-2`):

- `nuon-airgap-m0-proof-676549690856`
- `nuon-airgap-m0-proof-run2-676549690856`
- `nuon-airgap-m0-proof-run3-676549690856`
- `nuon-airgap-m0-proof-run4-676549690856`

SSM parameters:

- `/nuon/airgap-m0/proof` → `deployed_by=airgap-m0-synthetic at=2026-08-06T00:18:25Z`
- `/nuon/airgap-m0/proof-run2` → `... at=2026-08-06T00:23:48Z`
- `/nuon/airgap-m0/proof-run3` → `... at=2026-08-06T00:31:28Z`
- `/nuon/airgap-m0/proof-run4` → `... at=2026-08-06T00:32:21Z`

## Bugs found during testing and fixed in this change

1. **`terraform show -json` overwrote backend state.** The first apply
   succeeded in AWS but was marked failed: `UpdateTerraformStateJSON`
   serialized the show document over the loopback backend's state file, so
   Terraform later read invalid state. Fix: `Store.PutTFStateShow` writes
   introspection data to `tfstate/<workspace>.json.show`; backend state stays
   in `tfstate/<workspace>.json` (mirrors ctl-api's separate column).
2. **Resume lost create-plan results.** `c.results` was memory-only, so a
   restarted process couldn't late-bind the prior create-plan into the apply
   step. Fix: `Store.ReadResult` reads `steps/<step>/result.json` and
   `chainedPlanContents` lazily reloads persisted results.
3. **Interrupted/failed steps were stuck on restart.** Fix: `NewClient`
   keeps `finished`/`available` and resets every other status to `available`,
   clearing execution metadata. Note: this auto-retries *failed* steps too —
   acceptable for M0, worth revisiting (`--retry-failed` flag?).
4. **Saved plans embed the loopback backend port.** A create plan referenced
   `127.0.0.1:52875`; after restart a fresh ephemeral port invalidated the
   saved plan. Fix: `NewTFBackend(store, portFile)` persists/reuses the port
   (`<state>/tfbackend-port`); `Close` explicitly closes the listener.
5. **`--log-level` was ignored** (always `zap.NewProduction()`). Fixed.

## Real install evaluation: `eks-simple` — NO-GO

Install `inloo2uqtyk0ion6jzaltujdgj` (app `eks-simple-auto-0804-053024`).
Latest successful sandbox pair:

- `jobxgksfkijm16ybv4ah8fy3s6` — sandbox-terraform create-apply-plan
- `job2zd9znuk4hwuts5gfh5nrxr` — sandbox-terraform apply-plan

Inspected `plan_json` for the create job (2.5 MB, from local ctl-api
Postgres). Two independent blockers:

1. **The install is a sandbox-mode simulation.** The plan carries a populated
   `sandbox_mode` block with obviously fake outputs: account `123456789012`,
   `vpc-0abc123def456`, cluster `nuon-cluster`, role names containing
   `fake-...`. The original "successful" run never provisioned real EKS
   infrastructure, so replaying it proves nothing about a real deployment.
2. **Variables are bound to a nonexistent account.** `vars.*_iam_role_arn`
   reference `arn:aws:iam::100000000025:...`, `vars.vpc_id` is the simulated
   `vpc-LIhsyILQhEyrBvACCEHRNkMTk`, and DNS vars point at
   `inloo2uqtyk0ion6jzaltujdgj.nuon.run`. `--force-default-cloud-auth`
   rewrites the auth block only — it does not rewrite Terraform variables, so
   a forced-real run against ambient account `676549690856` would fail (or
   worse, half-provision an EKS cluster against wrong inputs).

Decision: stopped before provisioning rather than run a mismatched EKS plan.
A meaningful real-install test needs an install created **without** sandbox
mode against a reachable account, then export via:

```bash
bins/nuonctl/scripts/export-airgap-plan \
  --install-id <real-install> \
  --job-ids <create-job>,<apply-job> \
  --force-default-cloud-auth \
  --out ./airgap-envelope.json
```

Never include destroy jobs. OCI-sync jobs also carry simulated ECR
coordinates and must be excluded until rewritten.

## Validation

```bash
gofmt -w pkg/runner/airgap bins/nuonctl/scripts/internal/exportairgapplan bins/runner/cmd
goimports -w pkg/runner/airgap bins/nuonctl/scripts/internal/exportairgapplan bins/runner/cmd
go test ./pkg/runner/airgap/... ./bins/nuonctl/scripts/internal/exportairgapplan/...
go build ./bins/runner ./bins/nuonctl/scripts/internal/exportairgapplan
```

All pass. Tests added during this session:

- statestore: `ReadResult` round-trip; show document does not overwrite
  backend state.
- client: resume keeps finished steps finished, resets failed steps and
  clears execution metadata; resumed client late-binds a persisted
  create-plan result into `apply_plan_contents`.
- tfbackend: persisted port survives restarts (`TestTFBackendPortPersistsAcrossRestarts`).
- exporter: `selectSteps` explicit ordering + dependency rebuild, compatible
  create/apply chaining, incompatible pairs don't chain, unknown/empty ID
  errors.

## Remaining risks / follow-ups

- Failed steps auto-retry on resume; may need an explicit opt-in.
- Backend port file: no handling for malformed contents or an occupied port;
  no guard against two concurrent runner invocations on one state dir.
- No true egress enforcement was tested; M1 should run inside a
  no-egress VPC/network namespace to prove isolation.
- `GetSandboxConfigs` returns nil — fine for the terraform path exercised
  here; other handlers may need it.
- Action workflows, OCI sync, and health checks are explicitly unsupported
  (`unsupported(...)`) in M0.

## Real non-sandbox install replay (2026-08-06)

A genuine, non-sandbox-mode install was provisioned and its real plan pair
replayed offline. All resources live in **sandbox-ht (504178855485,
us-east-1)** and are left running.

### Live install

| Entity | ID |
| --- | --- |
| org / app | `orgtv99pb5p8m91plbvm0vxtak` / `appq1zzqiv3atv7k5a9ek98gk2` |
| install | `inlqwvpl9h932fzsjpyqedzn93` |
| workflow (success) | `inwawgbpcuqytls2ko4s48fk1x` |
| CF stack | `nuon-eks-simple-auto-inlqwvpl9h932fzsjpyqedzn93` |
| VPC / EKS | `vpc-0285026e7cb6ea08a` / `w-inlqwvpl9h932fzsjpyqedzn93` (ACTIVE) |
| plan pair | `job2yg42zdzfqgpr2klg8h6b3s` (create-apply-plan), `jobzm7dih7odlkhpa2m72ixrzn` (apply-plan) |
| tf workspace / state | `tfwzuug0qby98vjquirupyj9vp` (93 resources) |
| stand-in runner SG | `sg-067c5a9598b7c2846` (tagged `network.nuon.co/domain=runner`) |

### Exporter fix: sandbox config was missing from the envelope

First real replay panicked (nil deref) at
`bins/runner/internal/jobs/sandbox/terraform/workspace.go:138` —
`appCfg.Sandbox.TerraformVersion` with `Sandbox == nil`. The exporter's
`getAppConfig` serialized only the `app_configs` row; the runner needs the
associated `app_sandbox_configs` row under a `sandbox` key.

Fix in `bins/nuonctl/scripts/internal/exportairgapplan/main.go`:
- `getAppConfig` now embeds the latest non-deleted `app_sandbox_configs` row
  as `sandbox` via `jsonb_build_object` (hstore columns cast to JSON objects
  natively).
- Guard added: exporter errors if the app config has no sandbox config,
  instead of letting the runner panic at replay time.

### Replay result

```
airgap run complete  succeeded=true
create-apply-plan: finished
apply-plan:        finished  (exit 0, ~2.5 min)
```

State serial 11 → 12, still 93 resources, nothing destroyed; EKS remains
ACTIVE. Replay was run three times (once after the fix, twice more for
egress observation) — resume/reset behavior held up each time.

Replay invocation:

```bash
env -u NUON_API_URL -u NUON_API_TOKEN \
  AWS_PROFILE=nuonctl-dev AWS_REGION=us-east-1 \
  ./runner airgap --plan envelope-real.json --state state-real
```

### Egress observation (lsof polling, no root)

Packet capture requires root (no passwordless sudo), so a 0.2 s `lsof -i`
poll of the runner process tree recorded every socket during a full clean
replay (`observe-egress.sh` + `run-observed.sh` in
`/Users/harsh/work/nuonco/airgap-m0-test/`). Raw list:
`egress-observed.txt` (175 unique sockets).

**Verdict: zero connections to the Nuon control plane.** No localhost
ctl-api ports (8081–8083), no Tailscale IPs (100.64.0.0/10), no
`ht.tail2117d3.ts.net`. The only loopback traffic is terraform ↔ the
runner's local HTTP state backend (`127.0.0.1:62250`).

| Destination | Attribution | Verdict |
| --- | --- | --- |
| `127.0.0.1:62250` | runner's local tfstate HTTP backend | expected (airgap design) |
| ec2 `compute-1` / us-east-2 IPs :443 | AWS APIs (STS/EC2/EKS/ELB/S3) | expected — real cloud provisioning |
| `108.157.238.x` (CloudFront) | registry.terraform.io | **gap**: provider downloads |
| `13.32.251.x` (CloudFront) | releases.hashicorp.com | **gap**: provider binaries |
| `20.207.73.82` :443 and :22 (github.com), `185.199.x` (GitHub CDN) | terraform module sources fetched during init (incl. git-over-SSH) | **gap**: modules not vendored |
| `104.18.30.120` (Cloudflare) | unattributed CDN hit during init | investigate in M1 |

The runner itself logged `no provider mirror in artifact, using direct
registry resolution` — the gaps above are exactly the artifact-vendoring
work (provider mirror + vendored module sources) already planned; they are
terraform-init egress, not control-plane egress.

### Stage cleanup notes (unchanged)

- Accidental CF attempt in stage (676549690856) rolled back and confirmed
  deleted.
- Quota increases accidentally requested in stage/us-east-1: VPCs→10
  (applied), IGWs→10 (applied), EIPs→15 (pending), NAT GW/AZ→10 (pending).
- `runner_group_settings.local_awsiam_role_arn` for
  `rgret6maal3zyl0lvdyabtxvdq` must stay blank.

## M1 end-to-end result (2026-08-06)

Full publish → download → verify → inspect/extract → offline replay chain
succeeded with the corrected exporter.

Reference IDs: org `orgtv99pb5p8m91plbvm0vxtak`, app
`appq1zzqiv3atv7k5a9ek98gk2`, app config `appd10fmjx6eeyzhcpzn2ut2q8`,
bundle `agb2cp8glw4w5xizfxmxey0jts`.

1. **Publish**: re-published after fixing two exporter bugs (below); bundle
   went `queued → publishing → active`, 374,601,369 bytes.
2. **Download**: via download-grant presigned URL; transport checksum
   matched (`sha256:2ac472de…`).
3. **Verify**: `nuon-bundle verify` — transport checksum, bundle manifest,
   and 22 blob digests all ok (8 artifacts, 357.2 MiB).
4. **Inspect/extract**: PLAN section shows all 11 steps (sandbox
   create/apply, 3× oci-sync, k8s manifest create/apply, terraform deploy
   create/apply, helm create/apply) with correct depends_on/plan_from
   chains; INPUTS section lists 4 specs. `--extract-plan` wrote
   `plan-envelope.json` mode `0600` with 3 component config connections,
   each carrying its nested config (`terraform_module`, `helm`,
   `kubernetes_manifest`).
5. **Replay**: `runner airgap` with `AWS_PROFILE=airgap-provision`
   (install provision role in sandbox 504178855485), ctl-api env scrubbed.
   All 11 steps `finished`, runner exited 0. Verified on cluster
   `w-inlqwvpl9h932fzsjpyqedzn93`: `whoami` deployment 1/1 (k8s manifest),
   ALB controller 2/2 (helm), cert-manager stack healthy; certificate TF
   module applied.

### Bugs found and fixed during M1

1. **Envelope app config missing component configs**: the exported
   `app_config` only had raw table columns + sandbox. Deploy handlers
   resolve `ComponentConfigConnections` with nested per-type configs, so
   `terraform-deploy create-apply-plan` failed with `unable to find
   terraform config`. Fixed in `plan_export.go`:
   `exportComponentConfigConnections` mirrors `GetFullAppConfig` preloads
   incl. the `component_config_connections_latest_configs_view` fallback.
   Regression tests: `TestExportComponentConfigConnectionsNestedConfigsAndLatestFallback`,
   `TestExportComponentConfigConnectionsMissingConfigFails`.
2. **Publish not idempotent**: re-publishing a bundle whose earlier attempt
   had written artifact rows failed with `duplicated key not allowed`
   (unique `(bundle_id, kind, logical_name)`), exhausting Temporal retries.
   Fixed in `publish_bundle.go`: delete the bundle's artifact rows in the
   same transaction before re-inserting.

### Observations for M2

- `apply-plan` steps (`jobyxje9…`, `jobzm7di…`) write `outputs.json` +
  `executions.json` but no `result.json`; create-plan/exec steps write all
  three. M2's post-trip collation should normalize this.
- Top-level `status.json` has `install_id`, `run_id`, `started_at`,
  `heartbeat_at`, `outputs`, `steps`; no run-level terminal status field
  yet.

## M2 results (2026-08-06) — post-trip report: PASSED

Scope per agreement: post-trip job results + outputs only. Secrets/inputs
and heartbeats deferred.

### What was built

1. **Run-level terminal status** (`pkg/runner/airgap/statestore/store.go`):
   `status.json` now carries `status` (`in-progress`/`finished`/`failed`),
   `failed_step`, `finished_at`. Initialized to `in-progress` on first run;
   resume of a failed/killed run resets it to `in-progress` and clears
   `failed_step`/`finished_at`.
2. **Post-trip `report.json`** (`pkg/runner/airgap/report.go`): on run
   completion (success or failure) the client collates status + per-step
   `result.json`/`outputs.json`/`executions.json` into a single
   `report.json` in the state root. Steps without `result.json` (apply-plan
   style) infer success from step status. Written via
   `Client.finalizeLocked` (`pkg/runner/airgap/client.go`).
3. **Store additions** (`statestore/disk.go`): `ReadOutputs`,
   `ReadExecutions`, `WriteReport`, `ReadReport` — same locking +
   atomic-write pattern as existing methods.
4. **Customer UX**: `nuon-bundle results --state <dir> [--json] [--out
   report.json]` prints a per-step summary table (status, success,
   executions, outputs, error) or the raw JSON, and can copy the report
   out (`bins/nuon-bundle/cmd/results.go`). If the run hasn't finished it
   says so instead of failing cryptically.

### Validation

- Unit tests: `TestSuccessfulRunFinalizesStatusAndReport`,
  `TestFailedRunFinalizesStatusAndReport`,
  `TestResumeAfterFailureResetsRunStatus` (`report_test.go`); all
  `pkg/runner/airgap/...` + `bins/nuon-bundle/...` green.
- Real-state validation: copied the 11-step M1 replay state to
  `airgap-m0-test/state-real-m2`, re-ran `runner airgap` (resume path — no
  jobs re-executed, no AWS mutation). Client finalized:
  `status=finished`, wrote `report.json`. `nuon-bundle results` showed all
  11 steps `finished`/`success=true` with real outputs collated (sandbox
  `cluster`/`ecr`/`account`, oci-sync `image`, tf `public_domain_certificate_arn`,
  helm `deployments`/`ingresses`/`manifest`, k8s `diff`).
- Steps recording literal `null` outputs are reported as having no outputs.

## M3: real EC2 air-gap-shaped run (2026-08-06)

Full end-to-end on EC2 `i-02e27e986ce44e0a0` (sandbox `504178855485`,
`us-east-1`): runner binary extracted from the published bundle, executed
under `env -i` with no ctl-api URL/token, results delivered via customer
S3 only.

### Bundle republish

- Bundle `agb2cp8glw4w5xizfxmxey0jts` republished after the nil-outputs
  resume fix; publish completed in ~14 min (previous attempts died on a
  30-min activity timeout — the annotation is now 180m — and then on
  expired local SSO creds; see workarounds below).
- New artifact: manifest `sha256:0b73557332ca…`, transport checksum
  `8db1fb64b5fa1cfc2d1f7a6e5cfee10a3283ba96e671e9e57844b8bd026dc89c`,
  1376041684 bytes. `nuon-bundle verify`: transport checksum / manifest /
  47 blob digests all ok.
- Extracted runner SHA-256 matches the fixed binary:
  `b9ff132e68d01e619d207f2a84c7e12897f366fd47f38d8aa437d0e5fbb77616`.
- Bundle moved into the customer bucket in-AWS (presigned grant URL
  curled on the instance via SSM, checksum re-verified on-instance, then
  `aws s3 cp` to `s3://nuon-airgap-m3-504178855485/bundle/`) — no 1.4 GiB
  upload from the workstation.

### Run evidence

- Fresh bootstrap (state wiped, seed tfstate kept): download → verify →
  extract plan+runner from bundle → seed TF state → run.
- First pass failed at step 4 (`jobcymzvjhtcvurvrmrerxufs6`, k8s
  manifest): kubeconfig exec plugin `aws-iam-authenticator` not present
  on the host. Installed v0.7.4 and re-ran.
- Resume picked up from `status.json` (steps 1–3 stayed `finished`) and
  completed the remaining 8 steps — this exercised the nil-outputs resume
  fix on a real run.
- Final: `status=finished`, 11/11 steps finished, runner exit code 0,
  started 12:10:22Z finished 12:16:15Z (~6 min). `report.json` collated
  outputs per step (7/7/1/0/1/1/1/1/1/1/5). `state/DONE`, `runner.log`,
  and per-step `steps/**` artifacts all uploaded to the state bucket.
- runner.log shows both terminals: `succeeded:false failed_step:jobcymz…`
  then `succeeded:true`.

### Gaps found (for the no-egress milestone)

1. `aws-iam-authenticator` is a hidden host dependency: k8s-manifest jobs
   build a kubeconfig that shells out to it. Bundle should package it, or
   the runner should mint EKS tokens in-process.
2. CloudWatch agent stopped shipping logs after the log file was
   deleted/recreated (inode change). Moot once the OTEL pipeline replaces
   the CW agent, but worth knowing for reruns.
3. Bootstrap still uses egress: dnf, S3, CW, SSM, GitHub (authenticator
   download). VPC endpoints + bundled tooling needed before removing the
   NAT route.
4. Packaged runner/workload images still unproven against the local
   registry path (`bins/runner/cmd/airgap.go` BundleDir/RegistryDir);
   plan `oci-sync` steps may still hit external ECR.

### Credential workarounds used (local, reverted)

Local SSO expired mid-publish. Bridged without interactive login by
writing the still-valid cached `stage.NuonAdmin` role creds as static
keys into `~/.aws/credentials` (Go SDK prefers static over SSO within a
profile) and patching the cached SSO token expiry so the nuonctl
supervisor would restart services without popping login windows. Both
reverted after a real `aws sso login`.

## M4–M8: fresh customer stack replay + plan late-binding (2026-08-07) — PASSED

Replayed `bundle-v4.oci.tar.zst` against a **fresh customer
CloudFormation stack** in sandbox `504178855485` / `us-east-1` — i.e. an
environment whose VPC, subnets, IAM roles, and (once the sandbox
applied) Route53 zones all differ from the vendor reference install the
plans were rendered against.

- Stack: `nuon-eks-simple-auto-inlqwvpl9h932fzsjpyqedzn93`, actual VPC
  `vpc-09ed1a146b73f8343` (reference plans baked `vpc-0285026e7cb6ea08a`,
  since deleted).
- Stack outputs were produced by the **connected** phone-home path and
  saved to `airgap-m0-test/install-stack-outputs.json`; the replay itself
  ran `runner airgap` with `--install-stack-outputs` and no ctl-api
  URL/token. (S3-rendezvous bootstrap is implemented but not yet
  exercised — see gaps.)
- Workdir: `airgap-m0-test/m4-workdir`, runs m4→m8 resumed the same
  state; no AWS resources recreated between attempts.

### Root cause established: exported plans bake reference-install values

Two distinct classes of stale rendered values, both unfixable by
republishing (the customer environment is unknowable at export time):

1. **Install stack outputs** — the CloudFormation phone-home values of
   the vendor's reference install (VPC ID, subnets, IAM role ARNs) are
   interpolated into sandbox terraform vars and `install_stack.outputs`
   state snapshots.
2. **Sandbox outputs** — later deploy plans bake the reference sandbox's
   terraform outputs (Route53 `Z047625027FRB0M46S2JU`, domains, ECR
   URLs) into `deploy_plan.terraform.vars` and the
   `state.sandbox.outputs` / `state.install.sandbox.outputs` snapshots.
   The actual sandbox produced `Z09864612C5I36BZSUNN9` (public) /
   `Z00538421D4Z64RDOR78Q` (internal), so the terraform-deploy apply
   failed with `couldn't find resource` on the reference zone.

### Fixes (all in `pkg/runner/airgap/latebind.go` + client)

- **Install-stack-output rebinding**: `install_stack.outputs` snapshots
  are replaced with the target stack's outputs, and every rendered value
  elsewhere in the plan is substituted on exact match (whole string or
  comma-separated element).
- **Sandbox-output late binding** (new): after the sandbox apply
  finishes, its recorded outputs are structurally aligned against the
  plan's `sandbox.outputs` snapshots; changed string leaves (min length
  6 to avoid generic-value collisions) become old→new substitutions
  applied plan-wide, and the snapshots are replaced with the local
  outputs. Source of truth: most recent finished `sandbox`-group step's
  outputs in `status.json` (works across resumes).
- **Resume re-plans chained plan steps**: resuming a failed apply also
  resets its `plan_from_step` create-plan step, fixing
  `Error: Saved plan is stale` (`TestClientResumeReplansChainedPlanStep`).
- **Per-job NDJSON logs**: `JobLogDir` config; runner writes
  `<state>/job-logs/<job-id>.ndjson`, which is what surfaced both
  failures.

Tests: `TestRenderStepPlanRebindsInstallStackOutputs`,
`TestRenderStepPlanRebindsSandboxOutputs`,
`TestRenderStepPlanRebindsSandboxOutputsAfterResume` — full targeted
suite green (`pkg/runner/airgap/... bins/runner/... bins/nuon-bundle/...
services/ctl-api/internal/pkg/stacks/...`).

### m8 run evidence — 11/11 steps finished

`runner-m8` resumed the workdir; the failed terraform-deploy pair
re-planned and applied cleanly, then the remaining oci-sync and helm
steps completed. `status=finished`, no failed step, `report.json`
collated outputs for all 11 steps.

Verified live (no teardown):

- ACM cert `9263e680-…` for `*.inlqwvpl9h932fzsjpyqedzn93.nuon.run` with
  its DNS-validation CNAME created **in the fresh public zone
  `Z09864612C5I36BZSUNN9`** — direct proof the sandbox-output rebind
  redirected the terraform apply. (Cert `PENDING_VALIDATION`: the
  `nuon.run` parent NS delegation to the customer zone doesn't exist in
  an air-gapped account — expected.)
- EKS `w-inlqwvpl9h932fzsjpyqedzn93`: whoami deployment 1/1 Available +
  service (k8s-manifest step), ALB ingress
  `inlqwvpl9h932fzsjpyqedzn93-public` with the deploy-created cert ARN
  (helm step), all sandbox platform charts deployed
  (alb-ingress-controller, cert-manager, external-dns, ingress-nginx,
  kyverno, metrics-server).
- oci-sync pushed images to in-account ECR
  `504178855485.dkr.ecr.us-east-1.amazonaws.com/inlqwvpl9h932fzsjpyqedzn93`.
- No egress observer ran during m5–m8; the only egress evidence on file
  is `egress-observed-m4.txt`. Do not claim no-egress for m8.

### Remaining gaps (unchanged priorities)

1. **S3 rendezvous not yet end-to-end**: `phonehome.py` S3 mode,
   CloudFormation params/IAM, `runner airgap` S3 polling, and
   `nuon-bundle stack prepare` wiring are all implemented and unit
   tested, but ctl-api still serves the phone-home script from the
   GitHub `aws-v0.1.4` tag (`get_phonehome_script.go`) — the S3-mode
   script must be published/tagged in `nuonco/runner` (or embedded)
   before a real no-ctl-plane bootstrap run.
2. Root customer stack template availability inside the exported bundle.
3. Delete semantics for the S3 outputs object; bucket/key must exist
   before stack creation; runner role needs `s3:GetObject`, Lambda role
   scoped `s3:PutObject`.
4. M3 gaps still open: `aws-iam-authenticator` host dependency, local
   registry path for oci-sync unproven against packaged images.
