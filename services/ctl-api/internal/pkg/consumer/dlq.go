package consumer

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

// DLQConsumer drains the dead-letter topic into the same app.DLQRecord table the
// other consumers write to directly when a produce to this topic fails.
//
// It lives in the runtime package rather than with a domain because the topic
// isn't any domain's data — every consumer produces to it, and what it carries
// is a decode failure from some other topic.
//
// Its own decode failures go straight to that direct write rather than back
// through the topic — a dead-letter-about-a-dead-letter stops here, one level
// deep.
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
