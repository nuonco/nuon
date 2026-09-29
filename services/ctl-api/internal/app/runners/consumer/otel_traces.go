package consumer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	pkgconsumer "github.com/nuonco/nuon/services/ctl-api/internal/pkg/consumer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

type OtelTracesConsumer struct {
	*pkgconsumer.Sink
}

func NewOtelTracesConsumer(params pkgconsumer.Params) (*OtelTracesConsumer, error) {
	s := pkgconsumer.NewSink(params, pkgconsumer.NameOtelTraces, kafka.TopicOtelTraces)
	if s == nil {
		return nil, nil
	}

	c := &OtelTracesConsumer{Sink: s}
	if err := s.Start(params, c.handle); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *OtelTracesConsumer) handle(ctx context.Context, partition int32, recs []*kgo.Record) error {
	spans := pkgconsumer.Decode[app.OtelTraceIngestion](ctx, c.Sink, recs, kafka.TypeOtelTrace)
	return pkgconsumer.Insert(ctx, c.Sink, partition, recs, spans)
}
