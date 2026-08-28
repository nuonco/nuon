package customerbundle

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

const ociLayoutContents = `{"imageLayoutVersion":"1.0.0"}`

type Result struct {
	ManifestDescriptor ocispec.Descriptor
	BundleDescriptor   ocispec.Descriptor
	TransportSHA256    string
	Index              json.RawMessage
}

type GenerateOptions struct {
	MaxContentBytes int64
	MaxBlobBytes    int64
	OnBlobVerified  func(ocispec.Descriptor)
	BlobSink func(digest.Digest, []byte) error
}

func Generate(ctx context.Context, dst io.Writer, logical LogicalManifest, roots []Root) (Result, error) {
	return GenerateWithDocuments(ctx, dst, logical, Documents{}, roots)
}

func GenerateWithDocuments(ctx context.Context, dst io.Writer, logical LogicalManifest, documents Documents, roots []Root) (Result, error) {
	return GenerateWithOptions(ctx, dst, logical, documents, roots, GenerateOptions{})
}

func GenerateWithOptions(ctx context.Context, dst io.Writer, logical LogicalManifest, documents Documents, roots []Root, opts GenerateOptions) (Result, error) {
	if opts.MaxContentBytes < 0 || opts.MaxBlobBytes < 0 {
		return Result{}, fmt.Errorf("bundle content limits cannot be negative")
	}
	logical = canonicalize(logical)
	if err := validateMembers(logical, roots); err != nil {
		return Result{}, err
	}
	logicalJSON, err := json.Marshal(logical)
	if err != nil {
		return Result{}, fmt.Errorf("marshal logical manifest: %w", err)
	}
	logicalDesc := descriptor(LogicalManifestMediaType, logicalJSON)

	blobs := map[digest.Digest][]byte{logicalDesc.Digest: logicalJSON}
	rootDescs := make([]ocispec.Descriptor, len(roots))
	traversed := make(map[traversalKey]bool)
	var contentBytes int64
	for i, root := range roots {
		if root.Source == nil {
			return Result{}, fmt.Errorf("root %s has no source", root.Descriptor.Digest)
		}
		rootDescs[i] = canonicalDescriptor(root.Descriptor)
		if err := collect(ctx, root.Source, root.Descriptor, blobs, traversed, &contentBytes, opts); err != nil {
			return Result{}, fmt.Errorf("verify root %s: %w", root.Descriptor.Digest, err)
		}
	}
	sort.Slice(rootDescs, func(i, j int) bool { return rootDescs[i].Digest.String() < rootDescs[j].Digest.String() })
	layers := make([]ocispec.Descriptor, 0, 2)
	files := map[string][]byte{"oci-layout": []byte(ociLayoutContents), "bundle-manifest.json": logicalJSON}
	for _, document := range []struct {
		name      string
		mediaType string
		contents  json.RawMessage
	}{
		{name: "bundle-provenance.json", mediaType: ProvenanceMediaType, contents: documents.Provenance},
		{name: "qualification-report.json", mediaType: QualificationMediaType, contents: documents.QualificationReport},
		{name: "plan-envelope.json", mediaType: PlanEnvelopeMediaType, contents: documents.PlanEnvelope},
		{name: "release-source.json", mediaType: SourceArchiveMediaType, contents: documents.SourceArchive},
	} {
		if len(document.contents) == 0 {
			continue
		}
		if !json.Valid(document.contents) {
			return Result{}, fmt.Errorf("%s is not valid JSON", document.name)
		}
		desc := descriptor(document.mediaType, document.contents)
		blobs[desc.Digest] = document.contents
		layers = append(layers, desc)
		files[document.name] = document.contents
	}
	sort.Slice(layers, func(i, j int) bool { return layers[i].Digest.String() < layers[j].Digest.String() })
	bundleJSON, err := json.Marshal(ocispec.Manifest{
		Versioned:    specsVersioned(),
		MediaType:    ocispec.MediaTypeImageManifest,
		ArtifactType: BundleArtifactType,
		Config:       logicalDesc,
		Layers:       layers,
	})
	if err != nil {
		return Result{}, fmt.Errorf("marshal bundle manifest: %w", err)
	}
	bundleDesc := descriptor(ocispec.MediaTypeImageManifest, bundleJSON)
	bundleDesc.ArtifactType = BundleArtifactType
	blobs[bundleDesc.Digest] = bundleJSON
	indexRoots := append([]ocispec.Descriptor{bundleDesc}, rootDescs...)
	indexJSON, err := json.Marshal(ocispec.Index{Versioned: specsVersioned(), MediaType: ocispec.MediaTypeImageIndex, Manifests: indexRoots})
	if err != nil {
		return Result{}, fmt.Errorf("marshal index: %w", err)
	}

	files["index.json"] = indexJSON
	for dgst, data := range blobs {
		files["blobs/"+dgst.Algorithm().String()+"/"+dgst.Encoded()] = data
	}
	h := sha256.New()
	if err := writeArchive(io.MultiWriter(dst, h), files); err != nil {
		return Result{}, err
	}
	if opts.BlobSink != nil {
		digests := make([]digest.Digest, 0, len(blobs))
		for dgst := range blobs {
			digests = append(digests, dgst)
		}
		sort.Slice(digests, func(i, j int) bool { return digests[i].String() < digests[j].String() })
		for _, dgst := range digests {
			if err := opts.BlobSink(dgst, blobs[dgst]); err != nil {
				return Result{}, fmt.Errorf("sink blob %s: %w", dgst, err)
			}
		}
	}
	return Result{ManifestDescriptor: logicalDesc, BundleDescriptor: bundleDesc, TransportSHA256: hex.EncodeToString(h.Sum(nil)), Index: indexJSON}, nil
}

