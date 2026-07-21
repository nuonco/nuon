package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg *internal.Config
	L   *zap.Logger
	MW  metrics.Writer
	LC  fx.Lifecycle
}

type Producer struct {
	l       *zap.Logger
	mw      metrics.Writer
	client  *kgo.Client
	enabled bool
}

func New(params Params) (*Producer, error) {
	l := params.L.Named("kafka")

	if !params.Cfg.KafkaEnabled {
		l.Info("kafka disabled; producer is a no-op")
		return &Producer{l: l, mw: params.MW, enabled: false}, nil
	}

	cc := clientConfig{
		Brokers:          splitBrokers(params.Cfg.KafkaBrokers),
		ClientID:         params.Cfg.KafkaClientID,
		SecurityProtocol: params.Cfg.KafkaSecurityProtocol,
		SASLMechanism:    params.Cfg.KafkaSASLMechanism,
		SASLUsername:     params.Cfg.KafkaSASLUsername,
		SASLPassword:     params.Cfg.KafkaSASLPassword,
		TLSEnabled:       params.Cfg.KafkaTLSEnabled,
	}

	opts, err := cc.buildOpts()
	if err != nil {
		return nil, fmt.Errorf("kafka: build client opts: %w", err)
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("kafka: new client: %w", err)
	}

	p := &Producer{l: l, mw: params.MW, client: client, enabled: true}

	params.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Boot-safe: warn-only ping so a down broker never blocks startup.
			// franz-go dials lazily and retries on produce.
			if err := client.Ping(ctx); err != nil {
				l.Warn("kafka ping failed on startup; will retry on produce", zap.Error(err))
			} else {
				l.Info("kafka producer connected", zap.Strings("brokers", cc.Brokers))
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := client.Flush(ctx); err != nil {
				l.Warn("kafka flush on shutdown failed", zap.Error(err))
			}
			client.Close()
			return nil
		},
	})

	return p, nil
}

func (p *Producer) Enabled() bool { return p.enabled }

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

func (p *Producer) ProduceEnvelope(ctx context.Context, topic, key, typ string, payload any) error {
	value, err := Wrap(typ, payload)
	if err != nil {
		return fmt.Errorf("kafka: wrap %s: %w", typ, err)
	}
	p.Produce(ctx, topic, key, value)
	return nil
}

func splitBrokers(s string) []string {
	var out []string
	for _, b := range strings.Split(s, ",") {
		if b = strings.TrimSpace(b); b != "" {
			out = append(out, b)
		}
	}
	return out
}
