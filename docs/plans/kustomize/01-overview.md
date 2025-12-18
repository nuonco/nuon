# Kustomize Support: Overview & Architecture

## Overview

Add Kustomize support to the existing Kubernetes manifest component, enabling users to deploy Kustomize overlays while maintaining backwards compatibility with existing inline manifest deployments.

## Goals

1. **Backwards Compatible**: Existing `manifest` field continues to work as-is
2. **Unified Build Pipeline**: ALL manifests (inline, kustomize, BYOA) flow through OCI artifacts
3. **Build Runner**: Package manifests into OCI artifacts using kustomizer.dev Go SDK
4. **Install Runner**: Always pull and apply from OCI artifacts using kustomizer patterns (single code path)
5. **BYOA (Bring Your Own Artifact)**: Allow users to provide pre-packaged OCI artifacts

---

## Key Dependencies: Kustomizer.dev Go SDK

This implementation uses the **kustomizer.dev** Go SDK and its underlying libraries:

| Package | Purpose | Import Path |
|---------|---------|-------------|
| **kustomizer/pkg/registry** | OCI artifact push/pull operations | `github.com/stefanprodan/kustomizer/pkg/registry` |
| **fluxcd/pkg/ssa** | Server-side apply, object serialization | `github.com/fluxcd/pkg/ssa` |
| **kustomize/api** | Kustomize overlay building | `sigs.k8s.io/kustomize/api/krusty` |
| **go-containerregistry** | OCI image manipulation (used by kustomizer) | `github.com/google/go-containerregistry/pkg/crane` |

### Kustomizer SDK Key Functions

**Build Runner Operations:**
```go
// Build kustomize overlays → YAML
objects, _, err := buildManifests(ctx, kustomizePath, filePaths, artifacts, patches, identities)
yml, err := ssa.ObjectsToYAML(objects)

// Push to OCI registry (equivalent to "kustomizer push artifact")
digest, err := registry.Push(ctx, url, []byte(yml), &registry.Metadata{
    Version:        "1.0.0",
    Checksum:       fmt.Sprintf("%x", sha256.Sum256([]byte(yml))),
    Created:        time.Now().UTC().Format(time.RFC3339),
    SourceURL:      gitURL,
    SourceRevision: gitRevision,
}, nil) // no encryption
```

**Install Runner Operations:**
```go
// Pull from OCI registry (equivalent to "kustomizer pull artifact")
content, meta, err := registry.Pull(ctx, artifactURL, nil) // no decryption

// Parse YAML to Kubernetes objects
objects, err := ssa.ReadObjects(strings.NewReader(content))

// Apply using server-side apply (ResourceManager from fluxcd/pkg/ssa)
resMgr := ssa.NewResourceManager(client, nil, ssa.Owner{Field: "nuon-runner"})
changeSet, err := resMgr.ApplyAll(ctx, objects, applyOpts)
```

---

## Architecture

### Key Design Decision: Unified OCI Artifact Pipeline

**All kubernetes manifest deployments flow through OCI artifacts.** When a user provides an inline `manifest`, the build runner treats it as an implicit kustomization (empty `kustomization.yaml` + the manifest as a resource), packages it into an OCI artifact, and the install runner always pulls from the artifact.

**Benefits:**
- Single code path in install runner (always pull from OCI)
- Consistent artifact storage and versioning
- Enables future features: signing, caching, rollback to previous artifacts
- Simplifies drift detection (compare against stored artifact)

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
                                                           │     
  ┌─────────────┐                                          │      ┌─────────────┐
  │  kustomize  │ ──► Run `kustomize build` on path ───────┼──►   │             │
  │   (path)    │     + apply patches                      │      │  OCI        │
  └─────────────┘                                          ├──►   │  Artifact   │
                                                           │      │  Registry   │
  ┌─────────────┐                                          │      │             │
  │ oci_artifact│ ──► Mirror/copy to Nuon registry ────────┘      └──────┬──────┘
  │   (BYOA)    │                                                        │
  └─────────────┘                                                        │
                                                                         ▼
                                                                 ┌──────────────┐
                                                                 │ Pull artifact│
                                                                 │ Extract YAML │
                                                                 │ Apply to K8s │
                                                                 └──────────────┘