func specsVersioned() specs.Versioned { return specs.Versioned{SchemaVersion: 2} }

func canonicalize(m LogicalManifest) LogicalManifest {
	sort.Slice(m.Components, func(i, j int) bool { return m.Components[i].Name < m.Components[j].Name })
	sort.Slice(m.Images, func(i, j int) bool { return m.Images[i].Name < m.Images[j].Name })
	sort.Slice(m.Actions, func(i, j int) bool { return m.Actions[i].Name < m.Actions[j].Name })
	for i := range m.Actions {
		sort.Slice(m.Actions[i].Steps, func(j, k int) bool { return m.Actions[i].Steps[j].Name < m.Actions[i].Steps[k].Name })
	}
	sort.Slice(m.Runbooks, func(i, j int) bool { return m.Runbooks[i].Name < m.Runbooks[j].Name })
	sort.Slice(m.StackAssets, func(i, j int) bool { return m.StackAssets[i].Role < m.StackAssets[j].Role })
	return m
}

func validateMembers(logical LogicalManifest, roots []Root) error {
	if logical.SchemaVersion < 1 || logical.SchemaVersion > CurrentSchemaVersion {
		return fmt.Errorf("unsupported bundle manifest schema version %d", logical.SchemaVersion)
	}
	if logical.Target.OS != "linux" || logical.Target.Architecture != "amd64" {
		return fmt.Errorf("unsupported bundle target %s/%s", logical.Target.OS, logical.Target.Architecture)
	}
	expected := make(map[digest.Digest]ocispec.Descriptor)
	keys := make(map[string]bool)
	claim := func(key string) error {
		if key == "" {
			return fmt.Errorf("bundle member has an empty logical key")
		}
		if keys[key] {
			return fmt.Errorf("duplicate bundle member logical key %q", key)
		}
		keys[key] = true
		return nil
	}
	add := func(key string, artifact Artifact) error {
		if err := claim(key); err != nil {
			return err
		}
		desc := ocispec.Descriptor{MediaType: artifact.MediaType, Digest: digest.Digest(artifact.Digest), Size: artifact.Size}
		if artifact.PlatformOS != "" || artifact.PlatformArchitecture != "" {
			if artifact.PlatformOS != logical.Target.OS || artifact.PlatformArchitecture != logical.Target.Architecture {
				return fmt.Errorf("artifact %s platform does not match bundle target", key)
			}
			desc.Platform = &ocispec.Platform{OS: artifact.PlatformOS, Architecture: artifact.PlatformArchitecture}
		}
		if err := validateDescriptor(desc); err != nil {
			return fmt.Errorf("invalid artifact for %s: %w", key, err)
		}
		if previous, ok := expected[desc.Digest]; ok && (previous.MediaType != desc.MediaType || previous.Size != desc.Size || !samePlatform(previous.Platform, desc.Platform)) {
			return fmt.Errorf("conflicting descriptors for digest %s", desc.Digest)
		}
		expected[desc.Digest] = desc
		return nil
	}
	for _, component := range logical.Components {
		if component.Name == "" || component.Type == "" {
			return fmt.Errorf("component name and type are required")
		}
		if err := validateContentDigest(component.ConfigDigest); err != nil {
			return fmt.Errorf("component %s config digest: %w", component.Name, err)
		}
		if err := add("component:"+component.Name, component.Artifact); err != nil {
			return err
		}
	}
	if logical.Sandbox != nil {
		if logical.Sandbox.Type == "" {
			return fmt.Errorf("sandbox type is required")
		}
		if err := validateContentDigest(logical.Sandbox.ConfigDigest); err != nil {
			return fmt.Errorf("sandbox config digest: %w", err)
		}
		if err := add("sandbox", logical.Sandbox.Artifact); err != nil {
			return err
		}
	}
	for _, image := range logical.Images {
		if image.Name == "" || image.Repository == "" {
			return fmt.Errorf("image name and repository are required")
		}
		if err := add("image:"+image.Name, image.Artifact); err != nil {
			return err
		}
	}
	for _, asset := range logical.StackAssets {
		if asset.Role == "" || asset.SourceURL == "" {
			return fmt.Errorf("stack asset role and source URL are required")
		}
		if err := add("stack-asset:"+asset.Role, Artifact{MediaType: asset.MediaType, Digest: asset.Digest, Size: asset.Size}); err != nil {
			return err
		}
	}
	for _, action := range logical.Actions {
		if action.Name == "" {
			return fmt.Errorf("action name is required")
		}
		if err := validateContentDigest(action.ConfigDigest); err != nil {
			return fmt.Errorf("action %s config digest: %w", action.Name, err)
		}
		if err := claim("action:" + action.Name); err != nil {
			return err
		}
		for _, step := range action.Steps {
			if step.Name == "" {
				return fmt.Errorf("action %s step name is required", action.Name)
			}
			if step.InlineContentsDigest != "" {
				if err := validateContentDigest(step.InlineContentsDigest); err != nil {
					return fmt.Errorf("action %s step %s inline contents digest: %w", action.Name, step.Name, err)
				}
			}
			if step.Artifact != nil {
				if err := add("action:"+action.Name+"/step-artifact:"+step.Name, *step.Artifact); err != nil {
					return err
				}
			}
			if err := claim("action:" + action.Name + "/step:" + step.Name); err != nil {
				return err
			}
		}
	}
	for _, runbook := range logical.Runbooks {
		if runbook.Name == "" {
			return fmt.Errorf("runbook name is required")
		}
		if err := validateContentDigest(runbook.ConfigDigest); err != nil {
			return fmt.Errorf("runbook %s config digest: %w", runbook.Name, err)
		}
		if err := claim("runbook:" + runbook.Name); err != nil {
			return err
		}
	}
	if logical.Runner != nil {
		if logical.Runner.Binary == nil {
			return fmt.Errorf("runner binary artifact is required")
		}
		if err := add("runner:binary", *logical.Runner.Binary); err != nil {
			return err
		}
		if logical.Runner.Image != nil {
			if logical.Runner.Image.Repository == "" {
				return fmt.Errorf("runner image repository is required")
			}
			if err := add("runner:image", logical.Runner.Image.Artifact); err != nil {
				return err
			}
		}
	}
	actual := make(map[digest.Digest]ocispec.Descriptor, len(roots))
	for _, root := range roots {
		if err := validateDescriptor(root.Descriptor); err != nil {
			return fmt.Errorf("invalid root descriptor: %w", err)
		}
		if _, duplicate := actual[root.Descriptor.Digest]; duplicate {
			return fmt.Errorf("duplicate root descriptor %s", root.Descriptor.Digest)
		}
		actual[root.Descriptor.Digest] = root.Descriptor
	}
	for dgst, desc := range expected {
		root, ok := actual[dgst]
		if !ok {
			return fmt.Errorf("bundle member root %s is missing", dgst)
		}
		if root.MediaType != desc.MediaType || root.Size != desc.Size || !samePlatform(root.Platform, desc.Platform) {
			return fmt.Errorf("bundle member root %s does not match its manifest descriptor", dgst)
		}
	}
	for dgst := range actual {
		if _, ok := expected[dgst]; !ok {
			return fmt.Errorf("undeclared bundle root %s", dgst)
		}
	}
	return nil
}

