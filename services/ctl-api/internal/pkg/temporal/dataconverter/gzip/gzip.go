package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"go.uber.org/zap"

	"go.temporal.io/sdk/converter"

	commonpb "go.temporal.io/api/common/v1"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

var _ converter.PayloadCodec = (*dataConverter)(nil)

type dataConverter struct {
	cfg *internal.Config
	l   *zap.Logger
}

func (d *dataConverter) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))

	for i, payload := range payloads {
		if string(payload.Metadata[converter.MetadataEncoding]) == "binary/gzip" {
			result[i] = payload
			continue
		}

		// Compress the payload data
		var buf bytes.Buffer
		gzipWriter := gzip.NewWriter(&buf)

		if _, err := gzipWriter.Write(payload.Data); err != nil {
			return nil, fmt.Errorf("failed to write gzip data: %w", err)
		}

		if err := gzipWriter.Close(); err != nil {
			return nil, fmt.Errorf("failed to close gzip writer: %w", err)
		}

		// Create new payload with compressed data
		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				converter.MetadataEncoding: []byte("binary/gzip"),
			},
			Data: buf.Bytes(),
		}

		// Preserve original metadata if exists
		for k, v := range payload.Metadata {
			if k != converter.MetadataEncoding {
				result[i].Metadata[k] = v
			} else {
				result[i].Metadata["original-encoding"] = v
			}
		}
	}

	return result, nil
}

func (d *dataConverter) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))

	for i, payload := range payloads {
		// Check if payload is gzip encoded
		if string(payload.Metadata[converter.MetadataEncoding]) != "binary/gzip" {
			// Not gzip encoded, return as-is
			result[i] = payload
			continue
		}

		// Decompress the payload data
		reader := bytes.NewReader(payload.Data)
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()

		decompressed, err := io.ReadAll(gzipReader)
		if err != nil {
			return nil, fmt.Errorf("failed to read decompressed data: %w", err)
		}

		// Create new payload with decompressed data
		result[i] = &commonpb.Payload{
			Metadata: make(map[string][]byte),
			Data:     decompressed,
		}

		// Copy all metadata except the encoding
		if payload.Metadata != nil {
			for k, v := range payload.Metadata {
				if k != converter.MetadataEncoding {
					result[i].Metadata[k] = v
				}
			}
		}

		// Restore original encoding if it was preserved
		if originalEncoding, ok := payload.Metadata["original-encoding"]; ok {
			result[i].Metadata[converter.MetadataEncoding] = originalEncoding
		}
	}

	return result, nil
}
