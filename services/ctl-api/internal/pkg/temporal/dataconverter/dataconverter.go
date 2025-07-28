package dataconverter

import (
	"context"
	"fmt"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	pkgdataconverter "github.com/powertoolsdev/mono/pkg/temporal/dataconverter"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type Params struct {
	fx.In

	DB  *gorm.DB `name:"psql"`
	Cfg *internal.Config
	L   *zap.Logger
}

type dataConverter struct {
	base converter.DataConverter
	db   *gorm.DB
	cfg  *internal.Config
	l    *zap.Logger
}

// ToPayload converts single value to payload.
func (c *dataConverter) ToPayload(value any) (*commonpb.Payload, error) {
	payload, err := c.base.ToPayload(value)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get payload")
	}

	// anything less than 128KB we pass to the server
	if len(payload.Data) < 128*1024 {
		return payload, nil
	}

	c.l.Info("encoding using dc")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	dbPayload := app.TemporalPayload{
		Contents: payload.Data,
	}
	if res := c.db.WithContext(ctx).Create(&dbPayload); res.Error != nil {
		c.l.Error("error encoding using dc", zap.Error(res.Error))
		return nil, errors.Wrap(err, "unable to write temporal payload")
	}

	return &commonpb.Payload{
		Metadata: map[string][]byte{
			"db_temporal_payload": []byte("true"),
			"encoding":            []byte(c.Encoding()),
		},
		Data: []byte(dbPayload.ID),
	}, nil
}

// FromPayload converts single value from payload.
func (c *dataConverter) FromPayload(payload *commonpb.Payload, valuePtr any) error {
	_, isDB := payload.Metadata["db_temporal_payload"]
	if !isDB {
		return c.base.FromPayload(payload, valuePtr)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	var dbPayload app.TemporalPayload
	if res := c.db.WithContext(ctx).
		First(&dbPayload, "id = ?", string(payload.Data)); res.Error != nil {
		return errors.Wrap(res.Error, "unable to get payload")
	}

	return c.base.FromPayload(&commonpb.Payload{
		Metadata: map[string][]byte{
			"encoding": []byte(c.Encoding()),
		},
		Data: dbPayload.Contents,
	}, valuePtr)
}

// ToString converts payload object into human readable string.
func (c *dataConverter) ToString(payload *commonpb.Payload) string {
	return string(payload.GetData())
}

// Encoding returns MetadataEncodingJSON.
func (c *dataConverter) Encoding() string {
	return fmt.Sprintf("json/temporaljson")
}

var _ converter.PayloadConverter = (*dataConverter)(nil)

func New(params Params) converter.DataConverter {
	dc := &dataConverter{
		base: pkgdataconverter.NewJSONConverter(),
		db:   params.DB,
		cfg:  params.Cfg,
		l:    params.L,
	}

	// note: the order of this is inmportant, to ensure we also support protobufs as well.
	return converter.NewCompositeDataConverter(
		dc,
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		converter.NewProtoJSONPayloadConverter(),
		converter.NewProtoPayloadConverter(),
		converter.NewJSONPayloadConverter(),
	)
}
