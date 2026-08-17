package bundle

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/oci"
)

type Bundle struct {
	Manifest         LogicalManifest
	BundleDescriptor ocispec.Descriptor
	Roots            []ocispec.Descriptor
	Provenance       json.RawMessage
	Qualification    json.RawMessage
	PlanEnvelope     json.RawMessage
	store            oras.ReadOnlyTarget
}

type Member struct {
	Key       string        `json:"key"`
	Kind      string        `json:"kind"`
	Name      string        `json:"name"`
	MediaType string        `json:"media_type"`
	Digest    digest.Digest `json:"digest"`
	Size      int64         `json:"size"`
}

const (
	DefaultMaxArchiveEntries   = 1_000_000
	DefaultMaxArchiveFileBytes = int64(10 << 30)
	DefaultMaxArchiveBytes     = int64(20 << 30)
	DefaultMaxDecoderMemory    = uint64(1 << 30)
	DefaultMaxDecoderWindow    = uint64(1 << 30)
	DefaultMaxMetadataBytes    = int64(16 << 20)
)

type ExtractOptions struct {
	MaxEntries       int
	MaxFileBytes     int64
	MaxTotalBytes    int64
	MaxDecoderMemory uint64
	MaxDecoderWindow uint64
}

func Extract(dst string, r io.Reader) (string, error) {
	return ExtractWithOptions(dst, r, ExtractOptions{
		MaxEntries:       DefaultMaxArchiveEntries,
		MaxFileBytes:     DefaultMaxArchiveFileBytes,
		MaxTotalBytes:    DefaultMaxArchiveBytes,
		MaxDecoderMemory: DefaultMaxDecoderMemory,
		MaxDecoderWindow: DefaultMaxDecoderWindow,
	})
}

