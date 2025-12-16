# Kustomize Support for Kubernetes Manifest Component

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

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           CONFIGURATION LAYER                                    │
├─────────────────────────────────────────────────────────────────────────────────┤
│  KubernetesManifestComponentConfig                                              │
│  ├── manifest (existing - inline YAML) ─────────────────┐                       │
│  ├── kustomize (NEW - kustomize configuration) ─────────┼──► All get packaged  │
│  │   ├── path: "./overlays/production"                  │    into OCI artifacts │
│  │   ├── patches: ["patch1.yaml", "patch2.yaml"]       │                       │
│  │   └── enable_helm: false                             │                       │
│  ├── oci_artifact (NEW - pre-packaged artifact) ────────┘                       │
│  │   ├── url: "oci://registry/org/app-config"                                  │
│  │   ├── tag: "v1.0.0"                                                          │
│  │   └── digest: "sha256:..." (optional, for pinning)                          │
│  └── namespace                                                                  │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              BUILD RUNNER                                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│  NEW: kubernetes_manifest build handler                                          │
│                                                                                  │
│  Mode A: Inline Manifest (implicit kustomization)                               │
│  ┌────────────────────────────────────────────────────────────────────┐        │
│  │ 1. Write manifest to temp file (manifest.yaml)                     │        │
│  │ 2. Generate implicit kustomization.yaml with resource reference    │        │
│  │ 3. Run `kustomize build` (applies consistent processing)          │        │
│  │ 4. Package rendered manifest into OCI artifact                     │        │
│  │ 5. Push to Nuon artifact registry                                  │        │
│  └────────────────────────────────────────────────────────────────────┘        │
│                                                                                  │
│  Mode B: Kustomize Path (explicit kustomization)                                │
│  ┌────────────────────────────────────────────────────────────────────┐        │
│  │ 1. Fetch source (git clone)                                        │        │
│  │ 2. Run `kustomize build` on specified path                         │        │
│  │ 3. Apply any additional patches                                    │        │
│  │ 4. Package rendered manifests into OCI artifact                    │        │
│  │ 5. Push to Nuon artifact registry                                  │        │
│  └────────────────────────────────────────────────────────────────────┘        │
│                                                                                  │
│  Mode C: Pre-built OCI Artifact (BYOA)                                          │
│  ┌────────────────────────────────────────────────────────────────────┐        │
│  │ 1. Validate OCI artifact exists and is accessible                  │        │
│  │ 2. Copy/mirror to Nuon registry                                    │        │
│  │ 3. Record artifact reference in build result                       │        │
│  └────────────────────────────────────────────────────────────────────┘        │
│                                                                                  │
│  OUTPUT: All modes produce an OCI artifact reference                            │
└─────────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                             INSTALL RUNNER                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│  Enhanced kubernetes_manifest deploy handler (exec.go)                          │
│                                                                                  │
│  Single Code Path (always OCI artifact):                                        │
│  ┌────────────────────────────────────────────────────────────────────┐        │
│  │ 1. Pull OCI artifact from plan.OCIArtifact reference              │        │
│  │ 2. Extract manifest.yaml from artifact                            │        │
│  │ 3. Parse YAML documents                                            │        │
│  │ 4. Create/Apply/Teardown plan operations                          │        │
│  │ 5. Server-side apply with diff detection                          │        │
│  │ 6. Report results                                                  │        │
│  └────────────────────────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Detailed Implementation Plan

### Phase 1: Configuration Schema Updates

#### 1.1 Update `KubernetesManifestComponentConfig`

**File**: `pkg/config/kubernetes_manifest_component.go`

