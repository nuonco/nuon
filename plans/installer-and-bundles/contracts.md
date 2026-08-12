# Contracts (normative)

These five interfaces let specs 01–05 be built in parallel. **A project may not
change a contract unilaterally** — changing one requires updating this file and
telling every consumer.

Status: `DRAFT` until wave 0 sign-off, then `FROZEN`.

---

## C0 — The three execution axes

**Do not conflate these.** The most likely way this work goes wrong is overloading
an existing enum. Two axes already exist; we add one.

| Axis | Field | Values | Change |
| --- | --- | --- | --- |
| Who **drives the workflow** | `Workflow.ExecutionType` — **new** | `control-plane` (default) \| `external` | add field |
| Who **claims the runner job** | `RunnerJob.Executor` — `internal/app/runner_job.go:233-238` | `org-runner`, `control-plane`, `""` | add `external` |
| What **kind of step** it is (UI + semantics) | `WorkflowStep.ExecutionType` — `internal/app/workflow_step.go:19-27` | `system`, `user`, `approval`, `skipped`, `hidden` | **unchanged** |

Rationale for three separate axes: a bundle install's workflow is `external`, but
it still contains `approval` steps (customer approves a plan) and `system` steps
(control plane bookkeeping). A single enum cannot express that.

`RunnerJob.Executor` already has precedent for a non-runner executor:
`control-plane` exists so orgs with no runner can build, with a listing endpoint
`GET /v1/runner-jobs?executor=control-plane`
(`internal/app/runners/service/list_runner_jobs_ctl_api.go:36-64`). The installer's
job claim is that same shape with a different value — follow it rather than
inventing a new claim mechanism.

`WorkflowStep.ExecutionType == user` is currently used by exactly one thing: the
await-install-stack step (`internal/app/installs/workflows/v2/shared.go:139-141`).
That is the existing "a human or external system completes this out of band"
primitive. Spec 03 models on it.

### Defaults and backfill

`Workflow.ExecutionType` defaults to `control-plane`. Existing rows backfill to
`control-plane`. `RunnerJob.Executor` keeps its existing default behaviour
(`internal/app/runner_job.go:372-373`).

---

## C1 — `pkg/bundle` public API

