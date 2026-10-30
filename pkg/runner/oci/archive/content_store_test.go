package ociarchive

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/errdef"
)

func TestArchiveStoreVerifiesIngest(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name  string
		size  int64
		input string
		want  error
	}{
		{"wrong digest", 8, "tampered", content.ErrMismatchedDigest},
		{"truncated", 8, "short", io.ErrUnexpectedEOF},
		{"trailing", 8, "original+extra", content.ErrTrailingData},
		{"empty with bytes", 0, "original", content.ErrTrailingData},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestArchive(t, ctx)
			store := a.Ref().(oras.Target)
			desc := content.NewDescriptorFromBytes("test/blob", []byte("original"))
			desc.Size = tc.size
			require.ErrorIs(t, store.Push(ctx, desc, strings.NewReader(tc.input)), tc.want)
			exists, err := store.Exists(ctx, desc)
			require.NoError(t, err)
			require.False(t, exists)
			_, err = store.Fetch(ctx, desc)
			require.ErrorIs(t, err, errdef.ErrNotFound)
		})
	}
	for _, algorithm := range []digest.Algorithm{digest.SHA256, digest.SHA384, digest.SHA512} {
		for _, data := range [][]byte{nil, []byte("blob content")} {
			a := newTestArchive(t, ctx)
			store := a.Ref().(oras.Target)
			desc := v1.Descriptor{MediaType: "test/blob", Digest: algorithm.FromBytes(data), Size: int64(len(data))}
			require.NoError(t, store.Push(ctx, desc, bytes.NewReader(data)))
			require.ErrorIs(t, store.Push(ctx, desc, bytes.NewReader(data)), errdef.ErrAlreadyExists)
			got, err := content.FetchAll(ctx, store, desc)
			require.NoError(t, err)
			require.Equal(t, string(data), string(got))
		}
	}
}

func TestArchiveStoreRejectsInvalidDescriptors(t *testing.T) {
	ctx := context.Background()
	a := newTestArchive(t, ctx)
	store := a.Ref().(oras.Target)
	for _, desc := range []v1.Descriptor{
		{},
		{Digest: "sha256:../../canary", Size: 1},
		{Digest: digest.Digest("sha256:" + strings.Repeat("A", 64))},
		{Digest: digest.Digest("unsupported:" + strings.Repeat("a", 64))},
		{Digest: digest.FromString(""), Size: -1},
	} {
		require.Error(t, store.Push(ctx, desc, strings.NewReader("")))
		_, err := store.Fetch(ctx, desc)
		require.Error(t, err)
		_, err = store.Exists(ctx, desc)
		require.Error(t, err)
		require.Error(t, store.Tag(ctx, desc, "tag"))
	}
}

func TestArchiveStoreParallelCopy(t *testing.T) {
	ctx := context.Background()
	layers := make([]fixtureLayer, 64)
	for i := range layers {
		layers[i] = fixtureLayer{title: fmt.Sprintf("file-%d", i), data: []byte(fmt.Sprintf("content-%d", i%32))}
	}
	source := localFixture(t, layers...)
	a := newTestArchive(t, ctx)
	store := a.Ref().(oras.Target)
	const count = 8
	results := make(chan error, count)
	for i := range count {
		go func() {
			opts := oras.DefaultCopyOptions
			opts.Concurrency = 16
			_, err := oras.Copy(ctx, source, defaultLocalTag, store, fmt.Sprintf("copy-%d", i), opts)
			results <- err
		}()
	}
	for range count {
		require.NoError(t, <-results)
	}
	desc, err := source.Resolve(ctx, defaultLocalTag)
	require.NoError(t, err)
	for i := range count {
		got, err := store.Resolve(ctx, fmt.Sprintf("copy-%d", i))
		require.NoError(t, err)
		require.Equal(t, desc.Digest, got.Digest)
	}
	for i := range layers {
		got, err := afero.ReadFile(a.FS(), fmt.Sprintf("file-%d", i))
		require.NoError(t, err)
		require.Equal(t, layers[i].data, got)
	}
}
