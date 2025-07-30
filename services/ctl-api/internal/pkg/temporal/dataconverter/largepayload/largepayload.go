package largepayload

import (
	"context"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

var _ converter.PayloadCodec = (*dataConverter)(nil)

type dataConverter struct {
	cfg *internal.Config
	l   *zap.Logger
	db  *gorm.DB
}

func (d *dataConverter) Encode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))

	for i, payload := range payloads {
		// Skip if already encoded
		if string(payload.Metadata[converter.MetadataEncoding]) == "nuon/largepayload" {
			result[i] = payload
			continue
		}

		if len(payload.Data) < d.cfg.TemporalDataConverterLargePayloadSize {
			result[i] = payload
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		dbPayload := app.TemporalPayload{
			Contents: payload.Data,
		}
		if res := d.db.WithContext(ctx).Create(&dbPayload); res.Error != nil {
			d.l.Error("error encoding using large payload codec", zap.Error(res.Error))
			panic("error encoding" + res.Error.Error())
			return nil, errors.Wrap(res.Error, "unable to write temporal payload")
		}

		// Create new payload with compressed data
		result[i] = &commonpb.Payload{
			Metadata: map[string][]byte{
				converter.MetadataEncoding:  []byte("nuon/largepayload"),
				"nuon/largepayload/enabled": []byte("true"),
			},
			Data: []byte(dbPayload.ID),
		}
		// Preserve original metadata if exists
		for k, v := range payload.Metadata {
			if k != converter.MetadataEncoding {
				result[i].Metadata[k] = v
			} else {
				result[i].Metadata["nuon/large-payload/original-encoding"] = v
			}
		}
	}

	return result, nil
}

func (d *dataConverter) Decode(payloads []*commonpb.Payload) ([]*commonpb.Payload, error) {
	result := make([]*commonpb.Payload, len(payloads))

	for i, payload := range payloads {
		// Check if payload is larg payload encoded
		if string(payload.Metadata[converter.MetadataEncoding]) != "nuon/largepayload" {
			// Not large payload encoded, return as-is
			result[i] = payload
			continue
		}

		if string(payload.Metadata["nuon/largepayload/enabled"]) != "true" {
			result[i] = payload
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		var dbPayload app.TemporalPayload
		if res := d.db.WithContext(ctx).
			First(&dbPayload, "id = ?", string(payload.Data)); res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to get payload")
		}

		// Create new payload with decompressed data
		result[i] = &commonpb.Payload{
			Metadata: make(map[string][]byte),
			Data:     []byte(dbPayload.Contents),
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
		if originalEncoding, ok := payload.Metadata["nuon/large-payload/original-encoding"]; ok {
			result[i].Metadata[converter.MetadataEncoding] = originalEncoding
		}
	}

	return result, nil
}
