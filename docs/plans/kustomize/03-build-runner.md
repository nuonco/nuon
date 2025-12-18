# Kustomize Support: Build Runner Implementation

This document covers the new build handler for Kubernetes manifests in `bins/runner/internal/jobs/build/kubernetes_manifest/`.

> **Key Decisions**:
> - Use kustomize SDK (`krusty`) for build + existing ORAS for push
> - Consolidated handler (6 files vs original 11)
> - `Fetch()` is no-op for inline manifests
> - No template rendering - manifests applied as-is

---

## Overview

The build runner packages manifests into OCI artifacts. It handles two modes (MVP):
- **Mode A**: Inline manifest → write to temp file → ORAS pack+push
- **Mode B**: Kustomize path → `krusty.Run()` → get YAML → ORAS pack+push

> **Note**: Mode C (BYOA - pre-built OCI artifact) is deferred to post-MVP.

---

## New Directory Structure

**New Directory**: `bins/runner/internal/jobs/build/kubernetes_manifest/`

**Files to create** (consolidated from original 11 to 6):
| File | Description |
|------|-------------|
| `handler.go` | Handler struct, FX constructor, GracefulShutdown |
| `meta.go` | JobType(), Name() |
| `state.go` | handlerState struct |
| `exec.go` | Main logic: inline → YAML, kustomize → YAML, ORAS push |
| `fetch.go` | Fetch git source (no-op for inline) |
| `methods.go` | Reset(), Cleanup(), Validate(), Initialize(), Outputs() stubs |

---

## Build Execution Logic

**File**: `bins/runner/internal/jobs/build/kubernetes_manifest/exec.go`

This implementation uses kustomize SDK for build and existing ORAS infrastructure for push.

```go
package kubernetes_manifest

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/nuonco/nuon-runner-go/models"
    "go.uber.org/zap"
    "sigs.k8s.io/kustomize/api/krusty"
    "sigs.k8s.io/kustomize/kyaml/filesys"

    pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
    "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
    l, _ := pkgctx.Logger(ctx)
    
    plan := h.state.plan
    
    var manifestYAML []byte
    var err error
    
    switch plan.SourceType {
    case "inline":
        // Mode A: Inline manifest - use directly
        l.Info("using inline manifest")
        manifestYAML = []byte(plan.InlineManifest)
        
    case "kustomize":
        // Mode B: Kustomize path - build overlay
        l.Info("building kustomize overlay", zap.String("path", plan.KustomizePath))
        manifestYAML, err = h.buildKustomization(ctx, plan.KustomizePath)
        if err != nil {
            return fmt.Errorf("kustomize build failed: %w", err)
        }
    }
    
    l.Info("manifest ready", zap.Int("yaml_size", len(manifestYAML)))
    
    // Write manifest to temp file for ORAS pack
    manifestPath := filepath.Join(h.state.workDir, "manifest.yaml")
    if err := os.WriteFile(manifestPath, manifestYAML, 0644); err != nil {
        return fmt.Errorf("failed to write manifest: %w", err)
    }
    
    // Pack and push using existing ORAS infrastructure
    l.Info("pushing artifact to registry", zap.String("url", h.state.dstURL))
    
    digest, err := h.ociCopy.PackAndPush(ctx, h.state.workDir, h.state.dstCfg, h.state.dstTag)
    if err != nil {
        return fmt.Errorf("failed to push artifact: %w", err)
    }
    
    l.Info("artifact pushed successfully", zap.String("digest", digest))
    
    // Report result to API
    result := &models.CreateJobExecutionResultReq{
        OCIArtifact: &models.OCIArtifactResult{
            URL:    h.state.dstURL,
            Digest: digest,
        },
    }
    _, err = h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, result)
    return err
}
```

---

## Kustomize Build Helper

The `buildKustomization` helper is included in `exec.go` (no separate file needed):

```go
// buildKustomization runs kustomize build on the given path and returns YAML bytes
func (h *handler) buildKustomization(ctx context.Context, kustomizePath string) ([]byte, error) {
    opts := krusty.MakeDefaultOptions()
    
    // Configure load restrictor
    if h.state.plan.KustomizeConfig != nil {
        switch h.state.plan.KustomizeConfig.LoadRestrictor {
        case "none":
            opts.LoadRestrictions = types.LoadRestrictionsNone
        default:
            opts.LoadRestrictions = types.LoadRestrictionsRootOnly
        }
        
        // Enable Helm chart inflation if requested
        if h.state.plan.KustomizeConfig.EnableHelm {
            opts.PluginConfig.HelmConfig.Enabled = true
        }
    }
    
    k := krusty.MakeKustomizer(opts)
    resMap, err := k.Run(filesys.MakeFsOnDisk(), kustomizePath)
    if err != nil {
        return nil, err
    }
    
    return resMap.AsYaml()
}
```

> **Note**: Additional patches (`KustomizeConfig.Patches`) can be applied by creating a temporary kustomization.yaml that references the built output and the patches, then running kustomize again. Implementation deferred to iteration.

---

## Handler Registration

**File**: `bins/runner/internal/jobs/build/fx.go`

```go
func init() {
    // Add to existing handler registrations
    buildHandlers = append(buildHandlers, 
        // ... existing handlers
        kubernetes_manifest.New,
    )
}
```

---

