package consumer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	pkgconsumer "github.com/nuonco/nuon/services/ctl-api/internal/pkg/consumer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

// OtelTracesConsumer reads OTLP spans off Kafka and batch-writes them to
// ClickHouse. When it doesn't run — not selected, or Kafka disabled — New returns
// nil and producers keep writing to ClickHouse inline.
//
// Decodes into app.OtelTraceIngestion, the same struct the inline write path
// uses, so the row this inserts is the row that path would have inserted.
// Records arrive already carrying ID, CreatedByID, CreatedAt and UpdatedAt:
// OtelTraceIngestion's BeforeCreate hook resolves the first two from the GORM
// statement context and GORM autofills the timestamps at insert time, neither of
// which means anything here. See the producers in runners/service and
// controlplanejob.
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
