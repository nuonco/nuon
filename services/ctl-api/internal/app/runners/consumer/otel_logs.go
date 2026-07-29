package consumer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
	*sink
}

func NewOtelLogsConsumer(params Params) (*OtelLogsConsumer, error) {
	s := newSink(params, NameOtelLogs, kafka.TopicOtelLogRecords)
	if s == nil {
		return nil, nil
	}

	c := &OtelLogsConsumer{sink: s}
	if err := s.start(params, c.handle); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *OtelLogsConsumer) handle(ctx context.Context, partition int32, recs []*kgo.Record) error {
	logs := decode[app.OtelLogRecord](c.sink, recs, kafka.TypeOtelLogRecord)
	return insert(ctx, c.sink, partition, recs, logs)
}
