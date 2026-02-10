package s3payload

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

var _ converter.PayloadCodec = (*dataConverter)(nil)

type dataConverter struct {
	cfg     *internal.Config
	l       *zap.Logger
	blobSvc blobstore.Service
	mw      metrics.Writer
}

func (d *dataConverter) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	startTime := time.Now()
	tags := []string{"format:s3payload", "op:encode"}
	defer func() {
		duration := time.Since(startTime)
		d.mw.Incr("temporal.codec.incr", tags)
		d.mw.Timing("temporal.codec.duration", duration, tags)
	}()

	result := make([]*commonpb.Payload, len(payloads))

	for i, payload := range payloads {
		// Skip if already encoded
		if string(payload.Metadata[converter.MetadataEncoding]) == "nuon/s3payload" {
			result[i] = payload
			continue
		}

		// Skip if payload is below threshold
		if len(payload.Data) < d.cfg.TemporalDataConverterLargePayloadSize {
			result[i] = payload
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*30) // Longer timeout for S3 upload
		defer cancel()

		// Extract org ID from context (if available)
		// Note: Temporal workers may not have org context, so we'll use a default org path
		orgID := "temporal"
		if ctxOrgID, err := cctx.OrgIDFromContext(ctx); err == nil && ctxOrgID != "" {
			orgID = ctxOrgID
		}

		// Generate blob ID
		blobID := domains.NewBlobID()

		// Construct S3 key: org_id/blob_id
		s3Key := fmt.Sprintf("%s/%s", orgID, blobID)

		// Upload to S3 with streaming
		reader := strings.NewReader(string(payload.Data))
		checksum, err := d.blobSvc.UploadStream(ctx, s3Key, reader)
		if err != nil {
			d.l.Error("error encoding using s3 payload codec", zap.Error(err), zap.String("s3_key", s3Key))
			// Return original payload on error (graceful degradation)
			result[i] = payload
			d.mw.Incr("temporal.codec.s3.upload.error", tags)
			continue
		}

		d.mw.Incr("temporal.codec.s3.upload.success", tags)
		d.mw.Gauge("temporal.codec.s3.upload.size", float64(len(payload.Data)), tags)

		// Create new payload with S3 key
		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				converter.MetadataEncoding: []byte("nuon/s3payload"),
				"nuon/s3payload/enabled":   []byte("true"),
				"nuon/s3payload/s3_key":    []byte(s3Key),
				"nuon/s3payload/checksum":  []byte(checksum),
				"nuon/s3payload/size":      []byte(fmt.Sprintf("%d", len(payload.Data))),
			},
			Data: []byte(s3Key), // Store S3 key as payload data
		}

		// Preserve original metadata if exists
		for k, v := range payload.Metadata {
			if k != converter.MetadataEncoding {
				result[i].Metadata[k] = v
			} else {
				result[i].Metadata["nuon/s3payload/original-encoding"] = v
			}
		}
	}

	return result, nil
}

func (d *dataConverter) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	startTime := time.Now()
	tags := []string{"format:s3payload", "op:decode"}
	defer func() {
		duration := time.Since(startTime)
		d.mw.Incr("temporal.codec.incr", tags)
		d.mw.Timing("temporal.codec.duration", duration, tags)
	}()

	result := make([]*commonpb.Payload, len(payloads))

	for i, payload := range payloads {
		// Check if payload is S3 payload encoded
		if string(payload.Metadata[converter.MetadataEncoding]) != "nuon/s3payload" {
			// Not S3 payload encoded, return as-is
			result[i] = payload
			continue
		}

		if string(payload.Metadata["nuon/s3payload/enabled"]) != "true" {
			result[i] = payload
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*30) // Longer timeout for S3 download
		defer cancel()

		// Get S3 key from payload data
		s3Key := string(payload.Data)

		// Download from S3 with streaming
		reader, err := d.blobSvc.DownloadStream(ctx, s3Key)
		if err != nil {
			d.l.Error("error decoding using s3 payload codec", zap.Error(err), zap.String("s3_key", s3Key))
			d.mw.Incr("temporal.codec.s3.download.error", tags)
			return nil, errors.Wrap(err, "unable to download payload from S3")
		}
		defer reader.Close()

		// Read data from S3
		data, err := io.ReadAll(reader)
		if err != nil {
			d.l.Error("error reading s3 payload data", zap.Error(err), zap.String("s3_key", s3Key))
			d.mw.Incr("temporal.codec.s3.download.error", tags)
			return nil, errors.Wrap(err, "unable to read payload from S3")
		}

		d.mw.Incr("temporal.codec.s3.download.success", tags)
		d.mw.Gauge("temporal.codec.s3.download.size", float64(len(data)), tags)

		// Create new payload with decompressed data
		result[i] = &commonpb.Payload{
			Metadata: make(map[string][]byte),
			Data:     data,
		}

		// Copy all metadata except the encoding
		if payload.Metadata != nil {
			for k, v := range payload.Metadata {
				if k != converter.MetadataEncoding && !strings.HasPrefix(k, "nuon/s3payload/") {
					result[i].Metadata[k] = v
				}
			}
		}

		// Restore original encoding if it was preserved
		if originalEncoding, ok := payload.Metadata["nuon/s3payload/original-encoding"]; ok {
			result[i].Metadata[converter.MetadataEncoding] = originalEncoding
		}
	}

	return result, nil
}
