package consumer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

type DLQConsumer struct {
	*Sink
}

func NewDLQConsumer(params Params) (*DLQConsumer, error) {
	s := NewSink(params, NameDLQ, kafka.TopicDLQ)
	if s == nil {
		return nil, nil
	}
	s.deadLetterFallbackOnly = true

	c := &DLQConsumer{Sink: s}
	if err := s.Start(params, c.handle); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *DLQConsumer) handle(ctx context.Context, partition int32, recs []*kgo.Record) error {
	letters := Decode[app.DLQRecord](ctx, c.Sink, recs, kafka.TypeDLQ)
	return Insert(ctx, c.Sink, partition, recs, letters)
}
