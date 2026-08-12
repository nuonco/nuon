# 01 — `pkg/bundle`

> Owns: `pkg/bundle`. Depends on: nothing. Blocks: 02, 04, 05.
> Contracts: owns **C1**, owns **C3**.

## Goal

One well-tested library that can **build, sign, upload, download, and verify** a
bundle, usable identically from ctl-api, the runner, and the customer-run
installer.

## Why this is first

02, 04, and 05 all need it, and it has no dependencies of its own. It is also the
only project with a substantial existing implementation to harvest, so it should
be the fastest to a working state.

## Hard constraint

**`pkg/bundle` imports nothing from `github.com/nuonco/nuon`.**

Enforce in CI:

```bash
go list -deps ./pkg/bundle/... | grep nuonco && exit 1 || exit 0
```

This is achievable — the reference implementation already has zero internal
imports. It is what makes the package usable from a customer-facing binary without
dragging in gorm, fx, and `internal.Config`. Guard it, because the pressure to
import `pkg/config` for "first-class workflow support" will be real (see C3).

## Harvest

The reference implementation is on commit `982ff3ced`, **not** on this branch.

```bash
git show 982ff3ced:pkg/runner/oci/bundle/bundle.go       # 586 lines, write side
git show 982ff3ced:pkg/runner/oci/bundle/read.go         # 406 lines, read side
git show 982ff3ced:pkg/runner/oci/bundle/bundle_test.go  # 369 lines, 17 tests
git show 982ff3ced:pkg/runner/oci/bundle/read_test.go    # 185 lines, 7 tests
git show 982ff3ced:pkg/runner/oci/bundle/runner_test.go  # 113 lines, 5 tests
```

Dependencies — all five already in `go.mod`, no new modules needed:
`github.com/klauspost/compress/zstd` (promote from `// indirect`),
`github.com/opencontainers/go-digest`,
`github.com/opencontainers/image-spec/specs-go{,/v1}`,
`oras.land/oras-go/v2{,/content/oci}`.

**Zero importers on this tree**, so the move is free.

### Separately: archive hardening

That commit also improved `pkg/runner/oci/archive` — a new `tar.go`
(`writeTarLayer`, `validateTarPath`), a 195-line `archive_test.go`, and deltas to
`pack.go`/`unpack.go`/`archive.go`. Harvest those too, but land them **in place**.
`pkg/runner/oci/archive` has **34 importers** across `bins/runner/internal/jobs/*`
and `pkg/runner/jobs/build/*` and must not move.

## Keep as-is

This is the hard, proven part. Port it with its tests and resist rewriting:

- Digest-verified graph traversal — `collect`, `successors`, with the
  `traversalKey{digest, mediaType}` dedup.
- The deterministic tar+zstd envelope — sorted names, mode `0644`, zeroed
  times/uid/gid, PAX format, single zstd frame. Two builds of a logically-equal
  manifest produce byte-identical output.
- Bidirectional manifest↔roots validation: missing member, mismatched descriptor,
  and undeclared root are all errors.
- Extraction hardening — zip-slip, non-canonical paths, backslashes, duplicates,
  non-regular entries, `MaxEntries`/`MaxFileBytes`/`MaxTotalBytes`, zstd
  memory/window caps, must-be-empty destination, wipe-on-failure.
- `VerifyBlobs` — re-hash every reachable blob from the bundle descriptor and all
  roots.

## Redesign

### 1. Member-kind registry (replaces the flat struct)

`LogicalManifest` is a closed struct and `validateMembers` is a ~150-line function
that grows linearly with every new member kind. Since we need
`sandbox | component | action | workflow | runbook | image | stack_asset | binary`
as first-class, replace it:

```go
type LogicalManifest struct {
    SchemaVersion int      `json:"schema_version"` // 1
    Target        Target   `json:"target"`
    Members       []Member `json:"members"`
}
```

Members carry their `Kind` (C1). Validation becomes a per-kind rule table:
required fields, whether the kind permits nested steps, whether platform metadata
is required. Adding a kind becomes adding a table row plus a test.

Canonicalization sorts members by logical key. This is what makes
`ManifestDigest` stable and therefore a sound signing target.

### 2. Keyed documents (replaces three hardcoded slots)

```go
type DocumentKind string
const (
    DocProvenance DocumentKind = "provenance"
    DocWorkflows  DocumentKind = "workflows"
    DocSignature  DocumentKind = "signature"
)

func (b *builder) Document(kind DocumentKind, v any) error
func (b *Bundle) Document(kind DocumentKind, out any) error
```

Kind → media type mapping lives in one table. Adding a document kind touches one
place instead of four.

### 3. Drop the airgap-specific `Runner` member

`Runner{Version, SourceURL, Binary, Image}` and `ExtractRunnerBinary` hardcode
`RunnerBinaryArtifactType`. A generic bundle library should not know what a runner
is. Re-express as `KindBinary` with a name, and let the caller ask for
`binary:runner`. Replace `ExtractRunnerBinary` with the generic
`(*Bundle).Blob(ctx, digest, w)` — which must keep the digest-verifying reader the
old method had.

### 4. Collapse the constructor telescope

`Generate` / `GenerateWithDocuments` / `GenerateWithOptions` → `NewBuilder(target)`
plus functional options (C1).

### 5. Rename media types