```go
type KubernetesManifestComponentConfig struct {
    // Existing field - inline manifest (optional if kustomize or oci_artifact is set)
    Manifest      string  `mapstructure:"manifest,omitempty" features:"get,template"`
    
    // NEW: Kustomize configuration (optional)
    Kustomize     *KustomizeConfig `mapstructure:"kustomize,omitempty"`
    
    // NEW: Pre-packaged OCI artifact (optional)
    OCIArtifact   *OCIArtifactConfig `mapstructure:"oci_artifact,omitempty"`
    
    // Existing fields
    Namespace     string  `mapstructure:"namespace,omitempty" jsonschema:"required"`
    DriftSchedule *string `mapstructure:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`
}

// KustomizeConfig configures kustomize build options
type KustomizeConfig struct {
    // Path to kustomization directory (relative to source root)
    Path string `mapstructure:"path" jsonschema:"required"`
    
    // Additional patch files to apply after kustomize build
    Patches []string `mapstructure:"patches,omitempty"`
    
    // Enable Helm chart inflation during kustomize build
    EnableHelm bool `mapstructure:"enable_helm,omitempty"`
    
    // Load restrictor: none, rootOnly (default: rootOnly)
    LoadRestrictor string `mapstructure:"load_restrictor,omitempty"`
}

// OCIArtifactConfig references a pre-packaged OCI artifact
type OCIArtifactConfig struct {
    // OCI artifact URL (e.g., oci://ghcr.io/org/app-config)
    URL string `mapstructure:"url" jsonschema:"required"`
    
    // Tag to pull (e.g., v1.0.0, latest)
    Tag string `mapstructure:"tag,omitempty"`
    
    // Digest for immutable references (e.g., sha256:abc123...)
    Digest string `mapstructure:"digest,omitempty"`
    
    // Registry credentials reference (if private registry)
    CredentialsRef string `mapstructure:"credentials_ref,omitempty"`
}
```

#### 1.2 Update Validation

**File**: `pkg/config/kubernetes_manifest_component.go`

```go
func (t *KubernetesManifestComponentConfig) Validate() error {
    // Exactly one of manifest, kustomize, or oci_artifact must be set
    count := 0
    if t.Manifest != "" { count++ }
    if t.Kustomize != nil { count++ }
    if t.OCIArtifact != nil { count++ }
    
    if count == 0 {
        return errors.New("one of 'manifest', 'kustomize', or 'oci_artifact' must be specified")
    }
    if count > 1 {
        return errors.New("only one of 'manifest', 'kustomize', or 'oci_artifact' can be specified")
    }
    
    // Validate kustomize config
    if t.Kustomize != nil {
        if t.Kustomize.Path == "" {
            return errors.New("kustomize.path is required")
        }
    }
    
    // Validate OCI artifact config
    if t.OCIArtifact != nil {
        if t.OCIArtifact.URL == "" {
            return errors.New("oci_artifact.url is required")
        }
        if t.OCIArtifact.Tag == "" && t.OCIArtifact.Digest == "" {
            return errors.New("oci_artifact requires either tag or digest")
        }
    }
    
    return nil
}
```

---

### Phase 2: Deploy Plan Type Updates

#### 2.1 Update `KubernetesManifestDeployPlan`

**File**: `pkg/plans/types/kubernetes_manifest_deploy_plan.go`

With the unified pipeline, the deploy plan **always** contains an OCI artifact reference. The `Manifest` field is kept for runtime use (populated after unpacking the artifact in `Initialize()`).

```go
type KubernetesManifestDeployPlan struct {
    ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"`
    Namespace   string            `json:"namespace"`
    
    // Manifest content - populated at runtime after unpacking OCI artifact
    // Not serialized in the plan JSON, only used in-memory during execution
    Manifest    string            `json:"-"`
    
    // OCI artifact reference (ALWAYS set - all manifests flow through artifacts)
    OCIArtifact *OCIArtifactRef   `json:"oci_artifact"`
}

