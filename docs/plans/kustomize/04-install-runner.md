# Kustomize Support: Install Runner Changes

This document covers the changes needed in `bins/runner/internal/jobs/deploy/kubernetes_manifest/`.

---

## Overview

The install runner is updated to pull manifests from OCI artifacts. The key changes are:
- **Initialize()**: Pull OCI artifact using `registry.Pull()`
- **Exec()**: Existing diff logic remains unchanged (works with unstructured objects)
- **Optional**: SSA ResourceManager for enhanced apply semantics

---

## Initialize: OCI Artifact Pull

**File**: `bins/runner/internal/jobs/deploy/kubernetes_manifest/initialize.go`

```go
package kubernetes_manifest

import (
    "context"
    "fmt"
    "strings"

    "github.com/nuonco/nuon-runner-go/models"
    "github.com/stefanprodan/kustomizer/pkg/registry"
    "go.uber.org/zap"

    pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

func (h *handler) Initialize(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
    l, _ := pkgctx.Logger(ctx)
    
    // All kubernetes manifest deployments now come from OCI artifacts
    // (inline manifests are packaged into artifacts during build phase)
    artifact := h.state.plan.KubernetesManifestDeployPlan.OCIArtifact
    if artifact == nil {
        return fmt.Errorf("OCI artifact reference is required - this indicates a build phase issue")
    }
    
    // Build artifact URL (following kustomizer URL format)
    artifactURL := artifact.URL
    if artifact.Digest != "" {
        // Prefer digest for immutable references
        artifactURL = fmt.Sprintf("%s@%s", artifactURL, artifact.Digest)
    } else if artifact.Tag != "" {
        artifactURL = fmt.Sprintf("%s:%s", artifactURL, artifact.Tag)
    }
    
    l.Info("pulling OCI artifact using kustomizer registry.Pull", 
        zap.String("url", artifactURL))
    
    // Pull artifact using kustomizer SDK (equivalent to "kustomizer pull artifact")
    // This returns the manifest YAML content directly
    content, meta, err := registry.Pull(ctx, artifactURL, nil) // no decryption
    if err != nil {
        return fmt.Errorf("failed to pull artifact: %w", err)
    }
    
    l.Info("artifact pulled successfully",
        zap.String("digest", meta.Digest),
        zap.String("checksum", meta.Checksum),
        zap.String("created", meta.Created),
        zap.Int("content_size", len(content)))
    
    // Store manifest content in plan for Exec() to use with existing parsing
    // Exec() will call getKubernetesResourcesFromManifest() which handles:
    // - YAML parsing to unstructured objects
    // - GVR resolution via Kubernetes discovery API
    // - Namespace determination for each resource
    h.state.plan.KubernetesManifestDeployPlan.Manifest = content
    
    // Store artifact metadata for drift detection
    h.state.artifactMetadata = meta
    
    return nil
}
```

---

## Diff Logic Compatibility

**Key Design Decision**: The existing diff detection logic in `exec.go` requires **no changes**.

The current implementation compares `*unstructured.Unstructured` objects regardless of their source. Both parsing approaches produce the same type:

| Approach | Parser | Output Type |
|----------|--------|-------------|
| Current | `yaml.NewYAMLOrJSONDecoder` | `*unstructured.Unstructured` |
| Kustomizer | `ssa.ReadObjects()` | `*unstructured.Unstructured` ✓ |

**How it works:**

1. **Initialize()** pulls the OCI artifact and stores the raw YAML in `h.state.plan.KubernetesManifestDeployPlan.Manifest`
2. **Exec()** calls `getKubernetesResourcesFromManifest()` which:
   - Parses the manifest string into `[]*kubernetesResource`
   - Resolves GVR mappings via Kubernetes discovery API
   - Determines if resources are namespaced
3. Existing diff functions work unchanged:
   - `fetchLiveResources()` - gets current cluster state
   - `resourceDiffWithLive()` - compares desired vs live
   - `diff.DetectChanges()` - generates detailed diff entries
   - `execApply()` / `execDelete()` - applies changes with dry-run support

**Why we keep `getKubernetesResourcesFromManifest()`:**

While `ssa.ReadObjects()` parses YAML to unstructured objects, `getKubernetesResourcesFromManifest()` additionally:
- Resolves `GroupVersionResource` via discovery mapper
- Determines if resources are namespaced
- Creates the `kubernetesResource` wrapper with all metadata needed for apply/delete

