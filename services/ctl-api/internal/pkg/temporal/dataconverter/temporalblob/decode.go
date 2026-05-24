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
	result := make([]*commonpb.Payload, len(payloads))
	for i, payload := range payloads {
		decoded, err := d.decodePayload(payload)
		if err != nil {
			return nil, err
		}
		result[i] = decoded
	}

	return result, nil
}

func (d *dataConverter) decodePayload(payload *commonpb.Payload) (*commonpb.Payload, error) {
	// Only decode payloads with our encoding
	if string(payload.Metadata[converter.MetadataEncoding]) != encoding {
		return payload, nil
	}

	if string(payload.Metadata["nuon/temporal-blob/enabled"]) != "true" {
		return payload, nil
	}

	startTime := time.Now()
	status := "success"
	cache := "no"
	var size float64
	defer func() {
		tags := []string{"format:temporalblob", "status:" + status, "cache:" + cache}
		d.mw.Incr("temporal.dataconverter.decode", tags)
		d.mw.Timing("temporal.dataconverter.decode.latency", time.Since(startTime), tags)
		if status == "success" {
			d.mw.Gauge("temporal.dataconverter.decode.size", size, tags)
		}
	}()

	blobID := string(payload.Data)

	// Check local cache first
	if data, ok := d.cache.Get(blobID); ok {
		cache = "yes"
		size = float64(len(data))
		return d.restorePayload(payload, data), nil
	}

	// Cache miss: download from S3
	s3Key := string(payload.Metadata["nuon/temporal-blob/s3_key"])
	if s3Key == "" {
		status = "error"
		return nil, errors.New("temporal blob decode: missing s3_key in metadata")
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.TemporalBlobS3Timeout)
	defer cancel()

	reader, err := d.blobSvc.DownloadStream(ctx, s3Key)
	if err != nil {
		status = "error"
		d.l.Error("error downloading temporal blob from S3", zap.Error(err), zap.String("s3_key", s3Key))
		return nil, errors.Wrap(err, "unable to download temporal blob from S3")
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		status = "error"
		d.l.Error("error reading temporal blob data", zap.Error(err), zap.String("s3_key", s3Key))
		return nil, errors.Wrap(err, "unable to read temporal blob from S3")
	}

	size = float64(len(data))
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
