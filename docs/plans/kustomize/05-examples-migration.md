# Kustomize Support: Configuration Examples & Migration Guide

This document provides configuration examples and migration guidance for users.

---

## Configuration Examples

### Example 1: Inline Manifest (Existing - Unchanged)

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

### Example 2: Kustomize Overlay (New)

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
```

---

> **Note**: BYOA (Bring Your Own Artifact) examples are planned for future support. See [07-future-byoa-support.md](./07-future-byoa-support.md) for the roadmap.

## Kustomize Directory Structure

When using the `kustomize` option, your repository should have a structure like:

```
my-app/
├── nuon.toml
├── components/
│   └── my-app.toml
└── k8s/
    ├── base/
    │   ├── kustomization.yaml
    │   ├── deployment.yaml
    │   └── service.yaml
    ├── overlays/
    │   ├── staging/
    │   │   ├── kustomization.yaml
    │   │   └── replica-patch.yaml
    │   └── production/
    │       ├── kustomization.yaml
    │       └── resource-limits.yaml
    └── patches/
        └── resource-limits.yaml
```

**Example `kustomization.yaml`:**

```yaml
# k8s/overlays/production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - ../../base

patches:
  - path: resource-limits.yaml
    target:
      kind: Deployment

replicas:
  - name: my-app
    count: 3

images:
  - name: my-app
    newTag: v1.2.3
```

---

## Migration Guide

### For Existing Users

No action required. Existing configurations using the `manifest` field will continue to work unchanged.

**What happens behind the scenes:**
1. The build runner packages your inline manifest into an OCI artifact
2. The install runner pulls from the OCI artifact
3. All existing behavior (diff detection, apply, teardown) works identically

### For New Kustomize Users

1. **Create a `kustomization.yaml`** in your repository:
   ```yaml
   # k8s/overlays/production/kustomization.yaml
   apiVersion: kustomize.config.k8s.io/v1beta1
   kind: Kustomization
   
   resources:
     - deployment.yaml
     - service.yaml
   ```

2. **Update your component TOML** to use the `kustomize` table:
   ```toml
   # components/my-app.toml
   name = "my-app"
   type = "kubernetes_manifest"
   
   namespace = "{{.nuon.install.id}}"
   
   [kustomize]
   path = "./k8s/overlays/production"
   ```

3. **Run `nuon apps sync`** to detect the new configuration

4. **Deploy as usual**

---

## Kustomize Configuration Options

### `path` (required)
Path to the kustomization directory relative to the source root.

```toml
[kustomize]
path = "./k8s/overlays/production"
```

### `patches` (optional)
Additional patch files to apply after the kustomize build. Useful for environment-specific overrides.

```toml
[kustomize]
path = "./k8s/base"
patches = [
  "./k8s/patches/resource-limits.yaml",
  "./k8s/patches/replicas.yaml"
]
```

### `enable_helm` (optional)
Enable Helm chart inflation during kustomize build. This is equivalent to `--enable-helm` flag.

```toml
[kustomize]
path = "./k8s/with-helm-charts"
enable_helm = true
```

### `load_restrictor` (optional)
Control whether kustomize can load files outside the kustomization directory.

- `rootOnly` (default): Only load files within the kustomization root
- `none`: Allow loading files from any location

```toml
[kustomize]
path = "./k8s/overlays/production"
load_restrictor = "none"  # Use with caution
```

---

## Troubleshooting

### "kustomize.path is required"

Ensure your kustomize configuration includes the `path` field:

```toml
# ❌ Wrong
[kustomize]

# ✅ Correct
[kustomize]
path = "./k8s/overlays/production"
```

### "only one of 'manifest' or 'kustomize' can be specified"

You can only use one source type per component. Split into multiple components if needed:

```toml
# ❌ Wrong - both manifest and kustomize in same file
name = "my-app"
type = "kubernetes_manifest"
manifest = """
apiVersion: v1
kind: ConfigMap
...
"""

[kustomize]
path = "./k8s"
```

```toml
# ✅ Correct - use separate component files

# components/configmap.toml
name = "configmap"
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

```toml
# components/app.toml
name = "app"
type = "kubernetes_manifest"
namespace = "default"

[kustomize]
path = "./k8s"
```
