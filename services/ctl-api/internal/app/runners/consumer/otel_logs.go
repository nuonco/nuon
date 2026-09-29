package consumer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	pkgconsumer "github.com/nuonco/nuon/services/ctl-api/internal/pkg/consumer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

type OtelLogsConsumer struct {
	*pkgconsumer.Sink
}

func NewOtelLogsConsumer(params pkgconsumer.Params) (*OtelLogsConsumer, error) {
	s := pkgconsumer.NewSink(params, pkgconsumer.NameOtelLogs, kafka.TopicOtelLogRecords)
	if s == nil {
		return nil, nil
	}

	c := &OtelLogsConsumer{Sink: s}
	if err := s.Start(params, c.handle); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *OtelLogsConsumer) handle(ctx context.Context, partition int32, recs []*kgo.Record) error {
	logs := pkgconsumer.Decode[app.OtelLogRecord](ctx, c.Sink, recs, kafka.TypeOtelLogRecord)
	return pkgconsumer.Insert(ctx, c.Sink, partition, recs, logs)
}
