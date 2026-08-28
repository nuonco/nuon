package customerbundle

import (
	"bytes"
	"context"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
)

func runnerBinaryFixture(t *testing.T, ctx context.Context, contents []byte) (*memory.Store, ocispec.Descriptor) {
	store := memory.New()
	layer, err := oras.PushBytes(ctx, store, RunnerBinaryMediaType, contents)
	require.NoError(t, err)
	desc, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, RunnerBinaryArtifactType, oras.PackManifestOptions{Layers: []ocispec.Descriptor{layer}})
	require.NoError(t, err)
	return store, desc
}

func manifestWithRunner(binary Artifact, image *Image) LogicalManifest {
	m := emptyManifest()
	m.Runner = &Runner{Version: "v1.2.3", SourceURL: "file:///runner", Binary: &binary, Image: image}
	return m
}

func TestRunnerBinaryRoundTrip(t *testing.T) {
	ctx := context.Background()
	contents := []byte("#!/bin/sh\necho runner\n")
	binStore, binRoot := runnerBinaryFixture(t, ctx, contents)
	imgStore, imgRoot := fixtureWithData(t, ctx, []byte("runner image layer"))

	manifest := manifestWithRunner(artifactFor(binRoot), &Image{Name: "runner", Repository: "public.ecr.aws/nuon/runner:v1.2.3", Artifact: artifactFor(imgRoot)})
	var out bytes.Buffer
	_, err := Generate(ctx, &out, manifest, []Root{
		{Descriptor: binRoot, Source: binStore},
		{Descriptor: imgRoot, Source: imgStore},
	})
	require.NoError(t, err)

	dir := t.TempDir()
	_, err = Extract(dir, bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	b, err := Open(ctx, dir)
	require.NoError(t, err)

	keys := map[string]bool{}
	for _, m := range b.Members() {
		keys[m.Key] = true
	}
	require.True(t, keys["runner:binary"])
	require.True(t, keys["runner:image"])

	var extracted bytes.Buffer
	require.NoError(t, b.ExtractRunnerBinary(ctx, &extracted))
	require.Equal(t, contents, extracted.Bytes())
}

func TestRunnerRequiresBinary(t *testing.T) {
	manifest := emptyManifest()
	manifest.Runner = &Runner{Version: "v1.2.3"}
	_, err := Generate(context.Background(), bytes.NewBuffer(nil), manifest, nil)
	require.ErrorContains(t, err, "runner binary artifact is required")
}

func TestRunnerImageRequiresRepository(t *testing.T) {
	ctx := context.Background()
	binStore, binRoot := runnerBinaryFixture(t, ctx, []byte("bin"))
	imgStore, imgRoot := fixtureWithData(t, ctx, []byte("img"))

	manifest := manifestWithRunner(artifactFor(binRoot), &Image{Name: "runner", Artifact: artifactFor(imgRoot)})
	_, err := Generate(ctx, bytes.NewBuffer(nil), manifest, []Root{
		{Descriptor: binRoot, Source: binStore},
		{Descriptor: imgRoot, Source: imgStore},
	})
	require.ErrorContains(t, err, "runner image repository is required")
}

func TestExtractRunnerBinaryWithoutRunner(t *testing.T) {
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
	require.ErrorContains(t, b.ExtractRunnerBinary(ctx, bytes.NewBuffer(nil)), "no runner binary")
}

func TestExtractRunnerBinaryRejectsWrongLayerMediaType(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	layer, err := oras.PushBytes(ctx, store, "application/x-unexpected", []byte("bin"))
	require.NoError(t, err)
	desc, err := oras.PackManifest(ctx, store, oras.PackManifestVersion1_1, RunnerBinaryArtifactType, oras.PackManifestOptions{Layers: []ocispec.Descriptor{layer}})
	require.NoError(t, err)

	manifest := manifestWithRunner(artifactFor(desc), nil)
	var out bytes.Buffer
	_, err = Generate(ctx, &out, manifest, []Root{{Descriptor: desc, Source: store}})
	require.NoError(t, err)
	dir := t.TempDir()
	_, err = Extract(dir, bytes.NewReader(out.Bytes()))
	require.NoError(t, err)
	b, err := Open(ctx, dir)
	require.NoError(t, err)
	require.ErrorContains(t, b.ExtractRunnerBinary(ctx, bytes.NewBuffer(nil)), "no application/octet-stream layer")
}