Owned by **01**. Consumed by **02** (build) and **04**/**05** (read, verify).

Hard constraint: **`pkg/bundle` imports nothing from `github.com/nuonco/nuon`.**
Enforce in CI with `go list -deps ./pkg/bundle | grep nuonco` returning empty. This
is what keeps it usable from the installer, the runner, and ctl-api alike, and it
is achievable — the reference implementation already has zero internal imports.

### Member kinds

```go
type MemberKind string

const (
    KindComponent  MemberKind = "component"
    KindSandbox    MemberKind = "sandbox"
    KindImage      MemberKind = "image"
    KindAction     MemberKind = "action"
    KindWorkflow   MemberKind = "workflow"
    KindRunbook    MemberKind = "runbook"
    KindStackAsset MemberKind = "stack_asset"
    KindBinary     MemberKind = "binary"
)
```

Each member has a **logical key**, unique within a bundle:
`<kind>:<name>` (`component:api`, `runbook:rotate-certs`), and for nested content
`<kind>:<name>/<subkind>:<subname>` (`action:migrate/step:up`).

### Types

```go
type Member struct {
    Kind         MemberKind
    Name         string
    ConfigDigest string    // sha256 of the canonical config that produced this
    Source       *Source   // provenance only, never trusted for verification
    Artifact     Artifact
}

type Artifact struct {
    MediaType            string
    Digest               string // sha256:...
    Size                 int64
    PlatformOS           string
    PlatformArchitecture string
}

type Root struct {
    Descriptor ocispec.Descriptor
    Source     oras.ReadOnlyTarget
}
```

### Build

```go
type Builder interface {
    Add(m Member, r Root) error
    Document(kind DocumentKind, v any) error
    Build(ctx context.Context, dst io.Writer, opts ...Option) (Result, error)
}

func NewBuilder(target Target) Builder   // Target{OS, Architecture}

type Result struct {
    ManifestDigest  string // the signing target
    BundleDigest    string
    TransportSHA256 string // sha256 of the compressed archive
    Size            int64
}

// Options
func WithMaxContentBytes(int64) Option
func WithMaxBlobBytes(int64) Option
func WithOnBlobVerified(func(ocispec.Descriptor)) Option
```

### Read

```go
func Extract(dst string, r io.Reader) (transportSHA256 string, err error)
func ExtractWithOptions(dst string, r io.Reader, o ExtractOptions) (string, error)
func Open(ctx context.Context, dir string) (*Bundle, error)
func VerifyBlobs(dir string) error

func (b *Bundle) Members() []Member
func (b *Bundle) Member(key string) (Member, bool)
func (b *Bundle) Document(kind DocumentKind, out any) error
func (b *Bundle) ManifestDigest() string
func (b *Bundle) Store() oras.ReadOnlyTarget
func (b *Bundle) Blob(ctx context.Context, digest string, w io.Writer) error
```

### Sign / verify

```go
type Signature struct {
    Algorithm string `json:"algorithm"` // "ed25519"
    PublicKey string `json:"public_key"`// base64, for key identification
    Digest    string `json:"digest"`    // the manifest digest that was signed
    Value     string `json:"value"`     // base64 signature
    SignedAt  time.Time `json:"signed_at"`
}

func Sign(manifestDigest string, key ed25519.PrivateKey) (Signature, error)
func Verify(manifestDigest string, sig Signature, pub ed25519.PublicKey) error
```

The manifest is canonicalized before hashing (members sorted by logical key), so
the manifest digest is a stable signing target. **Signing covers the manifest
digest only** — integrity of everything else chains from it through the OCI
descriptor graph, which `VerifyBlobs` walks.

### Transport

```go
type Transport interface {
    Configured() bool
    Put(ctx context.Context, key string, r io.Reader, size int64, opts PutOptions) (ObjectRef, error)
    Get(ctx context.Context, ref ObjectRef, w io.Writer) error
    List(ctx context.Context, prefix string) ([]ObjectRef, error)
    Grant(ctx context.Context, ref ObjectRef, filename string, ttl time.Duration) (Grant, error)
}

type ObjectRef struct {
    Provider string // "aws_s3" | "gcs" | "azure_blob"
    Region   string
    Bucket   string
    Key      string
    Version  string
    Checksum string
    Size     int64
}

type PutOptions struct {
    ContentType      string
    ServerSideEncrypt bool   // MUST default true — see C1 security notes
    ChecksumSHA256   string
}

type Grant struct {
    URL           string
    ExpiresAt     time.Time
    SupportsRange bool
}
```

### C1 security requirements

Non-negotiable, and each one is a defect the air-gap prototype actually shipped:

1. **`Verify` fails closed.** No "ok" is printed or returned unless a comparison
   actually happened against a caller-supplied expected value.
2. **`PutOptions.ServerSideEncrypt` defaults true.** Terraform state and plan
   contents land in these buckets.
3. **Publish verifies read-back.** After `Put`, re-read the exact version and
   constant-time compare digest and size before returning success.
4. **Extraction stays hardened.** Keep every existing guard: zip-slip,
   non-canonical and duplicate paths, non-regular entries, entry/file/total caps,
   zstd window cap, must-be-empty destination.

---

## C2 — installer-api ↔ ctl-api HTTP contract

Server owned by **03**. Client owned by **04**/**05**. Generated into
`sdks/nuon-go`.

All routes are authenticated with the scoped token from C4 and are scoped to a
single install. The server must reject any request whose token install does not
match the path install — this is stricter than today's runner routes, which take
`:runner_id` from the path without cross-checking the token.

```
# discovery
GET    /v1/installer/install                         -> install metadata, management_mode, bucket config
GET    /v1/installer/bundles                         -> [{id, status, manifest_digest, transport_checksum, size, created_at}]
GET    /v1/installer/bundles/{bundle_id}             -> bundle detail + this install's signature + public key
POST   /v1/installer/bundles/{bundle_id}/download-grants -> {url, expires_at, filename, size, transport_checksum, manifest_digest, supports_range}
PUT    /v1/installer/bundles/{bundle_id}/status      <- {status, verified_at?, message?}

# work
GET    /v1/installer/jobs?status=available           -> claimable external runner jobs
POST   /v1/installer/jobs/{job_id}/claim             -> claim (idempotent per agent)
GET    /v1/installer/jobs/{job_id}/plan               -> composite plan
POST   /v1/installer/jobs/{job_id}/executions        -> create execution
POST   /v1/installer/jobs/{job_id}/executions/{exec_id}/result  <- results/outputs
PATCH  /v1/installer/jobs/{job_id}                   <- status

# workflow status
GET    /v1/installer/workflows                       -> external workflows for this install
GET    /v1/installer/workflows/{workflow_id}/steps   -> steps
POST   /v1/installer/workflows/{workflow_id}/steps/{step_id}/complete  <- external step completion
PATCH  /v1/installer/workflows/{workflow_id}/steps/{step_id}/status    <- progress

# approvals
GET    /v1/installer/approvals                       -> pending approvals + contents refs
GET    /v1/installer/approvals/{approval_id}/contents
POST   /v1/installer/approvals/{approval_id}/response <- {approved: bool, reason?}

# telemetry
POST   /v1/installer/component-health                <- health observations
POST   /v1/installer/heartbeats                      <- agent liveness
```

