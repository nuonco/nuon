package customerbundle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const ociLayoutContents = "{\"imageLayoutVersion\":\"1.0.0\"}\n"

// FetchBlob returns the raw bytes of a content-addressed bundle blob.
type FetchBlob func(ctx context.Context, dgst digest.Digest) ([]byte, error)

// Reassemble reconstructs the exact deterministic bundle archive from its OCI
// index and a content-addressed blob source, and returns the archive's
// transport SHA-256. Because writeArchive is deterministic, the checksum of a
// reassembled archive matches the checksum recorded when the bundle was
// generated, which proves byte-exact reconstruction after a differential
// download.
func Reassemble(ctx context.Context, dst io.Writer, indexJSON []byte, fetch FetchBlob) (string, error) {
	if fetch == nil {
		return "", fmt.Errorf("blob fetcher is required")
	}
	var index ocispec.Index
	if err := json.Unmarshal(indexJSON, &index); err != nil {
		return "", fmt.Errorf("parse bundle index: %w", err)
	}
	if index.SchemaVersion != 2 || len(index.Manifests) == 0 {
		return "", fmt.Errorf("invalid bundle index")
	}
	blobs := map[digest.Digest][]byte{}
	var walk func(desc ocispec.Descriptor) error
	walk = func(desc ocispec.Descriptor) error {
		if err := desc.Digest.Validate(); err != nil {
			return fmt.Errorf("invalid descriptor digest %q: %w", desc.Digest, err)
		}
		if _, ok := blobs[desc.Digest]; ok {
			return nil
		}
		data, err := fetch(ctx, desc.Digest)
		if err != nil {
			return fmt.Errorf("fetch blob %s: %w", desc.Digest, err)
		}
		if desc.Size >= 0 && desc.Size != int64(len(data)) {
			return fmt.Errorf("blob %s size mismatch: descriptor declares %d bytes, got %d", desc.Digest, desc.Size, len(data))
		}
		if computed := digest.FromBytes(data); computed != desc.Digest {
			return fmt.Errorf("blob digest mismatch: expected %s, got %s", desc.Digest, computed)
		}
		blobs[desc.Digest] = data
		children, err := Successors(desc.MediaType, data)
		if err != nil {
			return fmt.Errorf("traverse blob %s: %w", desc.Digest, err)
		}
		for _, child := range children {
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}
	var bundleManifest *ocispec.Manifest
	for _, root := range index.Manifests {
		if err := walk(root); err != nil {
			return "", err
		}
		if root.ArtifactType != BundleArtifactType {
			continue
		}
		var manifest ocispec.Manifest
		if err := json.Unmarshal(blobs[root.Digest], &manifest); err != nil {
			return "", fmt.Errorf("parse bundle manifest %s: %w", root.Digest, err)
		}
		bundleManifest = &manifest
	}
	if bundleManifest == nil {
		return "", fmt.Errorf("bundle index has no root with artifact type %s", BundleArtifactType)
	}
	if bundleManifest.Config.MediaType != LogicalManifestMediaType {
		return "", fmt.Errorf("bundle manifest config has media type %q, expected %q", bundleManifest.Config.MediaType, LogicalManifestMediaType)
	}
	files := map[string][]byte{
		"oci-layout":           []byte(ociLayoutContents),
		"index.json":           indexJSON,
		"bundle-manifest.json": blobs[bundleManifest.Config.Digest],
	}
	documentNames := map[string]string{
		ProvenanceMediaType:    "bundle-provenance.json",
		QualificationMediaType: "qualification-report.json",
		PlanEnvelopeMediaType:  "plan-envelope.json",
	}
	for _, layer := range bundleManifest.Layers {
		name, ok := documentNames[layer.MediaType]
		if !ok {
			return "", fmt.Errorf("bundle manifest layer %s has unknown media type %q", layer.Digest, layer.MediaType)
		}
		files[name] = blobs[layer.Digest]
	}
	for dgst, data := range blobs {
		files["blobs/"+dgst.Algorithm().String()+"/"+dgst.Encoded()] = data
	}
	h := sha256.New()
	if err := writeArchive(io.MultiWriter(dst, h), files); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