// OCIArtifactRef contains the resolved artifact reference
type OCIArtifactRef struct {
    // Full URL with registry (e.g., 123456789.dkr.ecr.us-west-2.amazonaws.com/nuon/manifests)
    URL     string `json:"url"`
    Tag     string `json:"tag,omitempty"`
    Digest  string `json:"digest,omitempty"`
    
    // Registry configuration for authentication
    RegistryType string `json:"registry_type"` // ecr, acr, docker
}
```

---

### Phase 3: Build Runner Implementation

#### 3.1 Create Kubernetes Manifest Build Handler

**New Directory**: `bins/runner/internal/jobs/build/kubernetes_manifest/`

**Files to create**:
- `handler.go` - Handler struct and FX constructor
- `meta.go` - Handler metadata (job type, name)
- `fetch.go` - Fetch source repository
- `initialize.go` - Initialize workspace and archive
- `validate.go` - Validate kustomization files exist
- `exec.go` - Build execution logic (using kustomizer SDK)
- `outputs.go` - Output handling
- `cleanup.go` - Cleanup temporary files
- `reset.go` - Reset handler state
- `state.go` - Handler state struct
- `kustomize.go` - Kustomize build helpers (adapted from kustomizer)

#### 3.2 Build Execution Logic (Using Kustomizer SDK)

**File**: `bins/runner/internal/jobs/build/kubernetes_manifest/exec.go`

This implementation follows the **kustomizer push artifact** pattern from `github.com/stefanprodan/kustomizer/pkg/registry`.

```go
package kubernetes_manifest

import (
    "context"
    "crypto/sha256"
    "fmt"
    "sort"
    "time"

    "github.com/fluxcd/pkg/ssa"
    "github.com/nuonco/nuon-runner-go/models"
    "github.com/stefanprodan/kustomizer/pkg/registry"
    "go.uber.org/zap"

    pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
    nuonRegistry "github.com/powertoolsdev/mono/bins/runner/internal/pkg/registry"
)

func (h *handler) Exec(ctx context.Context, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
    l, _ := pkgctx.Logger(ctx)
    
    plan := h.state.plan
    
    // Mode C: Pre-built OCI artifact - validate and mirror (skip kustomize build)
    if plan.OCIArtifact != nil {
        return h.execOCIArtifactMirror(ctx, l, job, jobExecution)
    }
    
    // Mode A & B: Both inline manifest and kustomize path go through kustomize build
    // Following kustomizer's buildManifests pattern
    return h.execKustomizeBuild(ctx, l, job, jobExecution)
}

// execKustomizeBuild handles both explicit kustomize paths and inline manifests
// This follows the kustomizer push artifact pattern:
// 1. Build manifests using kustomize API
// 2. Convert to sorted unstructured objects
// 3. Serialize to multi-doc YAML
// 4. Push to OCI registry with metadata
func (h *handler) execKustomizeBuild(ctx context.Context, l *zap.Logger, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
    
    // Step 1: Build manifests (following kustomizer's buildManifests pattern)
    // buildManifests returns []*unstructured.Unstructured objects
    var objects []*unstructured.Unstructured
    var err error
    
    if h.state.plan.Manifest != "" {
        // Mode A: Inline manifest - parse directly
        l.Info("parsing inline manifest")
        objects, err = ssa.ReadObjects(strings.NewReader(h.state.plan.Manifest))
        if err != nil {
            return fmt.Errorf("failed to parse inline manifest: %w", err)
        }
    } else {
        // Mode B: Kustomize path - build overlay
        l.Info("building kustomize overlay", zap.String("path", h.state.plan.Kustomize.Path))
        src := h.state.workspace.Source()
        kustomizePath := filepath.Join(src.AbsPath(), h.state.plan.Kustomize.Path)
        
        // Build kustomization (uses sigs.k8s.io/kustomize/api/krusty)
        objects, err = h.buildKustomization(ctx, kustomizePath)
        if err != nil {
            return fmt.Errorf("kustomize build failed: %w", err)
        }
        
        // Apply additional patches if specified
        if len(h.state.plan.Kustomize.Patches) > 0 {
            objects, err = h.applyPatches(ctx, objects, h.state.plan.Kustomize.Patches)
            if err != nil {
                return fmt.Errorf("failed to apply patches: %w", err)
            }
        }
    }
    
    // Step 2: Sort objects (following kustomizer pattern)
    // This ensures consistent ordering: CRDs first, then namespaces, then resources
    sort.Sort(ssa.SortableUnstructureds(objects))
    
    l.Info("built manifests", zap.Int("object_count", len(objects)))
    for _, obj := range objects {
        l.Debug("manifest object", zap.String("resource", ssa.FmtUnstructured(obj)))
    }
    
    // Step 3: Convert to multi-doc YAML (using fluxcd/pkg/ssa)
    yml, err := ssa.ObjectsToYAML(objects)
    if err != nil {
        return fmt.Errorf("failed to serialize manifests to YAML: %w", err)
    }
    
    // Step 4: Push to OCI registry using kustomizer's registry.Push
    // This creates an OCI artifact with the manifest YAML and metadata annotations
    artifactURL := h.buildArtifactURL()
    l.Info("pushing artifact to registry", 
        zap.String("url", artifactURL),
        zap.Int("yaml_size", len(yml)))
    
    meta := &registry.Metadata{
        Version:        h.state.buildVersion,
        Checksum:       fmt.Sprintf("%x", sha256.Sum256([]byte(yml))),
        Created:        time.Now().UTC().Format(time.RFC3339),
        SourceURL:      h.state.sourceURL,
        SourceRevision: h.state.sourceRevision,
    }
    
    digest, err := registry.Push(ctx, artifactURL, []byte(yml), meta, nil)
    if err != nil {
        return fmt.Errorf("failed to push artifact: %w", err)
    }
    
    l.Info("artifact pushed successfully", zap.String("digest", digest))
    
    // Step 5: Report result to API
    result := &models.CreateJobExecutionResultReq{
        OCIArtifact: &models.OCIArtifactResult{
            URL:    artifactURL,
            Digest: digest,
        },
    }
    _, err = h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, result)
    return err
}

