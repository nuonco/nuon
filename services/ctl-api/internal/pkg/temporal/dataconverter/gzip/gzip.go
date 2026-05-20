package gzip

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"
	"time"

	"go.uber.org/zap"

	"go.temporal.io/sdk/converter"

	commonpb "go.temporal.io/api/common/v1"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

var _ converter.PayloadCodec = (*dataConverter)(nil)

var gzipWriterPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

type dataConverter struct {
	cfg *internal.Config
	l   *zap.Logger
	mw  metrics.Writer
}

func (d *dataConverter) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	startTime := time.Now()
	tags := []string{"format:gzip", "op:encode"}
	defer func() {
		duration := time.Time.Sub(time.Now(), startTime)
		d.mw.Incr("temporal.codec.incr", tags)
		d.mw.Timing("temporal.codec.duration", duration, tags)
	}()

	result := make([]*commonpb.Payload, len(payloads))

	for i, payload := range payloads {
		if string(payload.Metadata[converter.MetadataEncoding]) == "binary/gzip" {
			result[i] = payload
			continue
		}

		var buf bytes.Buffer
		buf.Grow(len(payload.Data))

		gzipWriter := gzipWriterPool.Get().(*gzip.Writer)
		gzipWriter.Reset(&buf)

		_, err := gzipWriter.Write(payload.Data)
		if err != nil {
			gzipWriterPool.Put(gzipWriter)
			return nil, fmt.Errorf("failed to write gzip data: %w", err)
		}

		if err := gzipWriter.Close(); err != nil {
			gzipWriterPool.Put(gzipWriter)
			return nil, fmt.Errorf("failed to close gzip writer: %w", err)
		}

		gzipWriterPool.Put(gzipWriter)

		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				converter.MetadataEncoding: []byte("binary/gzip"),
			},
			Data: buf.Bytes(),
		}

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
	startTime := time.Now()
	tags := []string{"format:gzip", "op:decode"}
	defer func() {
		duration := time.Time.Sub(time.Now(), startTime)
		d.mw.Incr("temporal.codec.incr", tags)
		d.mw.Timing("temporal.codec.duration", duration, tags)
	}()

	result := make([]*commonpb.Payload, len(payloads))

	for i, payload := range payloads {
		if string(payload.Metadata[converter.MetadataEncoding]) != "binary/gzip" {
			result[i] = payload
			continue
		}

		reader := bytes.NewReader(payload.Data)
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}

		decompressed, err := io.ReadAll(gzipReader)
		gzipReader.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read decompressed data: %w", err)
		}

		result[i] = &commonpb.Payload{
			Metadata: make(map[string][]byte),
			Data:     decompressed,
		}

		if payload.Metadata != nil {
			for k, v := range payload.Metadata {
				if k != converter.MetadataEncoding {
					result[i].Metadata[k] = v
				}
			}
		}

		if originalEncoding, ok := payload.Metadata["original-encoding"]; ok {
			result[i].Metadata[converter.MetadataEncoding] = originalEncoding
		}
	}

	return result, nil
}
