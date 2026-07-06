package blob

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"

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
	cache := "no"
	defer func() {
		tags := []string{"format:blob", "status:" + status, "cache:" + cache}
		d.mw.Incr("temporal.dataconverter.encode", tags)
		d.mw.Timing("temporal.dataconverter.encode.latency", time.Since(startTime), tags)
		if status == "success" {
			d.mw.Gauge("temporal.dataconverter.encode.size", float64(len(payload.Data)), tags)
		}
	}()

	// Content-addressed key: identical payloads map to the same blob, so repeated
	// encodes of the same data (fanned out across activities, or replayed across
	// cron ticks) dedupe to a single S3 object and a single cache entry.
	sum := sha256.Sum256(payload.Data)
	hexSum := hex.EncodeToString(sum[:])
	blobID := blobIDPrefix + hexSum
	checksum := "sha256:" + hexSum
	s3Key := d.cfg.TemporalBlobS3Prefix + blobID

	// A tracked cache entry means this content is already durable in S3 (we either
	// uploaded it here or fetched it from S3 during a prior decode), so we can skip
	// the upload and DB write.
	if d.cache.Has(blobID) {
		cache = "yes"
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), d.cfg.TemporalBlobS3Timeout)
		defer cancel()

		reader := strings.NewReader(string(payload.Data))
		if _, err := d.blobSvc.UploadStream(ctx, s3Key, reader); err != nil {
			status = "error"
			d.l.Error("error uploading blob to S3", zap.Error(err), zap.String("s3_key", s3Key))
			// Graceful degradation: return original payload
			return payload, nil
		}

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

		if err := d.cache.Put(blobID, payload.Data); err != nil {
			d.l.Warn("error writing to blob cache", zap.Error(err), zap.String("blob_id", blobID))
		}
	}

	// Build encoded payload
	encoded := &commonpb.Payload{
		Metadata: map[string][]byte{
			converter.MetadataEncoding: []byte(encoding),
			"nuon/blob/enabled":        []byte("true"),
			"nuon/blob/blob_id":        []byte(blobID),
			"nuon/blob/s3_key":         []byte(s3Key),
			"nuon/blob/checksum":       []byte(checksum),
			"nuon/blob/size":           []byte(strconv.Itoa(len(payload.Data))),
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
