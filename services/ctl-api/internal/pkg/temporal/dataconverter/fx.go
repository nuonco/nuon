package dataconverter

import (
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	pkgdataconverter "github.com/nuonco/nuon/pkg/temporal/dataconverter"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

type Params struct {
	fx.In

	DB  *gorm.DB `name:"psql"`
	Cfg *internal.Config
	L   *zap.Logger

	Gzip            converter.PayloadCodec `name:"gzip"`
	LargePayload    converter.PayloadCodec `name:"largepayload"`
	Blob            converter.PayloadCodec `name:"blob"`
	SignalConverter *signaldb.PayloadConverter
}

func New(params Params) converter.DataConverter {
	// NOTE(jm): make this an FX dependency
	dc := pkgdataconverter.NewJSONConverter()

	cdc := converter.NewCompositeDataConverter(
		params.SignalConverter,
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		dc,
	)

	return workflow.DataConverterWithoutDeadlockDetection(converter.NewCodecDataConverter(cdc,
		params.Blob,         // S3 blob codec (encode when toggle=blob, always decode)
		params.LargePayload, // Legacy DB (encode when toggle=db, always decode)
		params.Gzip,         // Compression
	))
}
