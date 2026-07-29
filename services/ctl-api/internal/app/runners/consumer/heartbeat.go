package consumer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

// HeartbeatConsumer reads runner heartbeats off Kafka and batch-writes them to
// ClickHouse. When it doesn't run — not selected, or Kafka disabled — New returns
// nil and the inline heartbeater path remains in effect.
type HeartbeatConsumer struct {
	*sink
}

func NewHeartbeatConsumer(params Params) (*HeartbeatConsumer, error) {
	s := newSink(params, NameHeartbeats, kafka.TopicRunnerHeartBeats)
	if s == nil {
		return nil, nil
	}

	c := &HeartbeatConsumer{sink: s}
	if err := s.start(params, c.handle); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *HeartbeatConsumer) handle(ctx context.Context, partition int32, recs []*kgo.Record) error {
	hbs := decode[app.RunnerHeartBeat](c.sink, recs, kafka.TypeRunnerHeartBeat)
	return insert(ctx, c.sink, partition, recs, hbs)
}