// buildKustomization builds a kustomize overlay and returns unstructured objects
// This follows kustomizer's internal buildKustomization function
func (h *handler) buildKustomization(ctx context.Context, path string) ([]*unstructured.Unstructured, error) {
    // Use kustomize API (same as kustomizer uses internally)
    opts := krusty.MakeDefaultOptions()
    
    // Configure load restrictor
    if h.state.plan.Kustomize != nil {
        switch h.state.plan.Kustomize.LoadRestrictor {
        case "none":
            opts.LoadRestrictions = types.LoadRestrictionsNone
        default:
            opts.LoadRestrictions = types.LoadRestrictionsRootOnly
        }
        
        // Enable Helm chart inflation if requested
        if h.state.plan.Kustomize.EnableHelm {
            opts.PluginConfig.HelmConfig.Enabled = true
        }
    }
    
    k := krusty.MakeKustomizer(opts)
    resMap, err := k.Run(filesys.MakeFsOnDisk(), path)
    if err != nil {
        return nil, err
    }
    
    // Convert to YAML then parse as unstructured
    // (following kustomizer's pattern for consistent object handling)
    yml, err := resMap.AsYaml()
    if err != nil {
        return nil, err
    }
    
    return ssa.ReadObjects(strings.NewReader(string(yml)))
}

