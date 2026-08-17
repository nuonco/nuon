package bundle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestBlobSinkAndReassembleRoundTrip(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	sunk := map[digest.Digest][]byte{}
	var original bytes.Buffer
	result, err := GenerateWithOptions(ctx, &original, manifestFor(root), Documents{
		Provenance:          json.RawMessage(`{"source":"test"}`),
		QualificationReport: json.RawMessage(`{"qualified":true}`),
		PlanEnvelope:        json.RawMessage(`{"schema_version":1}`),
	}, []Root{{Descriptor: root, Source: store}}, GenerateOptions{
		BlobSink: func(dgst digest.Digest, data []byte) error {
			sunk[dgst] = append([]byte(nil), data...)
			return nil
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.Index)
	require.NotEmpty(t, sunk)

	files := untar(t, original.Bytes())
	blobCount := 0
	for name := range files {
		if len(name) > len("blobs/") && name[:len("blobs/")] == "blobs/" {
			blobCount++
		}
	}
	require.Equal(t, blobCount, len(sunk), "sink must receive exactly the archive's blobs")
	require.Equal(t, json.RawMessage(files["index.json"]), result.Index)

	var reassembled bytes.Buffer
	checksum, err := Reassemble(ctx, &reassembled, result.Index, func(_ context.Context, dgst digest.Digest) ([]byte, error) {
		data, ok := sunk[dgst]
		if !ok {
			return nil, fmt.Errorf("blob %s not found", dgst)
		}
		return data, nil
	})
	require.NoError(t, err)
	require.Equal(t, result.TransportSHA256, checksum)
	require.Equal(t, original.Bytes(), reassembled.Bytes())
}

func TestReassembleRejectsCorruptBlob(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	sunk := map[digest.Digest][]byte{}
	result, err := GenerateWithOptions(ctx, bytes.NewBuffer(nil), manifestFor(root), Documents{}, []Root{{Descriptor: root, Source: store}}, GenerateOptions{
		BlobSink: func(dgst digest.Digest, data []byte) error {
			sunk[dgst] = data
			return nil
		},
	})
	require.NoError(t, err)

	_, err = Reassemble(ctx, bytes.NewBuffer(nil), result.Index, func(_ context.Context, dgst digest.Digest) ([]byte, error) {
		return []byte("corrupt"), nil
	})
	require.ErrorContains(t, err, "mismatch")

	_, err = Reassemble(ctx, bytes.NewBuffer(nil), []byte(`{`), func(_ context.Context, dgst digest.Digest) ([]byte, error) {
		return sunk[dgst], nil
	})
	require.ErrorContains(t, err, "parse bundle index")
}

func TestBlobSinkFailureAbortsGenerate(t *testing.T) {
	ctx := context.Background()
	store, root := fixture(t, ctx)
	_, err := GenerateWithOptions(ctx, bytes.NewBuffer(nil), manifestFor(root), Documents{}, []Root{{Descriptor: root, Source: store}}, GenerateOptions{
		BlobSink: func(digest.Digest, []byte) error { return fmt.Errorf("sink unavailable") },
	})
	require.ErrorContains(t, err, "sink unavailable")
}
