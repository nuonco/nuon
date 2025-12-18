# Kustomize Support for Kubernetes Manifest Component

## Implementation Status

> **Status**: 🟡 In Progress
> 
> **Completed**:
> - ✅ Config & API changes (Phase 1)
> - ✅ Plan types (Phase 2 partial)
> 
> **Remaining**:
> - ⏳ Build plan generator (`plan_kubernetes_manifest_build.go`)
> - ⏳ Build runner handler (`bins/runner/internal/jobs/build/kubernetes_manifest/`)
> - ⏳ Install runner updates (OCI artifact pull)
> - ⏳ Deploy plan updates (`plan_kubernetes_manifest_deploy.go`)
> - ⏳ Kustomize SDK dependencies in `go.mod`
> - ⏳ CLI sync changes (`pkg/config/sync/components_kubernetes_manifest.go`)
> - ⏳ nuon-go client library update (external dependency)

---

## Overview

Add Kustomize support to the existing Kubernetes manifest component, enabling users to deploy Kustomize overlays while maintaining backwards compatibility with existing inline manifest deployments.

## Goals (MVP)

1. **Backwards Compatible**: Existing `manifest` field continues to work as-is
2. **Unified Build Pipeline**: ALL manifests (inline, kustomize) flow through OCI artifacts
3. **Build Runner**: Package manifests into OCI artifacts using kustomize SDK (`krusty`) + existing ORAS infrastructure
4. **Install Runner**: Always pull and apply from OCI artifacts (single code path)

> **Note**: BYOA (Bring Your Own Artifact) is deferred from MVP. See [07-future-byoa-support.md](./kustomize/07-future-byoa-support.md) for the future roadmap.

---

## Key Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| 1 | **Namespace-only templating** - manifest content not rendered | Simplifies build pipeline; namespace rendered at deploy time with install state |
| 2 | **Flat config schema** - `Manifest` / `Kustomize` as siblings | Simpler for end users |
| 3 | **Kustomize SDK + ORAS** - use `krusty` for build, existing ORAS for push | Reuses existing auth infrastructure |
| 4 | **Always use new build job** for all k8s manifest components | Unified pipeline |
| 5 | **Consolidated build handler** - 6 files, `Fetch()` no-op for inline | Minimal code footprint |
| 6 | **Install runner uses archive pattern** (like Helm) | Consistent with existing handlers |
| 7 | **Strip OCIArtifact for MVP** - only Kustomize fields | Keep MVP minimal, add BYOA later |

### Breaking Changes

> ⚠️ **Breaking**: The following features are removed in MVP:
> - Inline manifests no longer support `{{.nuon.install.id}}` template variables in manifest content
> - Users should use kustomize overlays for install-specific manifest customization
>
> ✅ **Preserved**: `Namespace` field continues to support template variables (e.g., `{{.nuon.install.id}}`)

---

## Detailed Documentation

This plan has been broken down into focused documents:

| Document | Description |
|----------|-------------|
| [01-overview.md](./kustomize/01-overview.md) | Architecture, SDK overview, file change summary |
| [02-config-schema.md](./kustomize/02-config-schema.md) | Configuration schema changes (`pkg/config/`) |
| [03-build-runner.md](./kustomize/03-build-runner.md) | Build runner implementation (`bins/runner/internal/jobs/build/`) |
| [04-install-runner.md](./kustomize/04-install-runner.md) | Install runner changes (`bins/runner/internal/jobs/deploy/`) |
| [05-examples-migration.md](./kustomize/05-examples-migration.md) | Configuration examples & migration guide |
| [06-api-data-model.md](./kustomize/06-api-data-model.md) | API & database schema changes (`services/ctl-api/`) |
| [07-future-byoa-support.md](./kustomize/07-future-byoa-support.md) | Future BYOA (Bring Your Own Artifact) roadmap |
| [08-ctl-api-workflow-changes.md](./kustomize/08-ctl-api-workflow-changes.md) | **ctl-api Temporal workflow & job type changes** |
| [09-cli-sync-changes.md](./kustomize/09-cli-sync-changes.md) | **CLI sync changes for kustomize support** |

---

## Quick Links

### For Developers
- **ctl-api Workflows**: [08-ctl-api-workflow-changes.md](./kustomize/08-ctl-api-workflow-changes.md) ⭐ Start here for backend changes
- **Config Schema**: [02-config-schema.md](./kustomize/02-config-schema.md)
- **API & Data Model**: [06-api-data-model.md](./kustomize/06-api-data-model.md)
- **Build Runner**: [03-build-runner.md](./kustomize/03-build-runner.md)
- **Install Runner**: [04-install-runner.md](./kustomize/04-install-runner.md)

