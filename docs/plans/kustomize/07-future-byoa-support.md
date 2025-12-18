# Future: BYOA (Bring Your Own Artifact) Support

This document outlines the planned BYOA feature for Kubernetes manifest components, deferred from MVP.

---

## Overview

BYOA allows users to provide pre-packaged OCI artifacts containing Kubernetes manifests, instead of having Nuon build them from source.

**Status**: Deferred from MVP - planned for future release.

---

## Why BYOA?

1. **Pre-existing CI/CD Pipelines**: Teams already building and packaging manifests in their CI
2. **External Vendors**: Third-party manifests distributed as OCI artifacts
3. **Build Optimization**: Skip Nuon build phase for pre-validated artifacts
4. **Compliance**: Artifacts built in customer-controlled environments

---

## Internal Architecture (Already Supported)

The internal runner architecture already handles OCI artifacts:

- **Build Runner**: Can mirror/copy external OCI artifacts to Nuon registry
- **Install Runner**: Always pulls and applies from OCI artifacts

This means adding BYOA UX is primarily a configuration layer change.

---

## Planned Configuration Schema

```go
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

---

## Example Configurations (Planned)

### Pre-built OCI Artifact

```toml
# components/vendor-app.toml
name = "vendor-app"
type = "kubernetes_manifest"

namespace = "vendor"

[oci_artifact]
url = "oci://ghcr.io/vendor/app-manifests"
tag = "v2.3.1"
# OR use digest for immutability:
# digest = "sha256:abc123..."
```

### Private Registry OCI Artifact

```toml
# components/internal-app.toml
name = "internal-app"
type = "kubernetes_manifest"

namespace = "internal"

[oci_artifact]
url = "oci://123456789.dkr.ecr.us-west-2.amazonaws.com/manifests"
tag = "v1.0.0"
credentials_ref = "aws-ecr-creds"
```

---

## User Workflow (Planned)

1. **Build Kustomize overlay** locally or in CI:
   ```bash
   kustomize build ./k8s/overlays/production > manifests.yaml
   ```

2. **Package into OCI artifact** using kustomizer:
   ```bash
   kustomizer push artifact oci://ghcr.io/org/app-manifests:v1.0.0 \
     -k ./k8s/overlays/production
   ```

3. **Reference the artifact** in component TOML:
   ```toml
   [oci_artifact]
   url = "oci://ghcr.io/org/app-manifests"
   tag = "v1.0.0"
   ```

4. **Deploy as usual**

---

## Implementation Tasks

When implementing BYOA support:

1. **Configuration Schema**: Add `OCIArtifact` field to `KubernetesManifestComponentConfig`
2. **Validation**: Implement mutual exclusivity (manifest/kustomize/oci_artifact)
3. **Build Runner**: Add artifact mirroring/copying logic for external registries
4. **API/CLI**: Expose configuration through API and CLI sync
5. **Dashboard UI**: Add BYOA configuration options in component editor
6. **Documentation**: Update user docs with BYOA examples

---

## Open Questions

1. **Registry Authentication**: How to securely reference credentials for private registries?
2. **Artifact Validation**: Should we validate artifact contents before mirroring?
3. **Digest Pinning**: Auto-resolve tag to digest for immutability?
4. **Cross-Cloud**: How to handle artifacts in registries not accessible from install clusters?

---

## Related Documents

- [Main Kustomize Plan](../kustomize-support-plan.md)
- [Config Schema](./02-config-schema.md)
- [Build Runner](./03-build-runner.md)
- [Install Runner](./04-install-runner.md)
