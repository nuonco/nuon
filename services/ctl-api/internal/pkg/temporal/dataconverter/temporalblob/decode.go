package temporalblob

import (
	"context"
	"io"
	"strings"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"

	"github.com/pkg/errors"
)

func (d *dataConverter) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	startTime := time.Now()
	tags := []string{"format:temporalblob", "op:decode"}
	defer func() {
		duration := time.Since(startTime)
		d.mw.Incr("temporal.codec.incr", tags)
		d.mw.Timing("temporal.codec.duration", duration, tags)
	}()

	result := make([]*commonpb.Payload, len(payloads))
	for i, payload := range payloads {
		decoded, err := d.decodePayload(payload, tags)
		if err != nil {
			return nil, err
		}
		result[i] = decoded
	}

	return result, nil
}

func (d *dataConverter) decodePayload(payload *commonpb.Payload, tags []string) (*commonpb.Payload, error) {
	// Only decode payloads with our encoding
	if string(payload.Metadata[converter.MetadataEncoding]) != encoding {
		return payload, nil
	}

	if string(payload.Metadata["nuon/temporal-blob/enabled"]) != "true" {
		return payload, nil
	}

	blobID := string(payload.Data)

	// Check local cache first
	if data, ok := d.cache.Get(blobID); ok {
		d.mw.Incr("temporal.codec.blob.cache.hit", tags)
		return d.restorePayload(payload, data), nil
	}

	d.mw.Incr("temporal.codec.blob.cache.miss", tags)

	// Cache miss: download from S3
	s3Key := string(payload.Metadata["nuon/temporal-blob/s3_key"])
	if s3Key == "" {
		return nil, errors.New("temporal blob decode: missing s3_key in metadata")
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.TemporalBlobS3Timeout)
	defer cancel()

	reader, err := d.blobSvc.DownloadStream(ctx, s3Key)
	if err != nil {
		d.l.Error("error downloading temporal blob from S3", zap.Error(err), zap.String("s3_key", s3Key))
		d.mw.Incr("temporal.codec.blob.download.error", tags)
		return nil, errors.Wrap(err, "unable to download temporal blob from S3")
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		d.l.Error("error reading temporal blob data", zap.Error(err), zap.String("s3_key", s3Key))
		d.mw.Incr("temporal.codec.blob.download.error", tags)
		return nil, errors.Wrap(err, "unable to read temporal blob from S3")
	}

	d.mw.Incr("temporal.codec.blob.download.success", tags)
	d.mw.Gauge("temporal.codec.blob.download.size", float64(len(data)), tags)

	// Populate local cache for future reads
	if err := d.cache.Put(blobID, data); err != nil {
		d.l.Warn("error populating temporal blob cache", zap.Error(err), zap.String("blob_id", blobID))
	}

	return d.restorePayload(payload, data), nil
}

func (d *dataConverter) restorePayload(encoded *commonpb.Payload, data []byte) *commonpb.Payload {
	restored := &commonpb.Payload{
		Metadata: make(map[string][]byte),
		Data:     data,
	}

	// Copy non-codec metadata
	if encoded.Metadata != nil {
		for k, v := range encoded.Metadata {
			if k != converter.MetadataEncoding && !strings.HasPrefix(k, "nuon/temporal-blob/") {
				restored.Metadata[k] = v
			}
		}
	}

	// Restore original encoding
	if originalEncoding, ok := encoded.Metadata["nuon/temporal-blob/original-encoding"]; ok {
		restored.Metadata[converter.MetadataEncoding] = originalEncoding
	}

	return restored
}
