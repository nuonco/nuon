package kafka

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
)

type Producer struct {
	l              *zap.Logger
	mw             metrics.Writer
	client         *kgo.Client
	source         string
	enabled        bool
	produceTimeout time.Duration
}

const defaultProduceTimeout = 5 * time.Second

type Message struct {
	Key     string
	Payload any
}

func NewProducer(cfg Config, l *zap.Logger, mw metrics.Writer) (*Producer, error) {
	opts, err := cfg.baseOpts(l)
	if err != nil {
		return nil, err
	}
	tracer := kotel.NewTracer(
		kotel.ClientID(cfg.ClientID),
		kotel.TracerPropagator(propagation.TraceContext{}),
	)
	opts = append(opts,
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerBatchCompression(kgo.Lz4Compression()),
		kgo.ProducerBatchMaxBytes(maxMessageBytes),
		kgo.WithHooks(kotel.NewKotel(kotel.WithTracer(tracer)).Hooks()...),
	)

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("new client: %w", err)
	}

	produceTimeout := cfg.ProduceTimeout
	if produceTimeout <= 0 {
		produceTimeout = defaultProduceTimeout
	}

	return &Producer{
		l:              l.Named("kafka-producer"),
		mw:             mw,
		client:         client,
		source:         cfg.ClientID,
		enabled:        true,
		produceTimeout: produceTimeout,
	}, nil
}

func DisabledProducer(l *zap.Logger, mw metrics.Writer) *Producer {
	return &Producer{l: l.Named("kafka-producer"), mw: mw, enabled: false}
}

func (p *Producer) Enabled() bool { return p.enabled }

func (p *Producer) Ping(ctx context.Context) error {
	if !p.enabled {
		return nil
	}
	return p.client.Ping(ctx)
}

func (p *Producer) Flush(ctx context.Context) error {
	if !p.enabled {
		return nil
	}
	return p.client.Flush(ctx)
}

func (p *Producer) Close() {
	if p.enabled {
		p.client.Close()
	}
}

func (p *Producer) writeMetrics(start time.Time, topic, status, reason string) {
	tags := []string{"topic:" + topic, "status:" + status}
	if reason != "" {
		tags = append(tags, "reason:"+reason)
	}
	p.mw.Timing("kafka.producer.latency", time.Since(start), tags)
	p.mw.Count("kafka.producer.message_count", 1, tags)
}

func (p *Producer) Produce(ctx context.Context, topic, key string, value []byte) {
	start := time.Now()
	if !p.enabled {
		p.writeMetrics(start, topic, "disabled", "")
		return
	}

	rec := &kgo.Record{Topic: topic, Key: []byte(key), Value: value}
	p.client.Produce(ctx, rec, func(_ *kgo.Record, err error) {
		if err != nil {
			p.l.Error("kafka produce failed", zap.String("topic", topic), zap.Error(err))
			p.writeMetrics(start, topic, "err", "broker_error")
			return
		}
		p.writeMetrics(start, topic, "ok", "")
	})
}

func (p *Producer) ProduceEnvelope(ctx context.Context, topic, key, typ string, payload any) error {
	value, err := Wrap(p.source, typ, payload)
	if err != nil {
		return fmt.Errorf("wrap %s: %w", typ, err)
	}
	p.Produce(ctx, topic, key, value)
	return nil
}

// why: ProduceEnvelopesSync wraps each message and produces the batch, blocking until
// every record is acked. Acks are already RequiredAcks(AllISRAcks) with
// idempotence on, so an ack here means the record is on every in-sync replica.
//
// Returns the indices of messages that were NOT acked, so a caller with a
// fallback can write exactly those rather than re-writing the whole batch and
// duplicating the ones that succeeded. An empty return means everything is
// durable in Kafka.
//
// Produces the whole batch in one call rather than looping: ProduceSync enqueues
// all records, cancels lingering, and waits once, so this costs one round trip's
// latency instead of N serialized ones.
//
// On timeout the outcome for an in-flight record is genuinely ambiguous.
// franz-go will not fail a buffered record that has already been sent while
// producing idempotently, because it cannot know whether the broker applied it.
// Such a record is reported as failed here — so the caller writes it via the
// fallback — and may still land in Kafka afterwards. That is an at-least-once
// choice: a duplicate row is recoverable, a silently dropped log line is not.
func (p *Producer) ProduceEnvelopesSync(ctx context.Context, topic, typ string, msgs []Message) []int {
	if len(msgs) == 0 {
		return nil
	}
	batchStart := time.Now()
	if !p.enabled {
		for range msgs {
			p.writeMetrics(batchStart, topic, "disabled", "")
		}
		return allIndices(len(msgs))
	}

	recs := make([]*kgo.Record, 0, len(msgs))
	// why: ProduceSync appends results in promise-completion order, not input order, so
	// results cannot be matched to messages positionally. Map by record pointer
	// instead — ProduceResult.Record is documented as always non-nil. Getting this
	// wrong would attribute a failure to the wrong message and have the caller
	// duplicate an acked record while dropping a failed one.
	msgIdx := make(map[*kgo.Record]int, len(msgs))
	var failed []int

	for i, msg := range msgs {
		value, err := Wrap(p.source, typ, msg.Payload)
		if err != nil {
			p.l.Error("unable to wrap record for produce",
				zap.String("topic", topic),
				zap.String("type", typ),
				zap.Error(err),
			)
			p.writeMetrics(batchStart, topic, "err", "wrap_error")
			failed = append(failed, i)
			continue
		}
		rec := &kgo.Record{Topic: topic, Key: []byte(msg.Key), Value: value}
		recs = append(recs, rec)
		msgIdx[rec] = i
	}

	if len(recs) == 0 {
		return failed
	}

	ctx, cancel := context.WithTimeout(ctx, p.produceTimeout)
	defer cancel()

	results := p.client.ProduceSync(ctx, recs...)

	acked := make(map[*kgo.Record]bool, len(recs))
	for _, res := range results {
		i, ok := msgIdx[res.Record]
		if !ok {
			// why: Cannot happen with records we just built, but silently discarding an
			// unmatched result would mean silently dropping a log record.
			p.l.Error("kafka sync produce returned an unrecognized record",
				zap.String("topic", topic),
				zap.Error(res.Err),
			)
			p.writeMetrics(batchStart, topic, "err", "unmatched_result")
			continue
		}

		if res.Err != nil {
			p.l.Error("kafka sync produce failed",
				zap.String("topic", topic),
				zap.Error(res.Err),
			)
			p.writeMetrics(batchStart, topic, "err", "broker_error")
			failed = append(failed, i)
			continue
		}
		acked[res.Record] = true
		p.writeMetrics(batchStart, topic, "ok", "")
	}

	for rec, i := range msgIdx {
		if acked[rec] {
			continue
		}
		if !slices.Contains(failed, i) {
			p.l.Error("kafka sync produce returned no result for a record",
				zap.String("topic", topic),
			)
			p.writeMetrics(batchStart, topic, "err", "missing_result")
			failed = append(failed, i)
		}
	}

	return failed
}

func allIndices(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}
