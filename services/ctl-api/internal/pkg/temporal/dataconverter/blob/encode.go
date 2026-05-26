package blob

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
	result := make([]*commonpb.Payload, len(payloads))
	for i, payload := range payloads {
		encoded, err := d.encodePayload(payload)
		if err != nil {
			return nil, err
		}
		result[i] = encoded
	}

	return result, nil
}

func (d *dataConverter) encodePayload(payload *commonpb.Payload) (*commonpb.Payload, error) {
	// Skip if already encoded
	if string(payload.Metadata[converter.MetadataEncoding]) == encoding {
		return payload, nil
	}

	// Skip if payload is below threshold
	if len(payload.Data) < d.cfg.TemporalDataConverterLargePayloadSize {
		return payload, nil
	}

	// Skip if encoding is disabled (toggle is set to "db")
	if !d.encodeEnabled {
		return payload, nil
	}

	startTime := time.Now()
	status := "success"
	defer func() {
		tags := []string{"format:blob", "status:" + status, "cache:no"}
		d.mw.Incr("temporal.dataconverter.encode", tags)
		d.mw.Timing("temporal.dataconverter.encode.latency", time.Since(startTime), tags)
		if status == "success" {
			d.mw.Gauge("temporal.dataconverter.encode.size", float64(len(payload.Data)), tags)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), d.cfg.TemporalBlobS3Timeout)
	defer cancel()

	// Generate blob ID and S3 key
	blobID := domains.NewTemporalBlob()
	s3Key := d.cfg.TemporalBlobS3Prefix + blobID

	// Upload to S3
	reader := strings.NewReader(string(payload.Data))
	checksum, err := d.blobSvc.UploadStream(ctx, s3Key, reader)
	if err != nil {
		status = "error"
		d.l.Error("error uploading blob to S3", zap.Error(err), zap.String("s3_key", s3Key))
		// Graceful degradation: return original payload
		return payload, nil
	}

	// Write pointer row to DB
	dbRecord := app.TemporalBlob{
		S3Key:    s3Key,
		Checksum: checksum,
		Size:     int64(len(payload.Data)),
	}
	if res := d.db.WithContext(ctx).Create(&dbRecord); res.Error != nil {
		d.l.Error("error writing blob record", zap.Error(res.Error), zap.String("s3_key", s3Key))
		// S3 upload succeeded but DB write failed; payload is in S3 and can be recovered
		// Still return the encoded payload since the s3_key is in metadata
	}

	// Write to local file cache
	if err := d.cache.Put(blobID, payload.Data); err != nil {
		d.l.Warn("error writing to blob cache", zap.Error(err), zap.String("blob_id", blobID))
	}

	// Build encoded payload
	encoded := &commonpb.Payload{
		Metadata: map[string][]byte{
			converter.MetadataEncoding: []byte(encoding),
			"nuon/blob/enabled":        []byte("true"),
			"nuon/blob/blob_id":        []byte(blobID),
			"nuon/blob/s3_key":         []byte(s3Key),
			"nuon/blob/checksum":       []byte(checksum),
			"nuon/blob/size":           []byte(fmt.Sprintf("%d", len(payload.Data))),
		},
		Data: []byte(blobID),
	}

	// Preserve original metadata
	for k, v := range payload.Metadata {
		if k != converter.MetadataEncoding {
			encoded.Metadata[k] = v
		} else {
			encoded.Metadata["nuon/blob/original-encoding"] = v
		}
	}

	return encoded, nil
}