func validateContentDigest(value string) error {
	dgst := digest.Digest(value)
	if err := dgst.Validate(); err != nil || dgst.Algorithm() != digest.SHA256 {
		return fmt.Errorf("valid sha256 digest is required")
	}
	return nil
}

func canonicalDescriptor(desc ocispec.Descriptor) ocispec.Descriptor {
	return ocispec.Descriptor{MediaType: desc.MediaType, Digest: desc.Digest, Size: desc.Size, Platform: desc.Platform}
}

func samePlatform(a, b *ocispec.Platform) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return a.OS == b.OS && a.Architecture == b.Architecture
}

func validateDescriptor(desc ocispec.Descriptor) error {
	if desc.MediaType == "" {
		return fmt.Errorf("media type is required")
	}
	if desc.Size < 0 {
		return fmt.Errorf("descriptor size cannot be negative")
	}
	if err := desc.Digest.Validate(); err != nil {
		return fmt.Errorf("invalid digest: %w", err)
	}
	if desc.Digest.Algorithm() != digest.SHA256 {
		return fmt.Errorf("unsupported digest algorithm %s", desc.Digest.Algorithm())
	}
	return nil
}

func descriptor(mediaType string, b []byte) ocispec.Descriptor {
	return ocispec.Descriptor{MediaType: mediaType, Digest: digest.FromBytes(b), Size: int64(len(b))}
}

