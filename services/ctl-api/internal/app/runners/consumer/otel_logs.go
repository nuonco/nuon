package consumer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	pkgconsumer "github.com/nuonco/nuon/services/ctl-api/internal/pkg/consumer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

// OtelLogsConsumer reads OTLP log records off Kafka and batch-writes them to
// ClickHouse. When it doesn't run — not selected, or Kafka disabled — New returns
// nil and producers keep writing to ClickHouse inline.
//
// Records arrive already fully populated, including ID, OrgID and CreatedByID:
// app.OtelLogRecord's BeforeCreate hook resolves those from the request context,
// which does not exist here, and org_id leads the destination table's ORDER BY.
// See the producers in runners/service and controlplanejob.
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
