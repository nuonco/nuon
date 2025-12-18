# Kustomize Support: ctl-api Workflow Changes

This document covers the required changes in `services/ctl-api/` to support the Kubernetes manifest build pipeline and OCI sync workflow.

> **Key Decisions**:
> - Always use `RunnerJobTypeKubernetesManifestBuild` for all k8s manifest components
> - **Namespace-only templating**: Manifest content not rendered (breaking change), namespace still rendered at deploy time
> - OCIArtifact (BYOA) deferred to post-MVP

---

## Overview

Currently, `KubernetesManifest` components use `RunnerJobTypeNOOPBuild` and `RunnerJobTypeNOOPSync`, meaning no actual build or sync work happens. The manifest string is passed directly in the deploy plan.

To support the unified OCI artifact pipeline, we need to:
1. Add a new build job type for kubernetes manifest builds
2. Change sync job type from NOOP to OCI sync
3. Create build plan generator for kubernetes manifests
4. Update deploy plan to reference OCI artifacts (remove manifest template rendering, keep namespace rendering)

---

## Phase 1: Runner Job Type Changes

### 1.1 Add New Build Job Type

**File**: `services/ctl-api/internal/app/runner_job.go`

Add new job type constant:

```go
const (
    // build job types
    RunnerJobTypeDockerBuild              RunnerJobType = "docker-build"
    RunnerJobTypeContainerImageBuild      RunnerJobType = "container-image-build"
    RunnerJobTypeTerraformModuleBuild     RunnerJobType = "terraform-module-build"
    RunnerJobTypeHelmChartBuild           RunnerJobType = "helm-chart-build"
    RunnerJobTypeKubernetesManifestBuild  RunnerJobType = "kubernetes-manifest-build"  // NEW
    RunnerJobTypeNOOPBuild                RunnerJobType = "noop-build"
    // ...
)
```

Update `Group()` method to include new type:

```go
func (r RunnerJobType) Group() RunnerJobGroup {
    switch r {
    // builds
    case RunnerJobTypeDockerBuild,
        RunnerJobTypeContainerImageBuild,
        RunnerJobTypeNOOPBuild,
        RunnerJobTypeTerraformModuleBuild,
        RunnerJobTypeHelmChartBuild,
        RunnerJobTypeKubernetesManifestBuild:  // NEW
        return RunnerJobGroupBuild
    // ...
    }
}
```

### 1.2 Update Component Type Mappings

**File**: `services/ctl-api/internal/app/component.go`

Update `BuildJobType()`:

```go
func (c ComponentType) BuildJobType() RunnerJobType {
    switch c {
    case ComponentTypeTerraformModule:
        return RunnerJobTypeTerraformModuleBuild
    case ComponentTypeHelmChart:
        return RunnerJobTypeHelmChartBuild
    case ComponentTypeDockerBuild:
        return RunnerJobTypeDockerBuild
    case ComponentTypeExternalImage:
        return RunnerJobTypeContainerImageBuild
    case ComponentTypeKubernetesManifest:
        return RunnerJobTypeKubernetesManifestBuild  // CHANGED from RunnerJobTypeNOOPBuild
    case ComponentTypeJob:
        return RunnerJobTypeNOOPBuild
    default:
    }
    return RunnerJobTypeUnknown
}
```

Update `SyncJobType()`:

```go
func (c ComponentType) SyncJobType() RunnerJobType {
    switch c {
    case ComponentTypeTerraformModule,
        ComponentTypeDockerBuild,
        ComponentTypeExternalImage,
        ComponentTypeHelmChart,
        ComponentTypeKubernetesManifest:  // ADDED - was previously in NOOP case
        return RunnerJobTypeOCISync

    case ComponentTypeJob:
        return RunnerJobTypeNOOPSync
    default:
    }
    return RunnerJobTypeUnknown
}
```

---

## Phase 2: Build Plan Types

### 2.1 Add Kubernetes Manifest Build Plan Type

**File**: `pkg/plans/types/kubernetes_manifest_build_plan.go` (NEW FILE)

