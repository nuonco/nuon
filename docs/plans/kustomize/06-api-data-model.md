# Kustomize Support: API & Data Model Changes

This document covers the API layer and database schema changes needed in `services/ctl-api/` to support Kustomize configurations.

> **Key Decisions**:
> - OCIArtifact (BYOA) stripped from MVP - only Kustomize fields added
> - No template rendering for Manifest or Namespace (breaking change)
> - Flat config schema: Manifest and Kustomize as siblings

---

## Overview

The Kustomize feature requires changes across three layers:
1. **Database Schema** - New JSONB field for `KustomizeConfig` in `KubernetesManifestComponentConfig`
2. **API Request/Response Models** - Updated create/read endpoints in components service
3. **TOML Config Schema** - Updated `pkg/config/` structs (already covered in [02-config-schema.md](./02-config-schema.md))

---

## Phase 1: Database Schema Changes

### 1.1 Update `KubernetesManifestComponentConfig` Model

**File**: `services/ctl-api/internal/app/kubernetes_manifest_component_config.go`

The existing model has `Manifest` and `Namespace` fields which remain fully supported. We add a new JSONB column for extended configurations (Kustomize and OCI artifact) while keeping the existing columns unchanged.

**Design Principle**: 
- `manifest` and `namespace` remain primary columns — inline manifests continue to use them directly
- New `kustomize_config` JSONB column stores Kustomize/OCI artifact configurations only
- Source type is determined by which fields are populated (not a separate enum)

```go
package app

import (
    "database/sql/driver"
    "encoding/json"
    "time"

    "github.com/pkg/errors"
    "gorm.io/gorm"
    "gorm.io/plugin/soft_delete"

    "github.com/powertoolsdev/mono/pkg/shortid/domains"
    "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
    "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type KubernetesManifestComponentConfig struct {
    ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero"`
    CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null"`
    CreatedBy   Account               `json:"-"`
    CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull"`
    UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull"`
    DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-"`

    // RLS
    OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true"`
    Org   Org    `json:"-" faker:"-"`

    // Parent reference
    ComponentConfigConnectionID string                    `json:"component_config_connection_id,omitzero" gorm:"notnull"`
    ComponentConfigConnection   ComponentConfigConnection `json:"-"`

    // Primary fields - used for inline manifests (fully supported)
    Manifest  string `json:"manifest,omitzero" gorm:"not null;default:''"`
    Namespace string `json:"namespace,omitzero" gorm:"not null;default:default"`

    // Kustomize configuration (mutually exclusive with Manifest)
    Kustomize *KustomizeConfig `json:"kustomize,omitzero" gorm:"type:jsonb"`
}

// KustomizeConfig defines kustomize build options
type KustomizeConfig struct {
    // Path to kustomization directory (relative to source root)
    Path string `json:"path"`

    // Additional patch files to apply after kustomize build
    Patches []string `json:"patches,omitempty"`

    // Enable Helm chart inflation during kustomize build
    EnableHelm bool `json:"enable_helm,omitempty"`

    // Load restrictor: "none" or "rootOnly" (default: "rootOnly")
    LoadRestrictor string `json:"load_restrictor,omitempty"`
}

// Scan implements the database/sql.Scanner interface
func (c *KustomizeConfig) Scan(v interface{}) (err error) {
    switch v := v.(type) {
    case nil:
        return nil
    case []byte:
        if err := json.Unmarshal(v, c); err != nil {
            return errors.Wrap(err, "unable to scan kustomize config")
        }
    }
    return
}

// Value implements the driver.Valuer interface
func (c *KustomizeConfig) Value() (driver.Value, error) {
    return json.Marshal(c)
}

func (KustomizeConfig) GormDataType() string {
    return "jsonb"
}

// SourceType returns the source type based on which fields are populated
func (k *KubernetesManifestComponentConfig) SourceType() string {
    if k.Kustomize != nil {
        return "kustomize"
    }
    return "inline"
}

// Existing methods unchanged...
func (k *KubernetesManifestComponentConfig) Indexes(db *gorm.DB) []migrations.Index {
    return []migrations.Index{
        {
            Name: indexes.Name(db, &KubernetesManifestComponentConfig{}, "org_id"),
            Columns: []string{"org_id"},
        },
    }
}

