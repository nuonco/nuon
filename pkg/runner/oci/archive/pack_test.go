package ociarchive

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"oras.land/oras-go/v2/content"
)

func packFixture(t *testing.T, count int) ocispec.Manifest {
	t.Helper()

	src := t.TempDir()
	files := make([]FileRef, 0, count)
	for i := range count {
		files = append(files, writeFile(t, src, fmt.Sprintf("f%d.tf", i), fmt.Sprintf("# %d\n", i), 0o644))
	}

	ctx := context.Background()
	a := New()
	require.NoError(t, a.Initialize(ctx))
	t.Cleanup(func() { _ = a.Cleanup(ctx) })

	require.NoError(t, a.Pack(ctx, zap.NewNop(), files))

	desc, err := a.store.Resolve(ctx, defaultLocalTag)
	require.NoError(t, err)

	byts, err := content.FetchAll(ctx, a.store, desc)
	require.NoError(t, err)

	var m ocispec.Manifest
	require.NoError(t, json.Unmarshal(byts, &m))
	return m
}

func TestPackKeepsPerFileLayersBelowThreshold(t *testing.T) {
	m := packFixture(t, 10)

	require.Len(t, m.Layers, 10)
	for _, layer := range m.Layers {
		require.NotEqual(t, tarballMediaType, layer.MediaType)
	}
}

func TestPackCollapsesLargeTreeIntoOneLayer(t *testing.T) {
	m := packFixture(t, maxPerFileLayers+1)

	require.Len(t, m.Layers, 1)
	require.Equal(t, tarballMediaType, m.Layers[0].MediaType)
	require.Equal(t, tarballName, m.Layers[0].Annotations[ocispec.AnnotationTitle])
}
