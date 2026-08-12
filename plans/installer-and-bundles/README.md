# Installer & Bundles

Implementation specs for the RFC at [`../../installer-and-bundles.md`](../../installer-and-bundles.md).
Flow diagrams: [`../installer-and-bundles.html`](../installer-and-bundles.html).

The RFC introduces a customer-run **installer** — a CLI that embeds a webapp and
an installer-api — which replaces the "stack" semantic. Bundles are 1:1 with an
app-branch-run. Bundle installs are executed and approved by the customer, with a
limited set of status pushed back to the control plane.

## Specs

| Spec | Owns | Depends on |
| --- | --- | --- |
| [`01-pkg-bundle.spec.md`](01-pkg-bundle.spec.md) | `pkg/bundle` — build, sign, upload, verify | — |
| [`02-app-config-bundles.spec.md`](02-app-config-bundles.spec.md) | `AppConfigBundle`, app config `bundle_enabled`, branch-run publish step, Bundles UI | C1 |
| [`03-bundle-installs.spec.md`](03-bundle-installs.spec.md) | Bundle installs, `external` execution, status push, install UI | — |
| [`04-installer-cli.spec.md`](04-installer-cli.spec.md) | `bins/installer` scaffolding, setup/connect, installer-api | C1, C2, C4 |
| [`05-installer-install-flow.spec.md`](05-installer-install-flow.spec.md) | The install + update flow driven by the installer | 04, C1, C2 |

[`contracts.md`](contracts.md) is **normative**. It defines the five interfaces
(C0–C4) that let these run in parallel. No spec may change a contract
unilaterally.

## Dependency graph

```
      ┌───────────────────────────────────┐
      │ 01  pkg/bundle       (library)    │  start now
      └────────────────┬──────────────────┘
                       │ C1
        ┌──────────────┴──────────────┐
        ▼                             ▼
┌──────────────────────┐   ┌──────────────────────┐
│ 02 app-config-bundles│   │ 04 installer-cli     │
│    (publish side)    │   │    (customer side)   │
└──────────────────────┘   └──────────┬───────────┘
                                      │             ┌──────────────────────┐
                                      ├────────────▶│ 05 install flow      │
                                      │             └──────────────────────┘
                            C2, C4    │
                           ┌──────────┴───────────┐
                           │ 03 bundle-installs   │  start now
                           └──────────────────────┘
```

01 and 03 have no dependencies and start immediately. 02 can build its model,
config plumbing, branch-run step, and UI against a C1 stub. 04 builds against the
written-down C1/C2/C4 without waiting for their implementations.

04 and 05 are the same binary and probably the same owner — 05 is split out only
because it is a real milestone boundary. Merge them if one person takes both.

## Sequencing

| Wave | Work | Blocked by |
| --- | --- | --- |
| 0 | Agree `contracts.md` (C0–C4) | — |
| 1 | **01** harvest + redesign · **03** data model + external execution | contracts |
| 2 | **02** config + branch-run step + publish · **04** A/B | C1, C2, C4 |
| 3 | **03** UI · **04** C · **05** | wave 2 |

Waves 1 and 2 are each two-person-parallel. Wave 0 is small and is what makes
that true; it should be agreed before anyone writes code.

## Branch state (2026-08-12)

Read this before starting.

**The air-gap MVP is not on this branch.** Commit `982ff3ced` (branch
`amp/airgap-bundle-mvp`) is *not an ancestor of HEAD*. `pkg/runner/oci/bundle`,
`pkg/runner/airgap`, `bins/nuon-bundle`, and `services/bundle-portal` do not exist
on disk. That commit is a **reference implementation to harvest**, read with git:

```bash
git show 982ff3ced --stat
git ls-tree -r 982ff3ced --name-only | grep -E 'pkg/runner/(oci/bundle|airgap)'
git show 982ff3ced:pkg/runner/oci/bundle/bundle.go
```

Where a spec cites a path that does not exist on disk, it is citing that commit
and says so.

**Two build breaks, both pre-existing and unrelated to this work:**

1. `services/ctl-api/internal/app/airgap/` held two orphaned gitignored `*_gen.go`
   files from that checkout. **Already deleted** — nothing tracked was lost.
2. `services/ctl-api/internal/app/runners/signals/processjob/signal.go:645`
   references `activities.AwaitRecordJobLifecycleCompositeError`, a generated
   Temporal wrapper that has not been generated. Introduced by `18683be3a`
   ("surface runner job lifecycle errors"). Fix with
   `go generate ./services/ctl-api/...` — **not** part of this work, but you will
   hit it on your first build.

## Naming

The RFC and the HTML diagram predate the current naming. Where they disagree with
these specs, the specs win:

| RFC / HTML | These specs | Why |
| --- | --- | --- |
| `execution_mode: installer` | `Workflow.ExecutionType = external` | Current naming; see C0 |
| `BUNDLE` | `AppConfigBundle` | Current naming |
| `InstallerAgent` join secret + HMAC handshake | install-id + install-token | Reuses the existing bootstrap-token precedent; see C4 |

Note `app.Installer` **already exists** (`services/ctl-api/internal/app/installer.go`)
— a dead self-hosted-installer model with only a migration, an org relation, and a
hard-delete entry. No routes, no services, no handlers. The new connected-CLI
model is therefore named `InstallerAgent`. Dropping or reclaiming the dead model is
separable cleanup and is not in scope for any spec here.