func (e *KubernetesManifestComponentConfig) BeforeCreate(tx *gorm.DB) error {
    e.ID = domains.NewComponentID()
    e.CreatedByID = createdByIDFromContext(tx.Statement.Context)
    e.OrgID = orgIDFromContext(tx.Statement.Context)
    return nil
}
```

### 1.2 Database Migration

Since we're adding a new JSONB column, a migration is needed. GORM's auto-migration should handle this, but for explicit control:

```sql
-- Migration: Add kustomize_config JSONB column
ALTER TABLE kubernetes_manifest_component_configs 
ADD COLUMN IF NOT EXISTS kustomize_config JSONB;
```

**Note**: No data migration is needed for existing inline manifests — they continue to use the `manifest` and `namespace` columns directly. The new `kustomize_config` column is only populated for Kustomize and OCI artifact configurations.

---

## Phase 2: API Request/Response Model Updates

### 2.1 Update Create Request

**File**: `services/ctl-api/internal/app/components/service/create_kubernetes_manifest_component_config.go`

```go
// CreateKubernetesManifestComponentConfigRequest represents the API request
type CreateKubernetesManifestComponentConfigRequest struct {
    AppConfigID string `json:"app_config_id"`

    References   []string `json:"references"`
    Checksum     string   `json:"checksum"`
    Dependencies []string `json:"dependencies"`

    // Inline manifest (mutually exclusive with Kustomize)
    // NOTE: Template variables are NO LONGER SUPPORTED
    Manifest      string  `json:"manifest,omitempty"`
    Namespace     string  `json:"namespace"`
    DriftSchedule *string `json:"drift_schedule,omitempty"`

    // NEW: Kustomize configuration (mutually exclusive with Manifest)
    Kustomize   *KustomizeConfigRequest   `json:"kustomize,omitempty"`
    
    // NOTE: OCIArtifact (BYOA) deferred to post-MVP
}

// KustomizeConfigRequest defines kustomize options in API requests
type KustomizeConfigRequest struct {
    Path           string   `json:"path" validate:"required"`
    Patches        []string `json:"patches,omitempty"`
    EnableHelm     bool     `json:"enable_helm,omitempty"`
    LoadRestrictor string   `json:"load_restrictor,omitempty"`
}

// Validate ensures exactly one source type is specified
func (c *CreateKubernetesManifestComponentConfigRequest) Validate(v *validator.Validate) error {
    if err := v.Struct(c); err != nil {
        return validatorPkg.FormatValidationError(err)
    }

    // Exactly one of manifest or kustomize must be set
    hasManifest := c.Manifest != ""
    hasKustomize := c.Kustomize != nil

    if !hasManifest && !hasKustomize {
        return errors.New("one of 'manifest' or 'kustomize' must be specified")
    }
    if hasManifest && hasKustomize {
        return errors.New("only one of 'manifest' or 'kustomize' can be specified")
    }

    // Validate kustomize config
    if c.Kustomize != nil {
        if c.Kustomize.Path == "" {
            return errors.New("kustomize.path is required")
        }
    }

    return nil
}
```

### 2.2 Update Create Handler

**File**: `services/ctl-api/internal/app/components/service/create_kubernetes_manifest_component_config.go`

```go
func (s *service) createKubernetesManifestComponentConfig(
    ctx context.Context, cmpID string, req *CreateKubernetesManifestComponentConfigRequest,
) (*app.KubernetesManifestComponentConfig, error) {
    parentCmp, err := s.getComponentWithParents(ctx, cmpID)
    if err != nil {
        return nil, err
    }

    depIDs, err := s.helpers.GetComponentIDs(ctx, parentCmp.AppID, req.Dependencies)
    if err != nil {
        return nil, errors.Wrap(err, "unable to get component ids")
    }

    // Build component config
    cfg := app.KubernetesManifestComponentConfig{
        Manifest:  req.Manifest, // Empty for kustomize/oci_artifact sources
        Namespace: req.Namespace,
    }

    // Populate kustomize config (mutually exclusive with Manifest)
    if req.Kustomize != nil {
        cfg.Kustomize = &app.KustomizeConfig{
            Path:           req.Kustomize.Path,
            Patches:        req.Kustomize.Patches,
            EnableHelm:     req.Kustomize.EnableHelm,
            LoadRestrictor: req.Kustomize.LoadRestrictor,
        }
    }

    componentConfigConnection := app.ComponentConfigConnection{
        KubernetesManifestComponentConfig: &cfg,
        ComponentID:                       parentCmp.ID,
        AppConfigID:                       req.AppConfigID,
        References:                        pq.StringArray(req.References),
        Checksum:                          req.Checksum,
        ComponentDependencyIDs:            pq.StringArray(depIDs),
    }

    if req.DriftSchedule != nil {
        _, err := cron.ParseStandard(*req.DriftSchedule)
        if err != nil {
            return nil, fmt.Errorf("invalid drift schedule: %s", err.Error())
        }
        componentConfigConnection.DriftSchedule = *req.DriftSchedule
    }

    if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
        return nil, fmt.Errorf("unable to create kubernetes component config: %w", res.Error)
    }

    return &cfg, nil
}
```

### 2.3 Update Swagger Documentation

Add markdown description file:

**File**: `services/ctl-api/docs/public/descriptions/create_kubernetes_manifest_component_config.md`

```markdown
Creates a Kubernetes manifest component configuration.