```

---

## Document Index

This plan is broken down into the following documents:

| Document | Description |
|----------|-------------|
| [01-overview.md](./01-overview.md) | This document - goals, architecture, SDK overview |
| [02-config-schema.md](./02-config-schema.md) | Configuration schema changes (`pkg/config/`) |
| [03-build-runner.md](./03-build-runner.md) | Build runner implementation (`bins/runner/internal/jobs/build/`) |
| [04-install-runner.md](./04-install-runner.md) | Install runner changes (`bins/runner/internal/jobs/deploy/`) |
| [05-examples-migration.md](./05-examples-migration.md) | Configuration examples & migration guide |
| [06-api-data-model.md](./06-api-data-model.md) | API & database schema changes (`services/ctl-api/`) |
| [07-future-byoa-support.md](./07-future-byoa-support.md) | Future BYOA (Bring Your Own Artifact) roadmap |
| [08-ctl-api-workflow-changes.md](./08-ctl-api-workflow-changes.md) | **ctl-api Temporal workflow & job type changes** |

---

## File Change Summary

### New Files

#### ctl-api (Workflow Orchestration)
| Path | Description |
|------|-------------|
| `pkg/plans/types/kubernetes_manifest_build_plan.go` | Build plan type for kubernetes manifest |
| `services/ctl-api/internal/app/components/worker/plan/plan_kubernetes_manifest_build.go` | Build plan generator |

#### Build Runner
| Path | Description |
|------|-------------|
| `bins/runner/internal/jobs/build/kubernetes_manifest/handler.go` | Build handler main struct |
| `bins/runner/internal/jobs/build/kubernetes_manifest/meta.go` | Handler metadata |
| `bins/runner/internal/jobs/build/kubernetes_manifest/fetch.go` | Source fetching |
| `bins/runner/internal/jobs/build/kubernetes_manifest/initialize.go` | Workspace initialization |
| `bins/runner/internal/jobs/build/kubernetes_manifest/validate.go` | Kustomization validation |
| `bins/runner/internal/jobs/build/kubernetes_manifest/exec.go` | Build execution (kustomizer SDK) |
| `bins/runner/internal/jobs/build/kubernetes_manifest/kustomize.go` | Kustomize build helpers |
| `bins/runner/internal/jobs/build/kubernetes_manifest/outputs.go` | Output handling |
| `bins/runner/internal/jobs/build/kubernetes_manifest/cleanup.go` | Cleanup logic |
| `bins/runner/internal/jobs/build/kubernetes_manifest/reset.go` | State reset |
| `bins/runner/internal/jobs/build/kubernetes_manifest/state.go` | Handler state |

#### Install Runner (Optional Enhancement)
| Path | Description |
|------|-------------|
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/exec_ssa.go` | Optional SSA ResourceManager apply |

### Modified Files

#### ctl-api (Workflow Orchestration)
| Path | Changes |
|------|---------|
| `services/ctl-api/internal/app/runner_job.go` | Add `RunnerJobTypeKubernetesManifestBuild` constant |
| `services/ctl-api/internal/app/component.go` | Update `BuildJobType()` and `SyncJobType()` mappings |
| `services/ctl-api/internal/app/components/worker/plan/plan_component_build.go` | Add kubernetes manifest case |
| `services/ctl-api/internal/app/installs/worker/plan/plan_kubernetes_manifest_deploy.go` | Use OCI artifact instead of inline manifest |
| `pkg/plans/types/build_plan.go` | Add `KubernetesManifestBuildPlan` field |

#### Config & Plan Types
| Path | Changes |
|------|---------|
| `pkg/config/kubernetes_manifest_component.go` | Add `Kustomize` and `OCIArtifact` config structs |
| `pkg/plans/types/kubernetes_manifest_deploy_plan.go` | Add `OCIArtifact` field, update Manifest to runtime-only |

#### Build Runner
| Path | Changes |
|------|---------|
| `bins/runner/internal/jobs/build/fx.go` | Register new build handler |

#### Install Runner
| Path | Changes |
|------|---------|
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/initialize.go` | Use `registry.Pull()` for OCI artifact |
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/state.go` | Add `manifestObjects`, `artifactMetadata` fields |
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/fetch.go` | Conditionally fetch artifact config |

#### Dependencies
| Path | Changes |
|------|---------|
| `go.mod` | Add kustomizer SDK and FluxCD SSA dependencies |

---

## Resolved Questions

1. **Kustomize Version**: Pin to latest stable **v5.4.x** (e.g., `v5.4.3`). Bundle the kustomize binary in the runner image.

2. **Helm Inflation**: Enable Helm chart inflation support. Opt-in via `enable_helm: true` in kustomize config.

3. **Build Caching**: No build caching. Each build produces a fresh OCI artifact.

4. **Secret Handling**: Deferred. No SOPS/encrypted secrets support in initial implementation.

5. **Multi-Document Artifacts**: One component → one kustomize path → one OCI artifact.

---

## References

### Kustomizer SDK Documentation

| Resource | Link | Description |
|----------|------|-------------|
| **Kustomizer Documentation** | [kustomizer.dev](https://kustomizer.dev/) | Official documentation site |
| **Kustomizer GitHub** | [github.com/stefanprodan/kustomizer](https://github.com/stefanprodan/kustomizer) | Source code and examples |
| **pkg/registry** | [registry package](https://github.com/stefanprodan/kustomizer/tree/main/pkg/registry) | OCI artifact Push/Pull functions |
| **FluxCD SSA** | [pkg.go.dev/github.com/fluxcd/pkg/ssa](https://pkg.go.dev/github.com/fluxcd/pkg/ssa) | Server-side apply utilities |
| **Kustomize API** | [sigs.k8s.io/kustomize](https://github.com/kubernetes-sigs/kustomize) | Kustomize Go SDK |

### Existing Nuon Runner Patterns

- [Existing Terraform Build Handler](../../bins/runner/internal/jobs/build/terraform/exec.go)
- [Existing Helm Deploy Handler](../../bins/runner/internal/jobs/deploy/helm/)
- [Existing OCI Archive Pack/Unpack (ORAS)](../../bins/runner/internal/pkg/oci/archive/)
- [Existing Kubernetes Manifest Deploy Handler](../../bins/runner/internal/jobs/deploy/kubernetes_manifest/)
