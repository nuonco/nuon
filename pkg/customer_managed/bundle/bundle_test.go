package customerbundle

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
)

func TestGenerateDeterministicLayoutAndCanonicalManifest(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	artifact := artifactFor(root)
	configDigest := digest.FromString("config").String()
	a := LogicalManifest{SchemaVersion: 1, Target: Target{OS: "linux", Architecture: "amd64"}, Components: []Component{{Name: "z", Type: "fixture", ConfigDigest: configDigest, Artifact: artifact}, {Name: "a", Type: "fixture", ConfigDigest: configDigest, Artifact: artifact}}, Actions: []Action{{Name: "2", ConfigDigest: configDigest}, {Name: "1", ConfigDigest: configDigest}}}
	b := LogicalManifest{SchemaVersion: 1, Target: Target{OS: "linux", Architecture: "amd64"}, Components: []Component{{Name: "a", Type: "fixture", ConfigDigest: configDigest, Artifact: artifact}, {Name: "z", Type: "fixture", ConfigDigest: configDigest, Artifact: artifact}}, Actions: []Action{{Name: "1", ConfigDigest: configDigest}, {Name: "2", ConfigDigest: configDigest}}}
	var first, second bytes.Buffer
	r1, err := Generate(ctx, &first, a, []Root{{Descriptor: root, Source: store}})
	require.NoError(t, err)
	r2, err := Generate(ctx, &second, b, []Root{{Descriptor: root, Source: store}})
	require.NoError(t, err)
	require.Equal(t, first.Bytes(), second.Bytes())
	require.Equal(t, r1, r2)
	sum := sha256.Sum256(first.Bytes())
	require.Equal(t, hex.EncodeToString(sum[:]), r1.TransportSHA256)

	files := untar(t, first.Bytes())
	require.JSONEq(t, `{"imageLayoutVersion":"1.0.0"}`, string(files["oci-layout"]))
	require.Contains(t, files, "index.json")
	require.Contains(t, files, "bundle-manifest.json")
	require.Contains(t, files, blobPath(root))
	require.Contains(t, string(files[blobPath(r1.ManifestDescriptor)]), `"components":[{"name":"a"`)
	var idx ocispec.Index
	require.NoError(t, json.Unmarshal(files["index.json"], &idx))
	require.Equal(t, r1.BundleDescriptor.Digest, idx.Manifests[0].Digest)
	require.Equal(t, root.Digest, idx.Manifests[1].Digest)
}

func TestGenerateVerifiesDescriptorClosure(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	var out bytes.Buffer
	_, err := Generate(ctx, &out, manifestFor(root), []Root{{Descriptor: root, Source: store}})
	require.NoError(t, err)
	files := untar(t, out.Bytes())
	require.Len(t, files, 7)
}

func TestGenerateIncludesQualificationDocuments(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	var out bytes.Buffer
	_, err := GenerateWithDocuments(ctx, &out, manifestFor(root), Documents{
		Provenance:          json.RawMessage(`{"app_config_id":"appcfg123"}`),
		QualificationReport: json.RawMessage(`{"warnings":["inline action is opaque"]}`),
		SourceArchive:       json.RawMessage(`{"files":{"policies/pass.rego":"package policies"}}`),
	}, []Root{{Descriptor: root, Source: store}})
	require.NoError(t, err)
	files := untar(t, out.Bytes())
	require.JSONEq(t, `{"app_config_id":"appcfg123"}`, string(files["bundle-provenance.json"]))
	require.JSONEq(t, `{"warnings":["inline action is opaque"]}`, string(files["qualification-report.json"]))
	require.JSONEq(t, `{"files":{"policies/pass.rego":"package policies"}}`, string(files["release-source.json"]))
}

