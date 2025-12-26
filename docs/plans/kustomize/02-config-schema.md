# Kustomize Support: Configuration Schema

This document covers the configuration schema changes needed in `pkg/config/` and `pkg/plans/types/`.

> **Key Decisions**:
> - Flat config schema: `Manifest` and `Kustomize` are sibling fields
> - **Namespace-only templating**: `Namespace` supports template variables, `Manifest` content does not (breaking change)
> - OCIArtifact (BYOA) deferred to post-MVP

---

## Phase 1: Configuration Schema Updates

### 1.1 Update `KubernetesManifestComponentConfig`

**File**: `pkg/config/kubernetes_manifest_component.go`

```go
type KubernetesManifestComponentConfig struct {
    // Existing field - inline manifest (optional if kustomize is set)
    // NOTE: Template variables (e.g., {{.nuon.install.id}}) are NO LONGER SUPPORTED in manifest content
    Manifest      string  `mapstructure:"manifest,omitempty" features:"get"`
    
    // NEW: Kustomize configuration (mutually exclusive with Manifest)
    Kustomize     *KustomizeConfig `mapstructure:"kustomize,omitempty"`
    
    // Existing fields
    // NOTE: Template variables ARE SUPPORTED in Namespace (rendered at deploy time)
    Namespace     string  `mapstructure:"namespace,omitempty" jsonschema:"required"`
    DriftSchedule *string `mapstructure:"drift_schedule,omitempty" nuonhash:"omitempty"`
}

// NOTE: OCIArtifactConfig is deferred to post-MVP. 
// See 07-future-byoa-support.md for planned BYOA support.

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
```

### 1.2 Update Validation

**File**: `pkg/config/kubernetes_manifest_component.go`

```go
func (t *KubernetesManifestComponentConfig) Validate() error {
    // Exactly one of manifest or kustomize must be set
    count := 0
    if t.Manifest != "" { count++ }
    if t.Kustomize != nil { count++ }
    
    if count == 0 {
        return errors.New("one of 'manifest' or 'kustomize' must be specified")
    }
    if count > 1 {
        return errors.New("only one of 'manifest' or 'kustomize' can be specified")
    }
    
    // Validate kustomize config
    if t.Kustomize != nil {
        if t.Kustomize.Path == "" {
            return errors.New("kustomize.path is required")
        }
    }
    
    return nil
}
```

---

## Phase 2: Deploy Plan Type Updates

### 2.1 Update `KubernetesManifestDeployPlan`

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

## Key SDK Imports

```go
// Build Runner imports (kustomize build + ORAS push)
import (
    "sigs.k8s.io/kustomize/api/krusty"
    "sigs.k8s.io/kustomize/api/types"
    "sigs.k8s.io/kustomize/kyaml/filesys"
    "github.com/fluxcd/pkg/ssa"  // for object parsing/serialization
    
    // Existing ORAS infrastructure for OCI push (no new registry dependencies)
    "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci"
)

// Install Runner imports (existing archive pattern)
import (
    // Uses existing archive infrastructure like Helm handler
    "github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci/archive"
)
```

---

## Dependency Versions

```go
// go.mod additions
require (
    github.com/fluxcd/pkg/ssa v0.22.0
    sigs.k8s.io/kustomize/api v0.12.1
    sigs.k8s.io/kustomize/kyaml v0.13.9
)
```

**Note**: No `github.com/stefanprodan/kustomizer` dependency needed - we use the kustomize SDK directly for build and existing ORAS infrastructure for OCI push.

---

## Testing Strategy

### Unit Tests

1. **Configuration Validation**
   - Test mutual exclusivity of manifest/kustomize
   - Test required field validation
   - Test JSON schema generation