We use `ssa.ReadObjects()` in the build runner (where we don't have cluster access) and continue using the existing parsing in the install runner (where we need GVR resolution).

---

## State Updates

**File**: `bins/runner/internal/jobs/deploy/kubernetes_manifest/state.go`

```go
package kubernetes_manifest

import (
    "time"

    "github.com/nuonco/nuon-runner-go/models"
    "github.com/stefanprodan/kustomizer/pkg/registry"

    plantypes "github.com/powertoolsdev/mono/pkg/types/components/plan"
)

type handlerState struct {
    plan           *plantypes.DeployPlan
    appCfg         *models.AppAppConfig
    jobID          string
    jobExecutionID string
    timeout        time.Duration
    kubeClient     *kubernetesClient
    outputs        map[string]interface{}
    
    // NEW: Artifact metadata from kustomizer registry.Pull
    // Used for drift detection (compare digest with last deployed)
    artifactMetadata *registry.Metadata
}
```

---

## Optional Enhancement: SSA ResourceManager

The existing `Exec()` function can optionally be enhanced to use the `fluxcd/pkg/ssa.ResourceManager` for server-side apply. This provides:
- Automatic CRD/Namespace ordering (applies CRDs first, waits, then applies resources)
- Garbage collection of stale resources
- Wait for readiness
- Better conflict handling

**File**: `bins/runner/internal/jobs/deploy/kubernetes_manifest/exec_ssa.go` (Optional new file)

```go
package kubernetes_manifest

import (
    "context"
    "fmt"
    "time"

    "github.com/fluxcd/pkg/ssa"
    "go.uber.org/zap"
    "sigs.k8s.io/controller-runtime/pkg/client"

    pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
)

// execApplySSA applies manifests using fluxcd/pkg/ssa ResourceManager
// This provides enhanced features like automatic CRD ordering and garbage collection
func (h *handler) execApplySSA(ctx context.Context, l *zap.Logger) error {
    
    // Create SSA ResourceManager (following kustomizer apply inventory pattern)
    resMgr := ssa.NewResourceManager(
        h.state.kubeClient.client, // controller-runtime client
        nil,                        // no poller
        ssa.Owner{
            Field: "nuon-runner",
            Group: "nuon.co",
        },
    )
    
    objects := h.state.manifestObjects
    namespace := h.state.plan.KubernetesManifestDeployPlan.Namespace
    
    // Set namespace on objects that don't have one
    for _, obj := range objects {
        if obj.GetNamespace() == "" {
            obj.SetNamespace(namespace)
        }
    }
    
    // Apply options
    applyOpts := ssa.ApplyOptions{
        Force:        true,  // Force apply to resolve conflicts
        FieldManager: "nuon-runner",
    }
    
    // Apply all objects with staged apply (CRDs first, then resources)
    // This is equivalent to "kustomizer apply inventory" behavior
    l.Info("applying manifests using SSA staged apply", 
        zap.Int("object_count", len(objects)))
    
    changeSet, err := resMgr.ApplyAllStaged(ctx, objects, applyOpts)
    if err != nil {
        return fmt.Errorf("SSA apply failed: %w", err)
    }
    
    // Log changes
    for _, change := range changeSet.Entries {
        l.Info("resource applied",
            zap.String("subject", change.Subject),
            zap.String("action", string(change.Action)))
    }
    
    // Optionally wait for resources to become ready
    waitOpts := ssa.WaitOptions{
        Interval: 5 * time.Second,
        Timeout:  h.state.timeout,
    }
    
    l.Info("waiting for resources to become ready")
    if err := resMgr.Wait(objects, waitOpts); err != nil {
        l.Warn("some resources not ready", zap.Error(err))
        // Don't fail on wait timeout - resources may still be reconciling
    }
    
    l.Info("apply completed successfully",
        zap.Int("changes", len(changeSet.Entries)))
    
    return nil
}
```

**Note**: The existing `Exec()` function in exec.go can continue to work with the current dynamic client approach. The SSA ResourceManager is an optional enhancement that provides Flux-style apply semantics.

---

## Implementation Notes: Kustomizer SDK vs ORAS

### Why Kustomizer SDK over raw ORAS?

The existing Nuon runner uses ORAS library directly for OCI operations (see `bins/runner/internal/pkg/oci/archive/`). The kustomizer SDK provides several advantages for Kubernetes manifest artifacts:

| Feature | Raw ORAS | Kustomizer SDK |
|---------|----------|----------------|
| **Artifact Format** | Generic blob storage | Kubernetes-specific (multi-doc YAML) |
| **Metadata** | Manual annotation management | Built-in metadata (checksum, version, source) |
| **Compression** | Manual implementation | Automatic tar/gzip handling |
| **Encryption** | Not included | Optional age encryption support |
| **Compatibility** | Custom format | Compatible with Flux, other GitOps tools |
| **Object Handling** | Manual YAML parsing | SSA integration (ReadObjects, ObjectsToYAML) |

### Integration Strategy

**Option A: Use kustomizer SDK directly (Recommended)**
- Import `github.com/stefanprodan/kustomizer/pkg/registry`
- Use `registry.Push()` and `registry.Pull()` functions
- Artifacts compatible with `kustomizer` CLI and Flux
- Simpler implementation, well-tested code

**Option B: Adapt kustomizer patterns to existing ORAS infrastructure**
- Keep using `bins/runner/internal/pkg/oci/archive` for pack/unpack
- Add kustomizer metadata annotations manually
- More integration work, but consistent with existing patterns

**Recommended: Option A** - Use kustomizer SDK directly for manifest artifacts.

---