`application/vnd.nuon.airgap.*` → `application/vnd.nuon.bundle.*`, all carrying
`.v1`. This is a wire-format break and must land before anything ships.

## New

### Signing

```go
func Sign(manifestDigest string, key ed25519.PrivateKey) (Signature, error)
func Verify(manifestDigest string, sig Signature, pub ed25519.PublicKey) error
```

ed25519 over the canonical manifest digest. Integrity of everything else chains
from that digest through the OCI descriptor graph, which `VerifyBlobs` walks — so
signing one 32-byte value covers the whole bundle, provided callers run both.

Provide a single helper that does the full check in the right order, so callers
can't accidentally do half of it:

```go
func VerifyBundle(ctx context.Context, dir string, expected Expectation) error
// Expectation{TransportSHA256, ManifestDigest, PublicKey} — any zero field is
// SKIPPED and the skip is reported in the returned error's detail, never silently.
```

### Transport

The `Transport` interface from C1. Start from the reference
`services/ctl-api/internal/app/airgap/transport.Store` shape on `982ff3ced`
(`Configured`/`Publish`/`Grant` with a provider-tagged `Replica`) — the
best-shaped storage interface in the repo — and generalize to get/put/list.

Implementations: `transport/s3` first. `gcs` and `azblob` are stubs that return a
clear "not implemented" error. Note there is **no `azblob` dependency in `go.mod`
today** — adding Azure is a new module.

## Security requirements

Each of these is a defect the air-gap prototype actually shipped. They are
acceptance criteria, not suggestions.

| # | Requirement | What went wrong before |
| --- | --- | --- |
| S1 | `Verify` fails closed; no "ok" without an actual comparison | `nuon-bundle verify` printed `transport checksum: ok` unconditionally, with no flag to supply an expected value |
| S2 | Skipped checks are reported, never silent | Callers could not tell which of the three checks ran |
| S3 | `PutOptions.ServerSideEncrypt` defaults true | Zero SSE on any air-gap write, while terraform state landed in the same bucket |
| S4 | `Put` re-reads the written version and constant-time compares | (this one the prototype got right — keep it) |
| S5 | No `README` or doc comment may claim "signed" until `Sign` is wired end to end | The prototype README claimed "a single, signed, checksummed archive"; nothing was signed |

## Files

```
pkg/bundle/
  bundle.go        LogicalManifest, Member, Artifact, Target, Source
  builder.go       Builder, NewBuilder, options
  kinds.go         MemberKind + the per-kind validation table
  document.go      DocumentKind, media-type mapping
  read.go          Extract, Open, Bundle, Members, Blob
  verify.go        VerifyBlobs, VerifyBundle, Expectation
  sign.go          Sign, Verify, Signature
  archive.go       tar+zstd envelope (harvested writeArchive/extract)
  transport/
    transport.go   Transport, ObjectRef, Grant, PutOptions
    s3/s3.go
    gcs/gcs.go     stub
    azblob/…       stub
```

## Milestones

1. **Harvest + rename.** Files moved to `pkg/bundle`, package renamed, media types
   renamed, all 29 tests ported and green. No API changes yet. Verifies the
   zero-internal-imports claim.
2. **Member-kind registry + keyed documents.** The two structural redesigns, with
   a round-trip test per kind.
3. **Signing.** `Sign`/`Verify`/`VerifyBundle` + `Expectation`, with tamper tests.
4. **Transport.** Interface + S3 implementation + read-back verification.
5. **Archive hardening** harvested in place under `pkg/runner/oci/archive`.

Milestone 1 alone unblocks 02 and 04 for compile-against purposes, so land it
early and announce the API.

## Tests

- All 29 harvested tests. The valuable ones: determinism
  (`TestGenerateDeterministicLayoutAndCanonicalManifest`), closure verification,
  corrupt/missing blob rejection, limits enforced *before* fetch, read bounded by
  declared size, same-digest-different-media-type traversal, malformed descriptor
  no-panic, manifest identity changing with packaged root and with platform.
- Read-side: round-trip + copy, missing reachable blob, metadata bound, conflicting
  size, zip-slip parent path, non-canonical and duplicate paths, unsupported entry
  type, non-empty destination.
- **New**: one build+open+verify round trip per `MemberKind`; a bundle carrying all
  eight kinds at once; signature valid / wrong key / tampered manifest / tampered
  blob; `VerifyBundle` with each `Expectation` field zeroed reports the skip;
  golden-file determinism (commit a fixture digest and assert it).
- CI check: `go list -deps ./pkg/bundle/... | grep nuonco` is empty.

## Out of scope

Plan envelopes, late binding, install semantics, deployment-ID rewriting, and
anything that requires knowing what a Nuon install is. If a member needs
app-specific semantics, it goes in a C3 document blob that this package does not
parse.

## Risks

- **C3 pressure.** "Make runbooks first-class" will be read as "import
  `pkg/config`". It must not be. The manifest carries a runbook *member* (name,
  config digest, artifact); the runbook's *semantics* live in the workflows
  document that 02 writes and 05 reads.
- **Rewrite temptation.** The traversal and extraction code is unglamorous and
  correct. Port it, don't reimplement it. Its test suite is the asset.
- **Rename timing.** If the media-type rename slips past 02 shipping, it becomes a
  migration instead of an edit.
