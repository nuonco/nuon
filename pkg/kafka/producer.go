package kafka

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
)

// Producer wraps a franz-go client. A disabled producer is a no-op, letting
// callers keep a legacy fallback path so downstream writes never depend on
// Kafka being present.
type Producer struct {
	l       *zap.Logger
	mw      metrics.Writer
	client  *kgo.Client
	source  string
	enabled bool
}

func NewProducer(cfg Config, l *zap.Logger, mw metrics.Writer) (*Producer, error) {
	opts, err := cfg.baseOpts()
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

	return &Producer{
		l:       l.Named("kafka-producer"),
		mw:      mw,
		client:  client,
		source:  cfg.ClientID,
		enabled: true,
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

// Produce sends a single pre-marshaled message. Fire-and-forget: the async
// callback records metrics and logs on failure, so callers never block on the
// broker.
func (p *Producer) Produce(ctx context.Context, topic, key string, value []byte) {
	if !p.enabled {
		p.mw.Incr("kafka.produce.disabled", nil)
		return
	}

	rec := &kgo.Record{Topic: topic, Key: []byte(key), Value: value}
	p.client.Produce(ctx, rec, func(_ *kgo.Record, err error) {
		if err != nil {
			p.l.Error("kafka produce failed", zap.String("topic", topic), zap.Error(err))
			p.mw.Incr("kafka.produce.error", []string{"topic:" + topic})
			return
		}
		p.mw.Incr("kafka.produce.count", []string{"topic:" + topic})
	})
}

// ProduceEnvelope wraps payload in the versioned envelope and produces it.
func (p *Producer) ProduceEnvelope(ctx context.Context, topic, key, typ string, payload any) error {
	value, err := Wrap(p.source, typ, payload)
	if err != nil {
		return fmt.Errorf("wrap %s: %w", typ, err)
	}
	p.Produce(ctx, topic, key, value)
	return nil
}