### For Users
- **Examples & Migration**: [05-examples-migration.md](./kustomize/05-examples-migration.md)

---

## Architecture Summary

```
┌────────────────────────────────────────────────────────────────────────────────┐
│                        UNIFIED BUILD → DEPLOY FLOW                              │
└────────────────────────────────────────────────────────────────────────────────┘

  CONFIG INPUT                    BUILD RUNNER                    INSTALL RUNNER
  ────────────                    ────────────                    ──────────────

  ┌─────────────┐                                                
  │   manifest  │ ──► Create implicit kustomization.yaml ──┐     
  │  (inline)   │     + manifest.yaml as resource          │     
  └─────────────┘                                          │     
                                                           │      ┌─────────────┐
  ┌─────────────┐                                          │      │             │
  │  kustomize  │ ──► Run `kustomize build` on path ───────┼──►   │  OCI        │
  │   (path)    │     + apply patches                      │      │  Artifact   │
  └─────────────┘                                          │      │  Registry   │
                                                           │      │             │
  ┌─────────────┐                                          │      └──────┬──────┘
  │ oci_artifact│ ──► (Future: Mirror to Nuon registry) ───┘             │
  │   (BYOA)    │     See 07-future-byoa-support.md                      │
  └─────────────┘                                                        ▼
                                                                 ┌──────────────┐
                                                                 │ Pull artifact│
                                                                 │ Extract YAML │
                                                                 │ Apply to K8s │
                                                                 └──────────────┘
```

---

## File Change Summary

### New Files

| Path | Description | Status |
|------|-------------|--------|
| `pkg/plans/types/kubernetes_manifest_build_plan.go` | Build plan type | ✅ Done |
| `services/ctl-api/.../plan/plan_kubernetes_manifest_build.go` | Build plan generator | ⏳ TODO |
| `bins/runner/internal/jobs/build/kubernetes_manifest/*.go` | Build handler (6 files: handler.go, meta.go, state.go, exec.go, fetch.go, methods.go) | ⏳ TODO |

### Modified Files

| Path | Changes | Status |
|------|---------|--------|
| `services/ctl-api/internal/app/runner_job.go` | Add `RunnerJobTypeKubernetesManifestBuild` | ✅ Done |
| `services/ctl-api/internal/app/component.go` | Update `BuildJobType()` → kubernetes-manifest-build, `SyncJobType()` → oci-sync | ✅ Done |
| `services/ctl-api/internal/app/kubernetes_manifest_component_config.go` | Add `KustomizeSourceConfig` JSONB type, `SourceType()` method | ✅ Done |
| `services/ctl-api/.../components/service/create_kubernetes_manifest_component_config.go` | Add `KustomizeConfigRequest`, update validation & handler | ✅ Done |
| `services/ctl-api/.../plan/plan_component_build.go` | Add kubernetes manifest case | ⏳ TODO |
| `services/ctl-api/.../plan/plan_kubernetes_manifest_deploy.go` | Use OCI artifact reference, remove manifest template rendering (keep namespace rendering) | ⏳ TODO |
| `pkg/config/kubernetes_manifest_component.go` | Add `Kustomize` config (no OCIArtifact for MVP) | ✅ Done |
| `pkg/plans/types/build_plan.go` | Add `KubernetesManifestBuildPlan` field | ✅ Done |
| `pkg/plans/types/kubernetes_manifest_deploy_plan.go` | Add `OCIArtifact` field | ✅ Done |
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/initialize.go` | OCI artifact pull via archive pattern | ⏳ TODO |
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/state.go` | Add manifest from artifact | ⏳ TODO |
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/handler.go` | Add archive dependency | ⏳ TODO |
| `bins/runner/internal/jobs/build/fx.go` | Register new handler | ⏳ TODO |
| `go.mod` | Add kustomize SDK dependencies (`sigs.k8s.io/kustomize/api`, `github.com/fluxcd/pkg/ssa`) | ⏳ TODO |
| `pkg/config/sync/components_kubernetes_manifest.go` | Map Kustomize config to API request | ⏳ TODO |
| `go.mod` | Bump nuon-go dependency (after nuon-go update) | ⏳ TODO |

---

## Resolved Questions

1. **Kustomize Version**: Pin to **v5.4.x**
2. **Helm Inflation**: Opt-in via `enable_helm: true`
3. **Build Caching**: None - fresh artifact each build
4. **Secret Handling**: Deferred - no SOPS support initially
5. **Multi-Document**: One component → one kustomize path → one artifact
