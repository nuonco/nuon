package consumer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

// DLQConsumer drains the dlq topic into the same app.DLQRecord table the
// other consumers write to directly when a produce to this topic fails. Its
// own decode failures go straight to that direct write (fallbackOnly=true on
// its sink) rather than back through the topic — a dead-letter-about-a-dead-
// letter stops here, one level deep.
type DLQConsumer struct {
	*sink
}

func NewDLQConsumer(params Params) (*DLQConsumer, error) {
	s := newSink(params, NameDLQ, kafka.TopicDLQ, true)
	if s == nil {
		return nil, nil
	}

	c := &DLQConsumer{sink: s}
	if err := s.start(params, c.handle); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *DLQConsumer) handle(ctx context.Context, partition int32, recs []*kgo.Record) error {
	letters := decode[app.DLQRecord](ctx, c.sink, recs, kafka.TypeDLQ)
	return insert(ctx, c.sink, partition, recs, letters)
}