Reuse existing request/response shapes wherever possible — in particular
`CreateRunnerJobExecutionResult`
(`internal/app/runners/service/create_runner_job_execution_result.go:83`) already
handles `Contents`, `ContentsDisplay`, the gzip variants, `ErrorMetadata`, and
`ErrorCode`. Do not invent a second result shape.

### Idempotency

Every write is idempotent on a client-supplied key. Job claims are idempotent per
agent. Step completion is idempotent on `(step_id, attempt)`. The installer may be
killed and restarted mid-flow at any point and must not double-apply.

---

## C3 — Bundle manifest and document schema

Owned by **01**. Filled by **02**. Read by **04**/**05**.

Two layers. **The split is the contract** — it is the one clearly good idea in the
reference implementation and it is what keeps `pkg/bundle` free of a `pkg/config`
dependency.

### Layer 1 — OCI logical manifest: content addressing only

Media type `application/vnd.nuon.bundle.manifest.v1+json`.

Per member: `{kind, name, config_digest, source?, artifact{media_type, digest,
size, platform_os?, platform_architecture?}}`. Nothing app-semantic. Verifiable by
a tool that knows nothing about Nuon.

### Layer 2 — typed document blobs: execution semantics

Each is an opaque `json.RawMessage` under its own media type. `pkg/bundle` stores
and digests them; **it never parses them.**

| `DocumentKind` | Media type | Owner | Contents |
| --- | --- | --- | --- |
| `provenance` | `…bundle.provenance.v1+json` | 02 | app config ID, branch run ID, build IDs keyed by logical key |
| `workflows` | `…bundle.workflows.v1+json` | 02 | the workflows manifest — see below |
| `signature` | `…bundle.signature.v1+json` | 02 | `Signature` from C1 |

### Workflows manifest

The executable surface. Shape owned by 02, consumed by 05.

```json
{
  "schema_version": 1,
  "org_id": "...", "app_id": "...", "app_config_id": "...",
  "app_branch_run_id": "...",
  "workflows": [
    {
      "id": "...", "name": "...", "type": "provision|deploy_components|runbook_run|action_workflow_run",
      "execution_type": "external",
      "steps": [
        { "id": "...", "name": "...", "kind": "job|approval|health-gate",
          "job_type": "...", "job_group": "...", "job_operation": "...",
          "depends_on": ["..."], "plan_from_step": "...",
          "member_key": "component:api",
          "composite_plan": { }
        }
      ]
    }
  ],
  "inputs": [ { "name": "...", "type": "...", "required": true, "secret": false, "default": null } ]
}
```

**Media type versioning**: every media type carries `.v1`. A breaking change to
any document shape means `.v2` and a reader that accepts both. Do not mutate a
`.v1` shape in place.

### Naming

All media types are `application/vnd.nuon.bundle.*`. The reference implementation
used `vnd.nuon.airgap.*`; that rename happens in 01 before anything ships, because
it is a wire-format break.

---

## C4 — Install connect handshake

Owned by **03**. Consumed by **04**.

```
installer connect --install-id <id> --install-token <token>
  -> POST /v1/installer-auth/connect
     body: {install_id, install_token, agent_fingerprint, version}
     200:  {token, expires_at, org_id, app_id, install_id, management_mode, bucket_config}
```

The install token is a long-lived, revocable, install-scoped credential the vendor
generates and hands to the customer alongside the binary. The returned `token` is
short-lived and scoped to the C2 route set only.

Build on the existing precedent rather than a new mechanism:
`POST /v1/installs/{install_id}/runner-bootstrap-token`
(`internal/app/installs/service/create_runner_bootstrap_token.go:34-68`) already
mints an install-scoped credential for a customer-side process, backed by
`internal/app/runners/helpers/token.go` (which mints an `app.Token` against the
service account `<id>@serviceaccount.nuon.co`).

### Requirements

- Install tokens are **revocable** and revocation takes effect on the next token
  refresh. Model on `InstallStackVersion.PhoneHomeTokenRevokedAt`.
- The scoped token is **refreshed**, not renewed indefinitely; the installer
  re-presents the install token.
- Connecting records or updates an `InstallerAgent` row and sets `status =
  connected`.
- `bucket_config` is returned but **the control plane never holds customer cloud
  credentials** — the installer provisions and accesses the bucket with the
  operator's or instance's own credentials.