// applyPatches applies additional patches to the objects
// This follows kustomizer's patch application pattern
func (h *handler) applyPatches(ctx context.Context, objects []*unstructured.Unstructured, patches []string) ([]*unstructured.Unstructured, error) {
    // Convert objects back to YAML for patching
    yml, err := ssa.ObjectsToYAML(objects)
    if err != nil {
        return nil, err
    }
    
    // Create temporary directory with objects and patches
    tmpDir, err := os.MkdirTemp("", "kustomize-patch-*")
    if err != nil {
        return nil, err
    }
    defer os.RemoveAll(tmpDir)
    
    // Write base manifests
    basePath := filepath.Join(tmpDir, "base.yaml")
    if err := os.WriteFile(basePath, []byte(yml), 0644); err != nil {
        return nil, err
    }
    
    // Create kustomization with patches
    // Build patch references
    patchRefs := make([]string, len(patches))
    for i, p := range patches {
        patchRefs[i] = filepath.Join(h.state.workspace.Source().AbsPath(), p)
    }
    
    kustomization := fmt.Sprintf(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - base.yaml
patches:
%s
`, h.formatPatchRefs(patchRefs))
    
    kustomizationPath := filepath.Join(tmpDir, "kustomization.yaml")
    if err := os.WriteFile(kustomizationPath, []byte(kustomization), 0644); err != nil {
        return nil, err
    }
    
    // Build with patches
    return h.buildKustomization(ctx, tmpDir)
}

// execOCIArtifactMirror mirrors a pre-built OCI artifact to Nuon registry
// For BYOA (Bring Your Own Artifact) mode
func (h *handler) execOCIArtifactMirror(ctx context.Context, l *zap.Logger, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
    artifact := h.state.plan.OCIArtifact
    
    // Build source URL
    sourceURL := artifact.URL
    if artifact.Tag != "" {
        sourceURL = fmt.Sprintf("%s:%s", sourceURL, artifact.Tag)
    } else if artifact.Digest != "" {
        sourceURL = fmt.Sprintf("%s@%s", sourceURL, artifact.Digest)
    }
    
    l.Info("validating and mirroring OCI artifact", zap.String("source", sourceURL))
    
    // Pull from source to validate it exists and get content
    content, meta, err := registry.Pull(ctx, sourceURL, nil)
    if err != nil {
        return fmt.Errorf("failed to pull source artifact: %w", err)
    }
    
    l.Info("artifact validated", 
        zap.String("checksum", meta.Checksum),
        zap.String("created", meta.Created),
        zap.Int("content_size", len(content)))
    
    // Push to Nuon registry (mirror)
    destURL := h.buildArtifactURL()
    digest, err := registry.Push(ctx, destURL, []byte(content), meta, nil)
    if err != nil {
        return fmt.Errorf("failed to mirror artifact: %w", err)
    }
    
    l.Info("artifact mirrored successfully", 
        zap.String("destination", destURL),
        zap.String("digest", digest))
    
    // Report result
    result := &models.CreateJobExecutionResultReq{
        OCIArtifact: &models.OCIArtifactResult{
            URL:    destURL,
            Digest: digest,
        },
    }
    _, err = h.apiClient.CreateJobExecutionResult(ctx, job.ID, jobExecution.ID, result)
    return err
}
```

#### 3.3 Required Dependencies

**File**: `go.mod` - Add kustomizer and related dependencies:

```go
require (
    // Kustomizer SDK for OCI artifact operations
    github.com/stefanprodan/kustomizer v2.2.1+incompatible
    
    // FluxCD SSA for server-side apply and object handling
    github.com/fluxcd/pkg/ssa v0.22.0
    
    // Kustomize API for building overlays
    sigs.k8s.io/kustomize/api v0.12.1
    sigs.k8s.io/kustomize/kyaml v0.13.9
    
    // OCI/container registry operations (used by kustomizer)
    github.com/google/go-containerregistry v0.12.1
    
    // CLI utils for object metadata
    sigs.k8s.io/cli-utils v0.34.0
)
```

**Note**: The kustomizer module path is `github.com/stefanprodan/kustomizer`. Key packages:
- `github.com/stefanprodan/kustomizer/pkg/registry` - Push/Pull OCI artifacts
- `github.com/stefanprodan/kustomizer/pkg/inventory` - Inventory management (optional)

---

### Phase 4: Install Runner Updates (Using Kustomizer SDK)

#### 4.1 Update Deploy Handler Initialization

**File**: `bins/runner/internal/jobs/deploy/kubernetes_manifest/initialize.go`

With the unified OCI artifact pipeline, the install runner **always** pulls from an OCI artifact using kustomizer's `registry.Pull` function. This follows the **kustomizer pull artifact** pattern.

```go
package kubernetes_manifest

import (
    "context"
    "fmt"

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

#### 4.2 Diff Logic Compatibility - No Changes Required

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

#### 4.3 Update State to Include Kustomizer Types

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

#### 4.4 Update Exec to Use SSA ResourceManager (Optional Enhancement)

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

### Phase 5: Handler Registration

#### 5.1 Register Build Handler

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

## File Change Summary

### New Files

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
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/exec_ssa.go` | Optional SSA ResourceManager apply |

### Modified Files

| Path | Changes |
|------|---------|
| `pkg/config/kubernetes_manifest_component.go` | Add `Kustomize` and `OCIArtifact` config structs |
| `pkg/plans/types/kubernetes_manifest_deploy_plan.go` | Add `OCIArtifact` field, update Manifest to runtime-only |
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/initialize.go` | Use `registry.Pull()` for OCI artifact |
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/state.go` | Add `manifestObjects`, `artifactMetadata` fields |
| `bins/runner/internal/jobs/deploy/kubernetes_manifest/fetch.go` | Conditionally fetch artifact config |
| `bins/runner/internal/jobs/build/fx.go` | Register new build handler |
| `go.mod` | Add kustomizer SDK and FluxCD SSA dependencies |

### Key SDK Imports

```go
// Build Runner imports
import (
    "github.com/fluxcd/pkg/ssa"
    "github.com/stefanprodan/kustomizer/pkg/registry"
    "sigs.k8s.io/kustomize/api/krusty"
    "sigs.k8s.io/kustomize/api/types"
    "sigs.k8s.io/kustomize/kyaml/filesys"
)

// Install Runner imports
import (
    "github.com/fluxcd/pkg/ssa"
    "github.com/stefanprodan/kustomizer/pkg/registry"
)
```

---

## Configuration Examples

### Example 1: Inline Manifest (Existing - Unchanged)

```yaml
# nuon.yaml
components:
  - name: my-configmap
    type: kubernetes_manifest
    kubernetes_manifest:
      namespace: default
      manifest: |
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: my-config
        data:
          key: value
```

### Example 2: Kustomize Overlay (New)

```yaml
# nuon.yaml
components:
  - name: my-app
    type: kubernetes_manifest
    kubernetes_manifest:
      namespace: "{{.nuon.install.id}}"
      kustomize:
        path: "./k8s/overlays/production"
        patches:
          - "./k8s/patches/resource-limits.yaml"
        enable_helm: false
```

### Example 3: Pre-built OCI Artifact (New - BYOA)

```yaml
# nuon.yaml
components:
  - name: vendor-app
    type: kubernetes_manifest
    kubernetes_manifest:
      namespace: vendor
      oci_artifact:
        url: "oci://ghcr.io/vendor/app-manifests"
        tag: "v2.3.1"
        # OR use digest for immutability:
        # digest: "sha256:abc123..."
```

### Example 4: Private Registry OCI Artifact

```yaml
# nuon.yaml
components:
  - name: internal-app
    type: kubernetes_manifest
    kubernetes_manifest:
      namespace: internal
      oci_artifact:
        url: "oci://123456789.dkr.ecr.us-west-2.amazonaws.com/manifests"
        tag: "v1.0.0"
        credentials_ref: "aws-ecr-creds"  # References install-level credentials
```

---

## Testing Strategy

### Unit Tests

1. **Configuration Validation**
   - Test mutual exclusivity of manifest/kustomize/oci_artifact
   - Test required field validation
   - Test JSON schema generation

2. **Kustomize Build**
   - Test basic kustomization build
   - Test with patches
   - Test with Helm enabled
   - Test error handling for invalid kustomization

3. **OCI Artifact**
   - Test artifact URL parsing
   - Test tag vs digest handling
   - Test registry type detection

### Integration Tests

1. **Build Runner**
   - End-to-end kustomize build and push
   - OCI artifact mirroring
   - Error handling for missing sources

2. **Install Runner**
   - Deploy from OCI artifact
   - Deploy inline manifest (regression)
   - Plan creation and apply operations


---

## Migration Guide

### For Existing Users

No action required. Existing configurations using `manifest` field will continue to work unchanged.

### For New Kustomize Users

1. Create a `kustomization.yaml` in your repository
2. Update `nuon.yaml` to use the `kustomize` field instead of `manifest`
3. Run `nuon apps sync` to detect the new configuration
4. Deploy as usual

### For BYOA Users

1. Build your Kustomize overlay locally or in CI
2. Package into OCI artifact: `kustomizer push artifact oci://registry/repo:tag -k ./path`
3. Reference the artifact in `nuon.yaml` using `oci_artifact`
4. Deploy as usual

---

## Open Questions

1. **Kustomize Version**: Which kustomize version should we pin to? Recommend latest stable (v5.x).

2. **Helm Inflation**: Should we enable Helm chart inflation by default in kustomize builds? Recommend opt-in via `enable_helm: true`.

3. **Build Caching**: Should we cache kustomize build outputs? Recommend initial implementation without caching, add later based on demand.

4. **Secret Handling**: How should we handle encrypted secrets (SOPS) in kustomize overlays? Recommend deferring to Phase 2.

5. **Multi-Document Artifacts**: Should a single OCI artifact contain multiple kustomization results? Recommend single artifact = single rendered manifest.

---

## Timeline Estimate

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Configuration Schema | 2-3 days | None |
| Phase 2: Deploy Plan Types | 1 day | Phase 1 |
| Phase 3: Build Runner | 5-7 days | Phase 1, 2 |
| Phase 4: Install Runner | 3-4 days | Phase 2 |
| Phase 5: Registration & Testing | 3-4 days | Phase 3, 4 |
| **Total** | **14-19 days** | |

---

## References

### Kustomizer SDK Documentation

| Resource | Link | Description |
|----------|------|-------------|
| **Kustomizer Documentation** | [kustomizer.dev](https://kustomizer.dev/) | Official documentation site |
| **Kustomizer GitHub** | [github.com/stefanprodan/kustomizer](https://github.com/stefanprodan/kustomizer) | Source code and examples |
| **pkg/registry** | [registry package](https://github.com/stefanprodan/kustomizer/tree/main/pkg/registry) | OCI artifact Push/Pull functions |
| **pkg/inventory** | [inventory package](https://github.com/stefanprodan/kustomizer/tree/main/pkg/inventory) | Inventory tracking for garbage collection |
| **FluxCD SSA** | [pkg.go.dev/github.com/fluxcd/pkg/ssa](https://pkg.go.dev/github.com/fluxcd/pkg/ssa) | Server-side apply utilities |
| **Kustomize API** | [sigs.k8s.io/kustomize](https://github.com/kubernetes-sigs/kustomize) | Kustomize Go SDK |

### Kustomizer CLI Command Reference

| Command | Equivalent SDK Function | Description |
|---------|------------------------|-------------|
| `kustomizer push artifact` | `registry.Push(ctx, url, data, meta, recipients)` | Package and push OCI artifact |
| `kustomizer pull artifact` | `registry.Pull(ctx, url, identities)` | Pull and extract OCI artifact |
| `kustomizer apply inventory` | `ssa.ResourceManager.ApplyAllStaged()` | Apply with server-side apply |
| `kustomizer build` | `krusty.MakeKustomizer().Run()` | Build kustomize overlay |

### Existing Nuon Runner Patterns

- [Existing Terraform Build Handler](bins/runner/internal/jobs/build/terraform/exec.go)
- [Existing Helm Deploy Handler](bins/runner/internal/jobs/deploy/helm/)
- [Existing OCI Archive Pack/Unpack (ORAS)](bins/runner/internal/pkg/oci/archive/)
- [Existing Kubernetes Manifest Deploy Handler](bins/runner/internal/jobs/deploy/kubernetes_manifest/)

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

**Recommended: Option A** - Use kustomizer SDK directly for manifest artifacts. The terraform/helm handlers can continue using the existing ORAS infrastructure since they have different artifact formats.

### Dependency Versions

```go
// go.mod additions (align with kustomizer v2.2.1)
require (
    github.com/stefanprodan/kustomizer v2.2.1
    github.com/fluxcd/pkg/ssa v0.22.0
    sigs.k8s.io/kustomize/api v0.12.1
    sigs.k8s.io/kustomize/kyaml v0.13.9
    sigs.k8s.io/cli-utils v0.34.0
    github.com/google/go-containerregistry v0.12.1
)
```

**Note**: Check for version conflicts with existing dependencies. The kustomizer SDK uses `go-containerregistry` which may have different version requirements than the existing ORAS-based implementation.