func TestPlanEnvelopeRoundTrip(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	envelope := json.RawMessage(`{"version":"v0","org_id":"org","app_id":"app","steps":[]}`)
	var out bytes.Buffer
	_, err := GenerateWithDocuments(ctx, &out, manifestFor(root), Documents{PlanEnvelope: envelope}, []Root{{Descriptor: root, Source: store}})
	require.NoError(t, err)

	files := untar(t, out.Bytes())
	require.JSONEq(t, string(envelope), string(files["plan-envelope.json"]))

	dir := t.TempDir()
	_, err = Extract(dir, bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	b, err := Open(ctx, dir)
	require.NoError(t, err)
	require.JSONEq(t, string(envelope), string(b.PlanEnvelope))
}

func TestOpenWithoutPlanEnvelope(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	var out bytes.Buffer
	_, err := Generate(ctx, &out, manifestFor(root), []Root{{Descriptor: root, Source: store}})
	require.NoError(t, err)
	dir := t.TempDir()
	_, err = Extract(dir, bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	b, err := Open(ctx, dir)
	require.NoError(t, err)
	require.Empty(t, b.PlanEnvelope)
}

func TestGenerateRejectsInvalidQualificationDocument(t *testing.T) {
	_, err := GenerateWithDocuments(context.Background(), io.Discard, emptyManifest(), Documents{Provenance: json.RawMessage(`{`)}, nil)
	require.ErrorContains(t, err, "not valid JSON")
}

func TestGenerateRejectsMissingAndCorruptBlob(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	var manifest ocispec.Manifest
	r, err := store.Fetch(ctx, root)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(r).Decode(&manifest))
	require.NoError(t, r.Close())

	missing := memory.New()
	require.NoError(t, missing.Push(ctx, root, bytes.NewReader(mustJSON(t, manifest))))
	_, err = Generate(ctx, io.Discard, manifestFor(root), []Root{{Descriptor: root, Source: missing}})
	require.ErrorContains(t, err, "fetch")

	corrupt := corruptTarget{ReadOnlyTarget: store, descriptor: root}
	_, err = Generate(ctx, io.Discard, manifestFor(root), []Root{{Descriptor: root, Source: corrupt}})
	require.ErrorContains(t, err, "digest mismatch")
}

func TestGenerateEnforcesContentLimitsBeforeFetch(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)

	_, err := GenerateWithOptions(ctx, io.Discard, manifestFor(root), Documents{}, []Root{{Descriptor: root, Source: store}}, GenerateOptions{MaxBlobBytes: root.Size - 1})
	require.ErrorContains(t, err, "exceeds limit")

	_, err = GenerateWithOptions(ctx, io.Discard, manifestFor(root), Documents{}, []Root{{Descriptor: root, Source: store}}, GenerateOptions{MaxContentBytes: root.Size})
	require.ErrorContains(t, err, "bundle content size exceeds limit")
}

func TestGenerateBoundsReadByDeclaredSize(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	target := &oversizedTarget{ReadOnlyTarget: store, descriptor: root, contents: bytes.Repeat([]byte("x"), 1024*1024)}

	_, err := GenerateWithOptions(ctx, io.Discard, manifestFor(root), Documents{}, []Root{{Descriptor: root, Source: target}}, GenerateOptions{MaxContentBytes: root.Size})
	require.ErrorContains(t, err, "size mismatch")
	require.LessOrEqual(t, target.bytesRead, root.Size+1)
}

func TestCollectTraversesSameDigestForEachMediaType(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	out := make(map[digest.Digest][]byte)
	traversed := make(map[traversalKey]bool)
	var contentBytes int64
	leafDescriptor := root
	leafDescriptor.MediaType = "application/octet-stream"
	target := mediaTypeAgnosticTarget{ReadOnlyTarget: store, descriptor: root}

	require.NoError(t, collect(ctx, target, leafDescriptor, out, traversed, &contentBytes, GenerateOptions{}))
	require.Len(t, out, 1)
	require.NoError(t, collect(ctx, target, root, out, traversed, &contentBytes, GenerateOptions{}))
	require.Len(t, out, 2)

	conflicting := root
	conflicting.Size++
	err := collect(ctx, target, conflicting, out, traversed, &contentBytes, GenerateOptions{})
	require.ErrorContains(t, err, "conflicting size")
}

func TestGenerateReportsVerifiedBlobs(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	var verified []digest.Digest

	_, err := GenerateWithOptions(ctx, io.Discard, manifestFor(root), Documents{}, []Root{{Descriptor: root, Source: store}}, GenerateOptions{
		OnBlobVerified: func(desc ocispec.Descriptor) { verified = append(verified, desc.Digest) },
	})
	require.NoError(t, err)
	require.Contains(t, verified, root.Digest)
	require.Len(t, verified, 2)
}

func TestGenerateRejectsRootsThatDoNotMatchManifest(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)

	_, err := Generate(ctx, io.Discard, emptyManifest(), []Root{{Descriptor: root, Source: store}})
	require.ErrorContains(t, err, "undeclared bundle root")

	_, err = Generate(ctx, io.Discard, manifestFor(root), nil)
	require.ErrorContains(t, err, "is missing")

	manifest := manifestFor(root)
	manifest.Components = append(manifest.Components, manifest.Components[0])
	_, err = Generate(ctx, io.Discard, manifest, []Root{{Descriptor: root, Source: store}})
	require.ErrorContains(t, err, "duplicate bundle member logical key")
}