```go
package plantypes

// KubernetesManifestBuildPlan contains build configuration for kubernetes manifest components.
// This is used by the build runner to package manifests into OCI artifacts.
type KubernetesManifestBuildPlan struct {
    // Labels for the OCI artifact
    Labels map[string]string `json:"labels,omitempty"`

    // SourceType indicates how manifests are sourced: "inline", "kustomize", or "oci_artifact"
    SourceType string `json:"source_type"`

    // InlineManifest contains the raw manifest YAML (for inline source type)
    // This is the rendered manifest after template processing
    InlineManifest string `json:"inline_manifest,omitempty"`

    // KustomizePath is the path to the kustomization directory (for kustomize source type)
    // Relative to the repository root
    KustomizePath string `json:"kustomize_path,omitempty"`

    // KustomizeConfig contains additional kustomize build options
    KustomizeConfig *KustomizeBuildConfig `json:"kustomize_config,omitempty"`

    // OCIArtifactSource contains the source artifact reference (for BYOA)
    OCIArtifactSource *OCIArtifactSource `json:"oci_artifact_source,omitempty"`
}

// KustomizeBuildConfig contains kustomize-specific build options
type KustomizeBuildConfig struct {
    // Patches are additional patch files to apply after kustomize build
    Patches []string `json:"patches,omitempty"`

    // EnableHelm enables Helm chart inflation during kustomize build
    EnableHelm bool `json:"enable_helm,omitempty"`

    // LoadRestrictor controls file loading: "none" or "rootOnly" (default)
    LoadRestrictor string `json:"load_restrictor,omitempty"`
}

// OCIArtifactSource references a pre-built OCI artifact (BYOA)
type OCIArtifactSource struct {
    URL            string `json:"url"`
    Tag            string `json:"tag,omitempty"`
    Digest         string `json:"digest,omitempty"`
    CredentialsRef string `json:"credentials_ref,omitempty"`
}
```

### 2.2 Update BuildPlan to Include New Type

**File**: `pkg/plans/types/build_plan.go`

```go
type BuildPlan struct {
    ComponentID      string `json:"component_id"`
    ComponentBuildID string `json:"component_build_id"`

    Src *GitSource `json:"git_source"`

    Dst    *configs.OCIRegistryRepository `json:"dst_registry" validate:"required"`
    DstTag string                         `json:"dst_tag" validate:"required"`

    HelmBuildPlan               *HelmBuildPlan               `json:"helm_build_plan,omitempty"`
    TerraformBuildPlan          *TerraformBuildPlan          `json:"terraform_build_plan,omitempty"`
    DockerBuildPlan             *DockerBuildPlan             `json:"docker_build_plan,omitempty"`
    ContainerImagePullPlan      *ContainerImagePullPlan      `json:"container_image_pull_plan,omitempty"`
    KubernetesManifestBuildPlan *KubernetesManifestBuildPlan `json:"kubernetes_manifest_build_plan,omitempty"`  // NEW

    MinSandboxMode
}
```

---

## Phase 3: Build Plan Creation

### 3.1 Create Kubernetes Manifest Build Plan Generator

**File**: `services/ctl-api/internal/app/components/worker/plan/plan_kubernetes_manifest_build.go` (NEW FILE)

```go
package plan

import (
    "go.temporal.io/sdk/workflow"

    "github.com/pkg/errors"

    plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
    "github.com/powertoolsdev/mono/pkg/render"
    "github.com/powertoolsdev/mono/services/ctl-api/internal/app"
    "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/worker/activities"
    "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

func (p *Planner) createKubernetesManifestBuildPlan(ctx workflow.Context, bld *app.ComponentBuild) (*plantypes.KubernetesManifestBuildPlan, error) {
    l, err := log.WorkflowLogger(ctx)
    if err != nil {
        return nil, err
    }

    cfg := bld.ComponentConfigConnection.KubernetesManifestComponentConfig
    if cfg == nil {
        return nil, errors.New("kubernetes manifest component config is nil")
    }

    plan := &plantypes.KubernetesManifestBuildPlan{
        Labels: map[string]string{
            "component_id":       bld.ComponentID,
            "component_build_id": bld.ID,
        },
    }

    // Determine source type based on config
    switch {
    case cfg.Kustomize != nil:
        // Kustomize source
        l.Info("generating kustomize build plan")
        plan.SourceType = "kustomize"
        plan.KustomizePath = cfg.Kustomize.Path
        plan.KustomizeConfig = &plantypes.KustomizeBuildConfig{
            Patches:        cfg.Kustomize.Patches,
            EnableHelm:     cfg.Kustomize.EnableHelm,
            LoadRestrictor: cfg.Kustomize.LoadRestrictor,
        }

    // NOTE: OCIArtifact (BYOA) case deferred to post-MVP

    default:
        // Inline manifest source
        l.Info("generating inline manifest build plan")
        plan.SourceType = "inline"
        plan.InlineManifest = cfg.Manifest
    }

    return plan, nil
}
```

### 3.2 Update Main Build Plan Creator

**File**: `services/ctl-api/internal/app/components/worker/plan/plan_component_build.go`

Add case for kubernetes manifest:

