package blob

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
	enc := string(payload.Metadata[converter.MetadataEncoding])

	// Handle current encoding and legacy encodings
	var metadataPrefix string
	switch enc {
	case encoding: // "nuon/blob"
		metadataPrefix = "nuon/blob/"
	case "nuon/temporal-blob":
		metadataPrefix = "nuon/temporal-blob/"
	default:
		return payload, nil
	}

	if string(payload.Metadata[metadataPrefix+"enabled"]) != "true" {
		return payload, nil
	}

	startTime := time.Now()
	status := "success"
	cache := "no"
	var size float64
	defer func() {
		tags := []string{"format:blob", "status:" + status, "cache:" + cache}
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
		return d.restorePayload(payload, data, metadataPrefix), nil
	}

	// Cache miss: download from S3
	s3Key := string(payload.Metadata[metadataPrefix+"s3_key"])
	if s3Key == "" {
		status = "error"
		return nil, errors.New("blob decode: missing s3_key in metadata")
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.TemporalBlobS3Timeout)
	defer cancel()

	reader, err := d.blobSvc.DownloadStream(ctx, s3Key)
	if err != nil {
		status = "error"
		d.l.Error("error downloading blob from S3", zap.Error(err), zap.String("s3_key", s3Key))
		return nil, errors.Wrap(err, "unable to download blob from S3")
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		status = "error"
		d.l.Error("error reading blob data", zap.Error(err), zap.String("s3_key", s3Key))
		return nil, errors.Wrap(err, "unable to read blob from S3")
	}

	size = float64(len(data))

	if err := d.cache.Put(blobID, data); err != nil {
		d.l.Warn("error writing decoded blob to cache", zap.Error(err), zap.String("blob_id", blobID))
	}

	return d.restorePayload(payload, data, metadataPrefix), nil
}

func (d *dataConverter) restorePayload(encoded *commonpb.Payload, data []byte, metadataPrefix string) *commonpb.Payload {
	restored := &commonpb.Payload{
		Metadata: make(map[string][]byte),
		Data:     data,
	}

	// Copy non-codec metadata
	if encoded.Metadata != nil {
		for k, v := range encoded.Metadata {
			if k != converter.MetadataEncoding && !strings.HasPrefix(k, metadataPrefix) {
				restored.Metadata[k] = v
			}
		}
	}

	// Restore original encoding
	if originalEncoding, ok := encoded.Metadata[metadataPrefix+"original-encoding"]; ok {
		restored.Metadata[converter.MetadataEncoding] = originalEncoding
	}

	return restored
}
