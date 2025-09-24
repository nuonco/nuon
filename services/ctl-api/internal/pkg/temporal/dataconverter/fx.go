package dataconverter

import (
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	pkgdataconverter "github.com/powertoolsdev/mono/pkg/temporal/dataconverter"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	signaldb "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/signal/db"
)

type Params struct {
	fx.In

	DB  *gorm.DB `name:"psql"`
	Cfg *internal.Config
	L   *zap.Logger

	Gzip            converter.PayloadCodec `name:"gzip"`
	LargePayload    converter.PayloadCodec `name:"largepayload"`
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
		params.LargePayload,
		params.Gzip,
	))
}