type traversalKey struct {
	digest    digest.Digest
	mediaType string
}

func collect(ctx context.Context, src oras.ReadOnlyTarget, desc ocispec.Descriptor, out map[digest.Digest][]byte, traversed map[traversalKey]bool, contentBytes *int64, opts GenerateOptions) error {
	if err := validateDescriptor(desc); err != nil {
		return err
	}
	b, fetched := out[desc.Digest]
	if fetched {
		if int64(len(b)) != desc.Size {
			return fmt.Errorf("conflicting size for %s: expected %d, got %d", desc.Digest, desc.Size, len(b))
		}
	} else {
		if opts.MaxBlobBytes > 0 && desc.Size > opts.MaxBlobBytes {
			return fmt.Errorf("blob %s size %d exceeds limit %d", desc.Digest, desc.Size, opts.MaxBlobBytes)
		}
		if opts.MaxContentBytes > 0 && desc.Size > opts.MaxContentBytes-*contentBytes {
			return fmt.Errorf("bundle content size exceeds limit %d", opts.MaxContentBytes)
		}
		r, err := src.Fetch(ctx, desc)
		if err != nil {
			return fmt.Errorf("fetch %s: %w", desc.Digest, err)
		}
		readLimit := desc.Size
		if readLimit < int64(^uint64(0)>>1) {
			readLimit++
		}
		b, err = io.ReadAll(io.LimitReader(r, readLimit))
		closeErr := r.Close()
		if err != nil {
			return fmt.Errorf("read %s: %w", desc.Digest, err)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", desc.Digest, closeErr)
		}
		if int64(len(b)) != desc.Size {
			return fmt.Errorf("size mismatch for %s: expected %d, got %d", desc.Digest, desc.Size, len(b))
		}
		if got := desc.Digest.Algorithm().FromBytes(b); got != desc.Digest {
			return fmt.Errorf("digest mismatch for %s: got %s", desc.Digest, got)
		}
		out[desc.Digest] = b
		*contentBytes += int64(len(b))
		if opts.OnBlobVerified != nil {
			opts.OnBlobVerified(desc)
		}
	}
	key := traversalKey{digest: desc.Digest, mediaType: desc.MediaType}
	if traversed[key] {
		return nil
	}
	traversed[key] = true
	children, err := Successors(desc.MediaType, b)
	if err != nil {
		return fmt.Errorf("parse %s: %w", desc.Digest, err)
	}
	for _, child := range children {
		if err := collect(ctx, src, child, out, traversed, contentBytes, opts); err != nil {
			return err
		}
	}
	return nil
}