func TestGenerateRejectsMalformedDescriptorWithoutPanicking(t *testing.T) {
	bad := ocispec.Descriptor{MediaType: ocispec.MediaTypeImageManifest, Digest: "not-a-digest", Size: 1}
	manifest := emptyManifest()
	manifest.Components = []Component{{Name: "bad", Type: "fixture", ConfigDigest: digest.FromString("config").String(), Artifact: artifactFor(bad)}}
	require.NotPanics(t, func() {
		_, err := Generate(context.Background(), io.Discard, manifest, []Root{{Descriptor: bad, Source: memory.New()}})
		require.ErrorContains(t, err, "invalid digest")
	})
}

func TestManifestIdentityChangesWithPackagedRoot(t *testing.T) {
	ctx := context.Background()
	firstStore, firstRoot := fixtureWithData(t, ctx, []byte("first"))
	secondStore, secondRoot := fixtureWithData(t, ctx, []byte("second"))

	first, err := Generate(ctx, io.Discard, manifestFor(firstRoot), []Root{{Descriptor: firstRoot, Source: firstStore}})
	require.NoError(t, err)
	second, err := Generate(ctx, io.Discard, manifestFor(secondRoot), []Root{{Descriptor: secondRoot, Source: secondStore}})
	require.NoError(t, err)
	require.NotEqual(t, first.ManifestDescriptor.Digest, second.ManifestDescriptor.Digest)
}

func TestActionStepArtifactParticipatesInRootClosure(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	artifact := artifactFor(root)
	manifest := emptyManifest()
	manifest.Actions = []Action{{Name: "healthcheck", ConfigDigest: digest.FromString("action").String(), Steps: []Step{{Name: "run", Artifact: &artifact}}}}

	_, err := Generate(ctx, io.Discard, manifest, nil)
	require.ErrorContains(t, err, "root")
	_, err = Generate(ctx, io.Discard, manifest, []Root{{Descriptor: root, Source: store}})
	require.NoError(t, err)
}

func TestGenerateRejectsStructurallyInvalidArtifactManifest(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	contents := []byte(`{}`)
	root := descriptor("application/vnd.oci.artifact.manifest.v1+json", contents)
	require.NoError(t, store.Push(ctx, root, bytes.NewReader(contents)))

	_, err := Generate(ctx, io.Discard, manifestFor(root), []Root{{Descriptor: root, Source: store}})
	require.ErrorContains(t, err, "invalid OCI artifact manifest")
}

func TestPlatformMetadataIsPartOfManifestIdentity(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	withoutPlatform, err := Generate(ctx, io.Discard, manifestFor(root), []Root{{Descriptor: root, Source: store}})
	require.NoError(t, err)

	platformRoot := root
	platformRoot.Platform = &ocispec.Platform{OS: "linux", Architecture: "amd64"}
	manifest := manifestFor(platformRoot)
	manifest.Components[0].Artifact.PlatformOS = "linux"
	manifest.Components[0].Artifact.PlatformArchitecture = "amd64"
	withPlatform, err := Generate(ctx, io.Discard, manifest, []Root{{Descriptor: platformRoot, Source: store}})
	require.NoError(t, err)
	require.NotEqual(t, withoutPlatform.ManifestDescriptor.Digest, withPlatform.ManifestDescriptor.Digest)

	wrongPlatform := platformRoot
	wrongPlatform.Platform = &ocispec.Platform{OS: "linux", Architecture: "arm64"}
	_, err = Generate(ctx, io.Discard, manifest, []Root{{Descriptor: wrongPlatform, Source: store}})
	require.ErrorContains(t, err, "does not match")
}

type corruptTarget struct {
	oras.ReadOnlyTarget
	descriptor ocispec.Descriptor
}

type oversizedTarget struct {
	oras.ReadOnlyTarget
	descriptor ocispec.Descriptor
	contents   []byte
	bytesRead  int64
}

type mediaTypeAgnosticTarget struct {
	oras.ReadOnlyTarget
	descriptor ocispec.Descriptor
}

func (m mediaTypeAgnosticTarget) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	if desc.Digest == m.descriptor.Digest {
		return m.ReadOnlyTarget.Fetch(ctx, m.descriptor)
	}
	return m.ReadOnlyTarget.Fetch(ctx, desc)
}