func ExtractWithOptions(dst string, r io.Reader, opts ExtractOptions) (_ string, retErr error) {
	if opts.MaxEntries <= 0 || opts.MaxFileBytes <= 0 || opts.MaxTotalBytes <= 0 || opts.MaxDecoderMemory == 0 || opts.MaxDecoderWindow == 0 {
		return "", fmt.Errorf("archive extraction limits must be positive")
	}
	if err := prepareExtractionDirectory(dst); err != nil {
		return "", err
	}
	defer func() {
		if retErr != nil {
			entries, _ := os.ReadDir(dst)
			for _, entry := range entries {
				_ = os.RemoveAll(filepath.Join(dst, entry.Name()))
			}
		}
	}()
	h := sha256.New()
	zr, err := zstd.NewReader(io.TeeReader(r, h),
		zstd.WithDecoderConcurrency(1),
		zstd.WithDecoderMaxMemory(opts.MaxDecoderMemory),
		zstd.WithDecoderMaxWindow(opts.MaxDecoderWindow),
	)
	if err != nil {
		return "", fmt.Errorf("unable to create zstd reader: %w", err)
	}
	defer zr.Close()
	tr := tar.NewReader(zr)
	seen := make(map[string]struct{})
	var entries int
	var totalBytes int64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("unable to read tar: %w", err)
		}
		entries++
		if entries > opts.MaxEntries {
			return "", fmt.Errorf("archive entry count exceeds limit %d", opts.MaxEntries)
		}
		clean := path.Clean(hdr.Name)
		if hdr.Name == "" || hdr.Name != clean || path.IsAbs(hdr.Name) || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(hdr.Name, `\`) {
			return "", fmt.Errorf("unable to extract unsafe path %q", hdr.Name)
		}
		if _, ok := seen[clean]; ok {
			return "", fmt.Errorf("unable to extract duplicate path %q", hdr.Name)
		}
		seen[clean] = struct{}{}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			return "", fmt.Errorf("unable to extract non-regular entry %q", hdr.Name)
		}
		if hdr.Size < 0 || hdr.Size > opts.MaxFileBytes {
			return "", fmt.Errorf("archive file %q size %d exceeds limit %d", hdr.Name, hdr.Size, opts.MaxFileBytes)
		}
		if hdr.Size > opts.MaxTotalBytes-totalBytes {
			return "", fmt.Errorf("archive expanded size exceeds limit %d", opts.MaxTotalBytes)
		}
		totalBytes += hdr.Size
		filePath := filepath.Join(dst, filepath.FromSlash(clean))
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return "", fmt.Errorf("unable to create parent directory: %w", err)
		}
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			return "", fmt.Errorf("unable to create %s: %w", clean, err)
		}
		_, copyErr := io.Copy(f, tr)
		closeErr := f.Close()
		if copyErr != nil {
			return "", fmt.Errorf("unable to write %s: %w", clean, copyErr)
		}
		if closeErr != nil {
			return "", fmt.Errorf("unable to close %s: %w", clean, closeErr)
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func prepareExtractionDirectory(dst string) error {
	if err := os.MkdirAll(dst, 0700); err != nil {
		return fmt.Errorf("unable to create extraction directory: %w", err)
	}
	info, err := os.Lstat(dst)
	if err != nil {
		return fmt.Errorf("unable to inspect extraction directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("extraction destination must be a directory, not a symlink")
	}
	entries, err := os.ReadDir(dst)
	if err != nil {
		return fmt.Errorf("unable to inspect extraction directory: %w", err)
	}
	if len(entries) != 0 {
		return fmt.Errorf("extraction destination must be empty")
	}
	return nil
}

func Open(ctx context.Context, dir string) (*Bundle, error) {
	store, err := oci.NewFromFS(ctx, os.DirFS(dir))
	if err != nil {
		return nil, fmt.Errorf("unable to open OCI layout: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "index.json"))
	if err != nil {
		return nil, fmt.Errorf("unable to read index: %w", err)
	}
	var index ocispec.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("unable to parse index: %w", err)
	}
	var bundleDescs []ocispec.Descriptor
	var roots []ocispec.Descriptor
	for _, desc := range index.Manifests {
		if desc.ArtifactType == BundleArtifactType {
			bundleDescs = append(bundleDescs, desc)
		} else {
			roots = append(roots, desc)
		}
	}
	if len(bundleDescs) != 1 {
		return nil, fmt.Errorf("unable to open bundle: expected exactly one bundle descriptor, got %d", len(bundleDescs))
	}
	var manifest ocispec.Manifest
	if err := fetchJSON(ctx, store, bundleDescs[0], &manifest); err != nil {
		return nil, err
	}
	if manifest.ArtifactType != BundleArtifactType || manifest.Config.MediaType != LogicalManifestMediaType {
		return nil, fmt.Errorf("unable to open bundle: invalid bundle manifest")
	}
	var logical LogicalManifest
	if err := fetchJSON(ctx, store, manifest.Config, &logical); err != nil {
		return nil, err
	}
	validationRoots := make([]Root, len(roots))
	for i, desc := range roots {
		validationRoots[i] = Root{Descriptor: desc, Source: store}
	}
	if err := validateMembers(logical, validationRoots); err != nil {
		return nil, fmt.Errorf("unable to validate bundle: %w", err)
	}
	b := &Bundle{Manifest: logical, BundleDescriptor: bundleDescs[0], Roots: roots, store: store}
	for _, layer := range manifest.Layers {
		var dst *json.RawMessage
		switch layer.MediaType {
		case ProvenanceMediaType:
			dst = &b.Provenance
		case QualificationMediaType:
			dst = &b.Qualification
		case PlanEnvelopeMediaType:
			dst = &b.PlanEnvelope
		default:
			continue
		}
		if err := fetchJSON(ctx, store, layer, dst); err != nil {
			return nil, err
		}
	}
	return b, nil
}

func fetchJSON(ctx context.Context, store oras.ReadOnlyTarget, desc ocispec.Descriptor, dst any) error {
	if desc.Size < 0 || desc.Size > DefaultMaxMetadataBytes {
		return fmt.Errorf("unable to parse %s: metadata size %d exceeds limit %d", desc.Digest, desc.Size, DefaultMaxMetadataBytes)
	}
	r, err := store.Fetch(ctx, desc)
	if err != nil {
		return fmt.Errorf("unable to fetch %s: %w", desc.Digest, err)
	}
	defer r.Close()
	if err := json.NewDecoder(io.LimitReader(r, DefaultMaxMetadataBytes+1)).Decode(dst); err != nil {
		return fmt.Errorf("unable to parse %s: %w", desc.Digest, err)
	}
	return nil
}

func (b *Bundle) Store() oras.ReadOnlyTarget { return b.store }

func (b *Bundle) Members() []Member {
	var out []Member
	add := func(key, kind, name string, a Artifact) {
		out = append(out, Member{Key: key, Kind: kind, Name: name, MediaType: a.MediaType, Digest: digest.Digest(a.Digest), Size: a.Size})
	}
	for _, v := range b.Manifest.Components {
		add("component:"+v.Name, "component", v.Name, v.Artifact)
	}
	if v := b.Manifest.Sandbox; v != nil {
		add("sandbox", "sandbox", v.Type, v.Artifact)
	}
	for _, v := range b.Manifest.Images {
		add("image:"+v.Name, "image", v.Name, v.Artifact)
	}
	for _, action := range b.Manifest.Actions {
		for _, step := range action.Steps {
			if step.Artifact != nil {
				add("action:"+action.Name+"/step-artifact:"+step.Name, "action-artifact", action.Name+"/"+step.Name, *step.Artifact)
			}
		}
	}
	for _, v := range b.Manifest.StackAssets {
		add("stack-asset:"+v.Role, "stack-asset", v.Role, Artifact{MediaType: v.MediaType, Digest: v.Digest, Size: v.Size})
	}
	if v := b.Manifest.Runner; v != nil {
		if v.Binary != nil {
			add("runner:binary", "runner-binary", "runner", *v.Binary)
		}
		if v.Image != nil {
			add("runner:image", "runner-image", v.Image.Name, v.Image.Artifact)
		}
	}
	return out
}

// ExtractRunnerBinary streams the packaged runner binary to w. The manifest
// artifact is the packed OCI manifest; the binary itself is its single
// octet-stream layer.
func (b *Bundle) ExtractRunnerBinary(ctx context.Context, w io.Writer) error {
	if b.Manifest.Runner == nil || b.Manifest.Runner.Binary == nil {
		return fmt.Errorf("bundle has no runner binary")
	}
	desc := ocispec.Descriptor{MediaType: b.Manifest.Runner.Binary.MediaType, Digest: digest.Digest(b.Manifest.Runner.Binary.Digest), Size: b.Manifest.Runner.Binary.Size}
	var manifest ocispec.Manifest
	if err := fetchJSON(ctx, b.store, desc, &manifest); err != nil {
		return fmt.Errorf("read runner binary manifest: %w", err)
	}
	var layer *ocispec.Descriptor
	for i := range manifest.Layers {
		if manifest.Layers[i].MediaType == RunnerBinaryMediaType {
			layer = &manifest.Layers[i]
			break
		}
	}
	if layer == nil {
		return fmt.Errorf("runner binary manifest has no %s layer", RunnerBinaryMediaType)
	}
	r, err := b.store.Fetch(ctx, *layer)
	if err != nil {
		return fmt.Errorf("fetch runner binary layer: %w", err)
	}
	defer r.Close()
	verifier := layer.Digest.Verifier()
	n, err := io.Copy(w, io.TeeReader(io.LimitReader(r, layer.Size+1), verifier))
	if err != nil {
		return fmt.Errorf("copy runner binary: %w", err)
	}
	if n != layer.Size || !verifier.Verified() {
		return fmt.Errorf("runner binary does not match descriptor digest %s", layer.Digest)
	}
	return nil
}

func VerifyBlobs(dir string) error {
	ctx := context.Background()
	b, err := Open(ctx, dir)
	if err != nil {
		return err
	}
	roots := append([]ocispec.Descriptor{b.BundleDescriptor}, b.Roots...)
	traversed := make(map[traversalKey]int64)
	for _, root := range roots {
		if err := verifyBlobGraph(dir, root, traversed); err != nil {
			return err
		}
	}
	return nil
}

func verifyBlobGraph(dir string, desc ocispec.Descriptor, traversed map[traversalKey]int64) error {
	if err := validateDescriptor(desc); err != nil {
		return fmt.Errorf("invalid descriptor: %w", err)
	}
	key := traversalKey{digest: desc.Digest, mediaType: desc.MediaType}
	if size, ok := traversed[key]; ok {
		if size != desc.Size {
			return fmt.Errorf("conflicting size for %s: expected %d, got %d", desc.Digest, size, desc.Size)
		}
		return nil
	}
	traversed[key] = desc.Size

	metadata := recognizedMetadata(desc.MediaType)
	if metadata && desc.Size > DefaultMaxMetadataBytes {
		return fmt.Errorf("metadata blob %s size %d exceeds limit %d", desc.Digest, desc.Size, DefaultMaxMetadataBytes)
	}
	f, err := os.Open(filepath.Join(dir, "blobs", "sha256", desc.Digest.Encoded()))
	if err != nil {
		return fmt.Errorf("unable to open required blob %s: %w", desc.Digest, err)
	}
	h := sha256.New()
	var contents bytes.Buffer
	w := io.Writer(h)
	if metadata {
		w = io.MultiWriter(h, &contents)
	}
	n, copyErr := io.Copy(w, io.LimitReader(f, desc.Size+1))
	closeErr := f.Close()
	if copyErr != nil {
		return fmt.Errorf("unable to read blob %s: %w", desc.Digest, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("unable to close blob %s: %w", desc.Digest, closeErr)
	}
	if n != desc.Size {
		return fmt.Errorf("blob size mismatch for %s: expected %d, got %d", desc.Digest, desc.Size, n)
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != desc.Digest.Encoded() {
		return fmt.Errorf("blob digest mismatch for %s: got sha256:%s", desc.Digest, got)
	}
	if !metadata {
		return nil
	}
	children, err := Successors(desc.MediaType, contents.Bytes())
	if err != nil {
		return fmt.Errorf("parse %s: %w", desc.Digest, err)
	}
	for _, child := range children {
		if err := verifyBlobGraph(dir, child, traversed); err != nil {
			return err
		}
	}
	return nil
}

func recognizedMetadata(mediaType string) bool {
	switch mediaType {
	case ocispec.MediaTypeImageManifest,
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.oci.artifact.manifest.v1+json",
		ocispec.MediaTypeImageIndex,
		"application/vnd.docker.distribution.manifest.list.v2+json":
		return true
	default:
		return false
	}
}