```go
func (p *Planner) createComponentBuildPlan(ctx workflow.Context, req *CreateComponentBuildPlanRequest) (*plantypes.BuildPlan, error) {
    // ... existing code ...

    switch build.ComponentConfigConnection.Type {
    case app.ComponentTypeDockerBuild:
        // ... existing ...

    case app.ComponentTypeExternalImage:
        // ... existing ...

    case app.ComponentTypeTerraformModule:
        // ... existing ...

    case app.ComponentTypeHelmChart:
        // ... existing ...

    case app.ComponentTypeKubernetesManifest:  // NEW CASE
        l.Info("generating kubernetes manifest build plan")
        k8sManifestPlan, err := p.createKubernetesManifestBuildPlan(ctx, build)
        if err != nil {
            return nil, errors.Wrap(err, "unable to create kubernetes manifest build plan")
        }
        plan.KubernetesManifestBuildPlan = k8sManifestPlan
    }

    // ... rest of existing code ...
}
```

---

## Phase 4: Deploy Plan Updates

### 4.1 Update Kubernetes Manifest Deploy Plan Type

**File**: `pkg/plans/types/kubernetes_manifest_deploy_plan.go`

```go
package plantypes

import (
    "github.com/powertoolsdev/mono/pkg/kube"
)

type KubernetesManifestDeployPlan struct {
    ClusterInfo *kube.ClusterInfo `json:"cluster_info,block"`

    Namespace string `json:"namespace"`

    // Manifest is populated at runtime from the OCI artifact
    // This field is no longer set during plan creation - it's populated by the runner
    // after pulling the OCI artifact during Initialize()
    Manifest string `json:"manifest,omitempty"`

    // NEW: OCI artifact reference (set during plan creation)
    OCIArtifact *OCIArtifactReference `json:"oci_artifact,omitempty"`
}

// OCIArtifactReference points to the packaged manifest artifact
type OCIArtifactReference struct {
    // URL is the full artifact URL (e.g., registry.nuon.co/org_id/app_id)
    URL string `json:"url"`

    // Tag is the artifact tag (typically the build ID)
    Tag string `json:"tag,omitempty"`

    // Digest is the immutable artifact digest (e.g., sha256:abc123...)
    Digest string `json:"digest,omitempty"`
}
```

### 4.2 Update Deploy Plan Creator

**File**: `services/ctl-api/internal/app/installs/worker/plan/plan_kubernetes_manifest_deploy.go`

```go
func (p *Planner) createKubernetesManifestDeployPlan(ctx workflow.Context, req *CreateDeployPlanRequest) (*plantypes.KubernetesManifestDeployPlan, error) {
    l, err := log.WorkflowLogger(ctx)
    if err != nil {
        return nil, err
    }

    // ... existing code to get install, stack, state, stateData, etc. ...

    cfg := compBuild.ComponentConfigConnection.KubernetesManifestComponentConfig

    // Namespace STILL supports template rendering (preserved behavior)
    namespace := cfg.Namespace
    renderedNamespace, err := render.RenderV2(namespace, stateData)
    if err != nil {
        l.Error("error rendering namespace",
            zap.String("namespace", namespace),
            zap.Error(err))
        return nil, errors.Wrap(err, "unable to render namespace")
    }

    clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state)
    if err != nil {
        return nil, errors.Wrap(err, "unable to get cluster info")
    }

    // Get OCI artifact reference from the synced artifact
    // The artifact was synced to the install registry during execSync
    ociArtifact, err := activities.AwaitGetOCIArtifactByOwnerID(ctx, req.InstallDeployID)
    if err != nil {
        return nil, errors.Wrap(err, "unable to get OCI artifact reference")
    }

    return &plantypes.KubernetesManifestDeployPlan{
        ClusterInfo: clusterInfo,
        Namespace:   renderedNamespace,  // Namespace templating preserved
        // Manifest is no longer set here - runner will pull from artifact
        OCIArtifact: &plantypes.OCIArtifactReference{
            URL:    ociArtifact.Repository,
            Tag:    ociArtifact.Tag,
            Digest: ociArtifact.Digest,
        },
    }, nil
}
```

---

## Phase 5: Sync Workflow Integration

### 5.1 How OCI Sync Already Works

The existing `execSync` workflow in `shared_execute_sync.go` already handles OCI artifact sync. When `SyncJobType()` returns `RunnerJobTypeOCISync`, the workflow:

1. Creates a sync job with the appropriate type
2. Creates a sync plan via `plan.AwaitCreateSyncPlan()`
3. Executes the job on the runner
4. Stores the resulting OCI artifact reference

