package consumer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	pkgconsumer "github.com/nuonco/nuon/services/ctl-api/internal/pkg/consumer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

type HeartbeatConsumer struct {
	*pkgconsumer.Sink
}

func NewHeartbeatConsumer(params pkgconsumer.Params) (*HeartbeatConsumer, error) {
	s := pkgconsumer.NewSink(params, pkgconsumer.NameHeartbeats, kafka.TopicRunnerHeartBeats)
	if s == nil {
		return nil, nil
	}

	c := &HeartbeatConsumer{Sink: s}
	if err := s.Start(params, c.handle); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *HeartbeatConsumer) handle(ctx context.Context, partition int32, recs []*kgo.Record) error {
	hbs := pkgconsumer.Decode[app.RunnerHeartBeat](ctx, c.Sink, recs, kafka.TypeRunnerHeartBeat)
	return pkgconsumer.Insert(ctx, c.Sink, partition, recs, hbs)
}
