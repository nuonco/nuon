package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	pkgkafka "github.com/nuonco/nuon/pkg/kafka"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

// buildDeadLetter is pure — no I/O — so Decode can collect every failure in a
// fetch before recordDeadLetters does anything over the network.
func (s *Sink) buildDeadLetter(rec *kgo.Record, reason string, cause error, env pkgkafka.Envelope) app.DLQRecord {
	return app.DLQRecord{
		Topic:         s.topic,
		Partition:     rec.Partition,
		Offset:        rec.Offset,
		ConsumerGroup: s.group,
		ConsumerName:  string(s.name),
		Reason:        reason,
		Error:         cause.Error(),
		EnvelopeType:  env.Type,
		ProducedAt:    env.ProducedAt,
		FailedAt:      time.Now(),
		RawValue:      string(rec.Value),
	}
}

// recordDeadLetters durably records every record Decode() couldn't process
// from one fetch, in a single batched call — one produce round trip and, on
// the fallback path, one ClickHouse insert, regardless of how many records
// failed. Produces to the dead-letter topic synchronously, same durability
// reasoning as the OTel logs path: this is the last chance to keep the
// evidence, so fire-and-forget isn't good enough. Falls back to a batched
// direct ClickHouse write, bounded by the same write timeout as Insert(), for
// whichever records the produce itself didn't ack.
//
// deadLetterFallbackOnly skips the produce step entirely: the DLQ consumer's
// own decode failures go straight to the direct write, so a
// dead-letter-about-a-dead-letter can't recurse through the topic.
func (s *Sink) recordDeadLetters(ctx context.Context, dead []app.DLQRecord) {
	if !s.deadLetterFallbackOnly {
		msgs := make([]kafka.Message, len(dead))
		for i, dl := range dead {
			msgs[i] = kafka.Message{
				Key:     fmt.Sprintf("%s:%d:%d", dl.Topic, dl.Partition, dl.Offset),
				Payload: dl,
			}
		}

		failed := s.producer.ProduceEnvelopesSync(ctx, kafka.TopicDLQ, kafka.TypeDLQ, msgs)
		if len(failed) == 0 {
			return
		}
		s.mw.Count("kafka.consumer.dead_letter", int64(len(failed)), s.baseTags("reason:produce_failed"))

		fallback := make([]app.DLQRecord, 0, len(failed))
		for _, i := range failed {
			fallback = append(fallback, dead[i])
		}
		dead = fallback
	}

	writeCtx, cancel := context.WithTimeout(ctx, s.writeTimeout)
	defer cancel()

	if err := s.chDB.WithContext(writeCtx).CreateInBatches(&dead, len(dead)).Error; err != nil {
		s.l.Error("failed to record dead letters via topic or direct write",
			zap.Int("count", len(dead)),
			zap.Error(err),
		)
		s.mw.Count("kafka.consumer.dead_letter", int64(len(dead)), s.baseTags("reason:lost"))
	}
}