Supports three mutually exclusive source types:
- **Inline manifest**: Provide YAML content directly in the `manifest` field
- **Kustomize**: Reference a kustomization directory via the `kustomize` field
- **OCI Artifact**: Reference a pre-built OCI artifact via the `oci_artifact` field

Only one source type can be specified per configuration.
```

---

## Phase 3: TOML Config Schema (pkg/config)

> **Note**: The TOML schema changes are already documented in [02-config-schema.md](./02-config-schema.md). 
> This section provides corrected TOML examples (the original doc incorrectly showed YAML syntax).

### 3.1 TOML Configuration Examples

#### Example 1: Inline Manifest (Existing - Unchanged)

```toml
# components/my-configmap.toml
name = "my-configmap"
type = "kubernetes_manifest"

namespace = "default"
manifest = """
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key: value
"""
```

#### Example 2: Kustomize Overlay (New)

```toml
# components/my-app.toml
name = "my-app"
type = "kubernetes_manifest"

namespace = "{{.nuon.install.id}}"

[kustomize]
path = "./k8s/overlays/production"
patches = [
  "./k8s/patches/resource-limits.yaml"
]
enable_helm = false
load_restrictor = "rootOnly"
```

> **Note**: OCI Artifact (BYOA) examples deferred to post-MVP. See [07-future-byoa-support.md](./07-future-byoa-support.md).

---

## Phase 4: CLI Integration

### 4.1 `nuon apps sync` Changes

The CLI's `apps sync` command parses TOML configs and calls the API. It needs to:

1. Parse the new `kustomize` and `oci_artifact` TOML tables
2. Map them to the API request structure
3. Send to the updated API endpoint

**File**: `bins/cli/internal/cmd/apps/sync.go` (or similar)

The existing config parsing in `pkg/config/` should handle this automatically once the structs are updated with proper `mapstructure` tags.

### 4.2 VCS/Source Handling

For Kustomize configurations:
- The `path` field is relative to the repository root
- VCS config (public_repo or connected_repo) is still required to specify where the kustomization lives
- Build runner will checkout the repo and run `kustomize build` on the specified path

---

## Phase 5: Response Model Updates

### 5.1 API Response Structure

**Inline Manifest Response:**
```json
{
  "id": "cmp_abc123...",
  "created_at": "2024-01-15T10:30:00Z",
  "org_id": "org_xyz...",
  "manifest": "apiVersion: v1\nkind: ConfigMap\n...",
  "namespace": "default",
  "kustomize_config": null
}
```

**Kustomize Response:**
```json
{
  "id": "cmp_abc123...",
  "created_at": "2024-01-15T10:30:00Z",
  "org_id": "org_xyz...",
  "manifest": "",
  "namespace": "production",
  "kustomize_config": {
    "kustomize": {
      "path": "./k8s/overlays/production",
      "patches": ["./k8s/patches/limits.yaml"],
      "enable_helm": false,
      "load_restrictor": "rootOnly"
    }
  }
}
```

### 5.2 Determining Source Type

Clients can determine the source type by checking which fields are populated:

| Source Type | `manifest` | `kustomize_config` |
|-------------|------------|-------------------|
| Inline | Non-empty | `null` |
| Kustomize | Empty | `{ "kustomize": {...} }` |

The `SourceType()` method on the model provides this logic server-side.

> **Note**: OCI Artifact response format deferred to post-MVP.

---

## Summary: Files to Modify

| File | Changes |
|------|---------|
| `services/ctl-api/internal/app/kubernetes_manifest_component_config.go` | Add `KustomizeConfig` JSONB type, `SourceType()` method |
| `services/ctl-api/internal/app/components/service/create_kubernetes_manifest_component_config.go` | Update request struct, validation, handler |
| `pkg/config/kubernetes_manifest_component.go` | Add Kustomize, OCIArtifact structs (already in 02-config-schema.md) |
| `services/ctl-api/docs/public/descriptions/create_kubernetes_manifest_component_config.md` | Update API docs |

---

## Testing Strategy

### Unit Tests

1. **Database Model Tests**
   - Test JSONB serialization/deserialization
   - Test AfterQuery backwards compatibility
   - Test BeforeCreate hooks

2. **API Validation Tests**
   - Test mutual exclusivity of manifest/kustomize/oci_artifact
   - Test required field validation for each source type
   - Test error messages

3. **TOML Parsing Tests**
   - Test inline manifest parsing (existing)
   - Test kustomize table parsing
   - Test oci_artifact table parsing
   - Test validation errors

### Integration Tests

1. **Create Component Config**
   - Create with inline manifest → verify stored correctly
   - Create with kustomize config → verify stored correctly
   - Create with oci_artifact config → verify stored correctly

2. **Read Component Config**
   - Verify legacy `manifest` field populated for inline configs
   - Verify `manifest_config` contains full configuration
   - Verify backwards compatibility with old clients

3. **Build Runner Integration**
   - Verify config is passed correctly to build runner
   - Test kustomize build produces valid OCI artifact