**No additional changes needed** in the sync workflow - changing `SyncJobType()` to return `RunnerJobTypeOCISync` for `ComponentTypeKubernetesManifest` is sufficient.

### 5.2 Sync Plan Already Supports This

The existing sync plan (`container_image_sync_plan.go` and related) handles copying artifacts from org registry to install registry. The kubernetes manifest OCI artifact will flow through the same path.

---

## File Change Summary

### New Files

| Path | Description |
|------|-------------|
| `pkg/plans/types/kubernetes_manifest_build_plan.go` | Build plan type for kubernetes manifest |
| `services/ctl-api/internal/app/components/worker/plan/plan_kubernetes_manifest_build.go` | Build plan generator |

### Modified Files

| Path | Changes |
|------|---------|
| `services/ctl-api/internal/app/runner_job.go` | Add `RunnerJobTypeKubernetesManifestBuild` |
| `services/ctl-api/internal/app/component.go` | Update `BuildJobType()` and `SyncJobType()` |
| `pkg/plans/types/build_plan.go` | Add `KubernetesManifestBuildPlan` field |
| `pkg/plans/types/kubernetes_manifest_deploy_plan.go` | Add `OCIArtifact` field |
| `services/ctl-api/internal/app/components/worker/plan/plan_component_build.go` | Add kubernetes manifest case |
| `services/ctl-api/internal/app/installs/worker/plan/plan_kubernetes_manifest_deploy.go` | Use OCI artifact instead of inline manifest, keep namespace rendering |

---

## Sequence Diagram: Complete Flow

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                     KUBERNETES MANIFEST BUILD → DEPLOY FLOW                      │
└─────────────────────────────────────────────────────────────────────────────────┘

  Config Sync                   Build Phase                         Deploy Phase
  ───────────                   ───────────                         ────────────

  1. nuon apps sync
       │
       ▼
  2. ComponentConfig created
     (manifest/kustomize config stored)
       │
       ▼
  3. ComponentBuild created ────►  4. plan_component_build.go
                                      - calls createKubernetesManifestBuildPlan()
                                      - determines source type (inline/kustomize/byoa)
                                        │
                                        ▼
                                   5. Build Runner executes
                                      - RunnerJobTypeKubernetesManifestBuild
                                      - Fetches source, runs kustomize, pushes OCI
                                        │
                                        ▼
                                   6. Artifact stored in Org Registry
                                      (nuon.ecr.aws/org_id/app_id:build_id)


  7. Deploy triggered ──────────────────────────────────────────────►

       8. execSync()
          - RunnerJobTypeOCISync
          - Copies artifact from Org Registry → Install Registry
            │
            ▼
       9. plan_kubernetes_manifest_deploy.go
          - Renders namespace with install state data
          - Gets OCI artifact reference
          - Creates DeployPlan with OCIArtifact + rendered namespace
            │
            ▼
      10. Install Runner executes
          - RunnerJobTypeKubernetesManifestDeploy
          - Initialize(): registry.Pull(artifact)
          - Exec(): applies to K8s cluster
```

---

## Testing Strategy

### Unit Tests

1. **Component Type Mapping Tests**
   - Verify `BuildJobType()` returns `RunnerJobTypeKubernetesManifestBuild`
   - Verify `SyncJobType()` returns `RunnerJobTypeOCISync`

2. **Build Plan Creation Tests**
   - Test inline manifest → correct SourceType
   - Test kustomize config → correct Path and options
   - Test OCI artifact config → correct Source reference

3. **Deploy Plan Creation Tests**
   - Test OCIArtifact reference is populated
   - Test Manifest field is empty (populated by runner)

### Integration Tests

1. **End-to-End Build Flow**
   - Create kubernetes manifest component
   - Trigger build
   - Verify OCI artifact is pushed to org registry

2. **End-to-End Deploy Flow**
   - Deploy to install
   - Verify sync copies artifact
   - Verify runner pulls and applies

---

## Backwards Compatibility

### Migration Strategy

Existing inline manifest components will continue to work:

1. **Database**: No schema changes for existing `manifest` column
2. **Build Phase**: Inline manifests are packaged into OCI artifacts (new behavior)
3. **Deploy Phase**: Runner pulls from OCI instead of receiving inline manifest

### Rollback Plan

If issues arise, revert by:
1. Changing `BuildJobType()` back to `RunnerJobTypeNOOPBuild`
2. Changing `SyncJobType()` back to `RunnerJobTypeNOOPSync`
3. Reverting deploy plan to pass `Manifest` directly

This allows rapid rollback without database migrations.