func (o *oversizedTarget) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	if desc.Digest != o.descriptor.Digest {
		return o.ReadOnlyTarget.Fetch(ctx, desc)
	}
	return io.NopCloser(io.TeeReader(bytes.NewReader(o.contents), countingWriter{count: &o.bytesRead})), nil
}

type countingWriter struct{ count *int64 }

func (w countingWriter) Write(p []byte) (int, error) {
	*w.count += int64(len(p))
	return len(p), nil
}

func (c corruptTarget) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	if desc.Digest == c.descriptor.Digest {
		return io.NopCloser(bytes.NewReader(bytes.Repeat([]byte("x"), int(desc.Size)))), nil
	}
	return c.ReadOnlyTarget.Fetch(ctx, desc)
}

func fixture(t *testing.T, ctx context.Context) (*memory.Store, ocispec.Descriptor) {
	return fixtureWithData(t, ctx, []byte("artifact data"))
}

func fixtureWithData(t *testing.T, ctx context.Context, layerData []byte) (*memory.Store, ocispec.Descriptor) {
	store := memory.New()
	layer := descriptor("application/octet-stream", layerData)
	require.NoError(t, store.Push(ctx, layer, bytes.NewReader(layerData)))
	m := ocispec.Manifest{Versioned: specs.Versioned{SchemaVersion: 2}, MediaType: ocispec.MediaTypeImageManifest, Config: layer, Layers: []ocispec.Descriptor{layer}}
	b := mustJSON(t, m)
	root := descriptor(ocispec.MediaTypeImageManifest, b)
	require.NoError(t, store.Push(ctx, root, bytes.NewReader(b)))
	return store, root
}

func mustJSON(t *testing.T, value any) []byte {
	b, err := json.Marshal(value)
	require.NoError(t, err)
	return b
}
func blobPath(d ocispec.Descriptor) string {
	return "blobs/" + d.Digest.Algorithm().String() + "/" + d.Digest.Encoded()
}

func artifactFor(desc ocispec.Descriptor) Artifact {
	return Artifact{MediaType: desc.MediaType, Digest: desc.Digest.String(), Size: desc.Size}
}

func manifestFor(desc ocispec.Descriptor) LogicalManifest {
	manifest := emptyManifest()
	manifest.Components = []Component{{Name: "fixture", Type: "fixture", ConfigDigest: digest.FromString("config").String(), Artifact: artifactFor(desc)}}
	return manifest
}

func emptyManifest() LogicalManifest {
	return LogicalManifest{SchemaVersion: 1, Target: Target{OS: "linux", Architecture: "amd64"}}
}

func untar(t *testing.T, compressed []byte) map[string][]byte {
	zr, err := zstd.NewReader(bytes.NewReader(compressed))
	require.NoError(t, err)
	defer zr.Close()
	tr := tar.NewReader(zr)
	files := map[string][]byte{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		files[h.Name], err = io.ReadAll(tr)
		require.NoError(t, err)
	}
	return files
}

func TestTotalSizeSumsUniqueBlobs(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	layerA, err := oras.PushBytes(ctx, store, "application/octet-stream", bytes.Repeat([]byte("a"), 1000))
	require.NoError(t, err)
	layerB, err := oras.PushBytes(ctx, store, "application/octet-stream", bytes.Repeat([]byte("b"), 500))
	require.NoError(t, err)
	desc, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, "application/vnd.test.artifact.v1", oras.PackManifestOptions{Layers: []ocispec.Descriptor{layerA, layerB, layerA}})
	require.NoError(t, err)

	manifestBytes, err := io.ReadAll(mustFetch(t, ctx, store, desc))
	require.NoError(t, err)
	var m struct {
		Config ocispec.Descriptor `json:"config"`
	}
	require.NoError(t, json.Unmarshal(manifestBytes, &m))

	total, err := TotalSize(ctx, store, desc)
	require.NoError(t, err)
	require.Equal(t, desc.Size+m.Config.Size+1000+500, total)
}

func TestTotalSizeLeafBlobIsNotFetched(t *testing.T) {
	ctx := context.Background()
	leaf := ocispec.Descriptor{MediaType: "application/octet-stream", Digest: digest.FromString("never-fetched"), Size: 12345}
	total, err := TotalSize(ctx, memory.New(), leaf)
	require.NoError(t, err)
	require.Equal(t, int64(12345), total)
}

func mustFetch(t *testing.T, ctx context.Context, store oras.ReadOnlyTarget, desc ocispec.Descriptor) io.Reader {
	t.Helper()
	r, err := store.Fetch(ctx, desc)
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })
	return r
}
