package temporalblob

import (
	"context"
	"fmt"
	"strings"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (d *dataConverter) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	startTime := time.Now()
	tags := []string{"format:temporalblob", "op:encode"}
	defer func() {
		duration := time.Since(startTime)
		d.mw.Incr("temporal.codec.incr", tags)
		d.mw.Timing("temporal.codec.duration", duration, tags)
	}()

	result := make([]*commonpb.Payload, len(payloads))
	for i, payload := range payloads {
		encoded, err := d.encodePayload(payload, tags)
		if err != nil {
			return nil, err
		}
		result[i] = encoded
	}

	return result, nil
}

func (d *dataConverter) encodePayload(payload *commonpb.Payload, tags []string) (*commonpb.Payload, error) {
	// Skip if already encoded
	if string(payload.Metadata[converter.MetadataEncoding]) == encoding {
		return payload, nil
	}

	// Skip if payload is below threshold
	fmt.Println("JM_TEST", d.cfg.TemporalDataConverterLargePayloadSize)
	if len(payload.Data) < d.cfg.TemporalDataConverterLargePayloadSize {
		return payload, nil
	}

	// Skip if encoding is disabled (toggle is set to "db")
	if !d.encodeEnabled {
		return payload, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.TemporalBlobS3Timeout)
	defer cancel()

	// Generate blob ID and S3 key
	blobID := domains.NewTemporalBlob()
	s3Key := d.cfg.TemporalBlobS3Prefix + blobID

	// Upload to S3
	reader := strings.NewReader(string(payload.Data))
	checksum, err := d.blobSvc.UploadStream(ctx, s3Key, reader)
	if err != nil {
		d.l.Error("error uploading temporal blob to S3", zap.Error(err), zap.String("s3_key", s3Key))
		d.mw.Incr("temporal.codec.blob.upload.error", tags)
		// Graceful degradation: return original payload
		return payload, nil
	}

	d.mw.Incr("temporal.codec.blob.upload.success", tags)
	d.mw.Gauge("temporal.codec.blob.upload.size", float64(len(payload.Data)), tags)

	// Write pointer row to DB
	dbRecord := app.TemporalBlob{
		S3Key:    s3Key,
		Checksum: checksum,
		Size:     int64(len(payload.Data)),
	}
	if res := d.db.WithContext(ctx).Create(&dbRecord); res.Error != nil {
		d.l.Error("error writing temporal blob record", zap.Error(res.Error), zap.String("s3_key", s3Key))
		// S3 upload succeeded but DB write failed; payload is in S3 and can be recovered
		// Still return the encoded payload since the s3_key is in metadata
	}

	// Write to local file cache
	if err := d.cache.Put(blobID, payload.Data); err != nil {
		d.l.Warn("error writing to temporal blob cache", zap.Error(err), zap.String("blob_id", blobID))
	}

	// Build encoded payload
	encoded := &commonpb.Payload{
		Metadata: map[string][]byte{
			converter.MetadataEncoding:    []byte(encoding),
			"nuon/temporal-blob/enabled":  []byte("true"),
			"nuon/temporal-blob/blob_id":  []byte(blobID),
			"nuon/temporal-blob/s3_key":   []byte(s3Key),
			"nuon/temporal-blob/checksum": []byte(checksum),
			"nuon/temporal-blob/size":     []byte(fmt.Sprintf("%d", len(payload.Data))),
		},
		Data: []byte(blobID),
	}

	// Preserve original metadata
	for k, v := range payload.Metadata {
		if k != converter.MetadataEncoding {
			encoded.Metadata[k] = v
		} else {
			encoded.Metadata["nuon/temporal-blob/original-encoding"] = v
		}
	}

	return encoded, nil
}
