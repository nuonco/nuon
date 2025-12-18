# CLI Sync Changes for Kustomize Support

## Overview

The CLI's `nuon apps sync` command uses the `pkg/config/sync` package to sync component configurations to the API. This document details the changes required to support Kustomize configurations.

## Current State

**File**: `pkg/config/sync/components_kubernetes_manifest.go`

The current implementation only sends `Manifest` and `Namespace` fields:

```go
configRequest := &models.ServiceCreateKubernetesManifestComponentConfigRequest{
    AppConfigID:  s.appConfigID,
    Dependencies: comp.Dependencies,
    Checksum:     comp.Checksum,

    Namespace: comp.KubernetesManifest.Namespace,
    Manifest:  comp.KubernetesManifest.Manifest,
}
```

**Missing**: The `Kustomize` field is not mapped from the parsed config to the API request.

---

## Dependencies

### 1. nuon-go Client Library Update (External)

The `github.com/nuonco/nuon-go` client library needs to be updated to include the `Kustomize` field in the request model.

**Required model changes in nuon-go**:

```go
// ServiceCreateKubernetesManifestComponentConfigRequest
type ServiceCreateKubernetesManifestComponentConfigRequest struct {
    AppConfigID   string   `json:"app_config_id"`
    References    []string `json:"references"`
    Checksum      string   `json:"checksum"`
    Dependencies  []string `json:"dependencies"`
    
    // Existing fields
    Manifest      string   `json:"manifest,omitempty"`
    Namespace     string   `json:"namespace"`
    DriftSchedule string   `json:"drift_schedule,omitempty"`
    
    // NEW: Kustomize configuration
    Kustomize *KustomizeConfigRequest `json:"kustomize,omitempty"`
}

// NEW: KustomizeConfigRequest
type KustomizeConfigRequest struct {
    Path           string   `json:"path"`
    Patches        []string `json:"patches,omitempty"`
    EnableHelm     bool     `json:"enable_helm,omitempty"`
    LoadRestrictor string   `json:"load_restrictor,omitempty"`
}
```

**Action**: Update nuon-go and bump the dependency in `go.mod`.

---

## Implementation

### File Changes

| File | Changes | Status |
|------|---------|--------|
| `pkg/config/sync/components_kubernetes_manifest.go` | Map Kustomize config to API request | ⏳ TODO |
| `go.mod` | Bump nuon-go dependency (after nuon-go update) | ⏳ TODO |

### Code Changes

**File**: `pkg/config/sync/components_kubernetes_manifest.go`

```go
func (s *sync) createKubernetesManifestComponentConfig(
    ctx context.Context, resource, compID string, comp *config.Component,
) (string, string, error) {
    _ = comp.KubernetesManifest

    configRequest := &models.ServiceCreateKubernetesManifestComponentConfigRequest{
        AppConfigID:  s.appConfigID,
        Dependencies: comp.Dependencies,
        Checksum:     comp.Checksum,

        Namespace: comp.KubernetesManifest.Namespace,
        Manifest:  comp.KubernetesManifest.Manifest,
    }

    // NEW: Map Kustomize configuration
    if comp.KubernetesManifest.Kustomize != nil {
        configRequest.Kustomize = &models.KustomizeConfigRequest{
            Path:           comp.KubernetesManifest.Kustomize.Path,
            Patches:        comp.KubernetesManifest.Kustomize.Patches,
            EnableHelm:     comp.KubernetesManifest.Kustomize.EnableHelm,
            LoadRestrictor: comp.KubernetesManifest.Kustomize.LoadRestrictor,
        }
    }

    if comp.KubernetesManifest.DriftSchedule != nil {
        configRequest.DriftSchedule = *comp.KubernetesManifest.DriftSchedule
    }

    // ... rest of function unchanged
}
```

---

## Data Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         CLI SYNC FLOW                                    │
└─────────────────────────────────────────────────────────────────────────┘

  TOML Config                pkg/config                 pkg/config/sync
  ──────────                 ──────────                 ───────────────

  [component.my-app]         KubernetesManifest         API Request
  type = "kubernetes_manifest"    ComponentConfig
                                      │
  [component.my-app.kustomize]        │
  path = "./k8s/overlays/prod"   ─────┼─────►  Kustomize: {
  patches = ["./patches/mem.yaml"]    │          Path: "./k8s/overlays/prod"
  enable_helm = false                 │          Patches: ["./patches/mem.yaml"]
                                      │          EnableHelm: false
                                      ▼        }
                             ┌────────────────┐
                             │ nuon-go client │
                             │    models      │
                             └───────┬────────┘
                                     │
                                     ▼
                              POST /v1/apps/{app_id}/components/{id}/configs/kubernetes-manifest
```

---

## Testing

### Manual Testing

1. Create a test app with kustomize config:

```toml
# nuon.toml
[app]
name = "kustomize-test"

[component.my-manifests]
type = "kubernetes_manifest"
namespace = "default"

[component.my-manifests.kustomize]
path = "./k8s/overlays/production"
patches = ["./k8s/patches/limits.yaml"]
enable_helm = false
```

2. Run sync:
```bash
nuon apps sync ./
```

3. Verify the API received the kustomize configuration by checking the component config in the dashboard or via API.

### Unit Tests

Add test cases to verify:
- Kustomize config is correctly mapped to API request
- Inline manifest continues to work (backwards compatibility)
- Mutual exclusivity validation (manifest OR kustomize, not both)

---

## Rollout Order

1. **nuon-go**: Add `Kustomize` field to request models
2. **mono**: Bump nuon-go dependency
3. **mono**: Update `components_kubernetes_manifest.go` sync logic
4. **Release**: New CLI version with kustomize support

---

## Related Documents

- [02-config-schema.md](./02-config-schema.md) - Config struct definitions (already done)
- [06-api-data-model.md](./06-api-data-model.md) - API request/response models (already done)
- [05-examples-migration.md](./05-examples-migration.md) - User-facing config examples