func TotalSize(ctx context.Context, src oras.ReadOnlyTarget, desc ocispec.Descriptor) (int64, error) {
	seen := map[digest.Digest]bool{}
	var total int64
	var walk func(ocispec.Descriptor) error
	walk = func(d ocispec.Descriptor) error {
		if err := validateDescriptor(d); err != nil {
			return err
		}
		if seen[d.Digest] {
			return nil
		}
		seen[d.Digest] = true
		total += d.Size
		if !IsManifestMediaType(d.MediaType) {
			return nil
		}
		r, err := src.Fetch(ctx, d)
		if err != nil {
			return fmt.Errorf("fetch %s: %w", d.Digest, err)
		}
		b, err := io.ReadAll(io.LimitReader(r, d.Size+1))
		closeErr := r.Close()
		if err != nil {
			return fmt.Errorf("read %s: %w", d.Digest, err)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", d.Digest, closeErr)
		}
		if int64(len(b)) != d.Size {
			return fmt.Errorf("size mismatch for %s: expected %d, got %d", d.Digest, d.Size, len(b))
		}
		children, err := Successors(d.MediaType, b)
		if err != nil {
			return fmt.Errorf("parse %s: %w", d.Digest, err)
		}
		for _, child := range children {
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(desc); err != nil {
		return 0, err
	}
	return total, nil
}

func IsManifestMediaType(mediaType string) bool {
	switch mediaType {
	case ocispec.MediaTypeImageManifest, "application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.oci.artifact.manifest.v1+json",
		ocispec.MediaTypeImageIndex, "application/vnd.docker.distribution.manifest.list.v2+json":
		return true
	default:
		return false
	}
}

func Successors(mediaType string, b []byte) ([]ocispec.Descriptor, error) {
	switch mediaType {
	case ocispec.MediaTypeImageManifest, "application/vnd.docker.distribution.manifest.v2+json":
		var m struct {
			SchemaVersion int                  `json:"schemaVersion"`
			Config        ocispec.Descriptor   `json:"config"`
			Layers        []ocispec.Descriptor `json:"layers"`
			Subject       *ocispec.Descriptor  `json:"subject"`
		}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		if m.SchemaVersion != 2 || m.Config.Digest == "" {
			return nil, fmt.Errorf("invalid OCI image manifest")
		}
		var result []ocispec.Descriptor
		result = append(result, m.Config)
		result = append(result, m.Layers...)
		if m.Subject != nil {
			result = append(result, *m.Subject)
		}
		return result, nil
	case "application/vnd.oci.artifact.manifest.v1+json":
		var m struct {
			SchemaVersion int                  `json:"schemaVersion"`
			MediaType     string               `json:"mediaType"`
			ArtifactType  string               `json:"artifactType"`
			Blobs         []ocispec.Descriptor `json:"blobs"`
			Subject       *ocispec.Descriptor  `json:"subject"`
		}
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, err
		}
		if m.SchemaVersion != 2 || m.MediaType != mediaType || m.ArtifactType == "" || m.Blobs == nil {
			return nil, fmt.Errorf("invalid OCI artifact manifest")
		}
		if m.Subject != nil {
			m.Blobs = append(m.Blobs, *m.Subject)
		}
		return m.Blobs, nil
	case ocispec.MediaTypeImageIndex, "application/vnd.docker.distribution.manifest.list.v2+json":
		var idx struct {
			SchemaVersion int                  `json:"schemaVersion"`
			Manifests     []ocispec.Descriptor `json:"manifests"`
			Subject       *ocispec.Descriptor  `json:"subject"`
		}
		if err := json.Unmarshal(b, &idx); err != nil {
			return nil, err
		}
		if idx.SchemaVersion != 2 {
			return nil, fmt.Errorf("invalid OCI image index")
		}
		if idx.Subject != nil {
			idx.Manifests = append(idx.Manifests, *idx.Subject)
		}
		return idx.Manifests, nil
	default:
		return nil, nil
	}
}

func writeArchive(dst io.Writer, files map[string][]byte) error {
	zw, err := zstd.NewWriter(dst, zstd.WithEncoderConcurrency(1), zstd.WithEncoderCRC(true))
	if err != nil {
		return fmt.Errorf("create zstd writer: %w", err)
	}
	tw := tar.NewWriter(zw)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		h := &tar.Header{Name: strings.TrimPrefix(name, "/"), Mode: 0644, Size: int64(len(files[name])), ModTime: time.Unix(0, 0), AccessTime: time.Unix(0, 0), ChangeTime: time.Unix(0, 0), Uid: 0, Gid: 0, Format: tar.FormatPAX}
		if err := tw.WriteHeader(h); err != nil {
			return fmt.Errorf("write tar header: %w", err)
		}
		if _, err := io.Copy(tw, bytes.NewReader(files[name])); err != nil {
			return fmt.Errorf("write tar file: %w", err)
		}
	}
	if err := tw.Close(); err != nil {
		return fmt.Errorf("close tar: %w", err)
	}
	if err := zw.Close(); err != nil {
		return fmt.Errorf("close zstd: %w", err)
	}
	return nil
}
