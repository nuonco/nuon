package kafka

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
)

// Producer wraps a franz-go client. A disabled producer is a no-op, letting
// callers keep a legacy fallback path so downstream writes never depend on
// Kafka being present.
//
// Two produce modes, chosen by what the caller's current write does:
//
//   - Produce / ProduceEnvelope are fire-and-forget. Correct where the existing
//     path is already lossy — heartbeats buffer in memory and flush on a 5s
//     ticker, so a crash already drops a few seconds of them and nothing cares.
//   - ProduceEnvelopesSync waits for the acks. Correct where the existing path is
//     synchronously durable, as the OTLP log write is: it blocks on a ClickHouse
//     insert before returning 201, so producing fire-and-forget instead would
//     hand the caller a success for records that live only in a process buffer.
type Producer struct {
	l              *zap.Logger
	mw             metrics.Writer
	client         *kgo.Client
	source         string
	enabled        bool
	produceTimeout time.Duration
}

// defaultProduceTimeout bounds a sync produce when config leaves it unset.
const defaultProduceTimeout = 5 * time.Second

// Message is one record for a batch produce.
type Message struct {
	Key     string
	Payload any
}

func NewProducer(cfg Config, l *zap.Logger, mw metrics.Writer) (*Producer, error) {
	opts, err := cfg.baseOpts(l)
	if err != nil {
		return nil, err
	}
	opts = append(opts,
		// idempotent producer is on by default with acks=all
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerBatchCompression(kgo.Lz4Compression()),
		kgo.ProducerBatchMaxBytes(maxMessageBytes),
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

// writeMetrics records both the timing and the volume for one produce outcome.
// Timing gives status/reason percentiles; message_count exists as its own Count
// rather than relying on Timing's derived .count so volume tracking doesn't
// silently change meaning if the per-message sampling choice below ever changes.
func (p *Producer) writeMetrics(start time.Time, topic, status, reason string) {
	tags := []string{"topic:" + topic, "status:" + status}
	if reason != "" {
		tags = append(tags, "reason:"+reason)
	}
	p.mw.Timing("kafka.producer.latency", time.Since(start), tags)
	p.mw.Count("kafka.producer.message_count", 1, tags)
}

// Produce sends a single pre-marshaled message. Fire-and-forget: the async
// callback records metrics and logs on failure, so callers never block on the
// broker.
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

// ProduceEnvelope wraps payload in the versioned envelope and produces it.
// Fire-and-forget; see ProduceEnvelopesSync when the caller needs the ack.
func (p *Producer) ProduceEnvelope(ctx context.Context, topic, key, typ string, payload any) error {
	value, err := Wrap(p.source, typ, payload)
	if err != nil {
		return fmt.Errorf("wrap %s: %w", typ, err)
	}
	p.Produce(ctx, topic, key, value)
	return nil
}

// ProduceEnvelopesSync wraps each message and produces the batch, blocking until
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
	// ProduceSync appends results in promise-completion order, not input order, so
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
			// Cannot happen with records we just built, but silently discarding an
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

	// ProduceSync waits on a promise per record, so every record should be
	// accounted for. Treat anything unaccounted as failed rather than assuming it
	// landed: the whole point of the sync path is that an unacked record goes down
	// the caller's fallback instead of being lost.
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
