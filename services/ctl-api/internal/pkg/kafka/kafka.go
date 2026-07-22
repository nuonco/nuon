package kafka

import (
	"context"
	"strings"

	"go.uber.org/fx"
	"go.uber.org/zap"

	pkgkafka "github.com/nuonco/nuon/pkg/kafka"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// Producer is re-exported so ctl-api call sites depend on this glue package
// rather than the generic transport directly.
type Producer = pkgkafka.Producer

// Topics this service produces to / consumes from. Names mirror the destination
// ClickHouse tables.
const (
	TopicRunnerHeartBeats = "runner_heart_beats"
	TopicOtelLogRecords   = "otel_log_records"
)

// Envelope message types.
const (
	TypeRunnerHeartBeat = "runner_heart_beat"
	TypeOtelLogRecord   = "otel_log_record"
)

// ClientConfig maps ctl-api config into the generic Kafka client config. Shared
// by the producer here and the domain consumers.
func ClientConfig(cfg *internal.Config) pkgkafka.Config {
	return pkgkafka.Config{
		Brokers:          splitBrokers(cfg.KafkaBrokers),
		ClientID:         cfg.KafkaClientID,
		SecurityProtocol: cfg.KafkaSecurityProtocol,
		SASLMechanism:    cfg.KafkaSASLMechanism,
		SASLUsername:     cfg.KafkaSASLUsername,
		SASLPassword:     cfg.KafkaSASLPassword,
		TLSEnabled:       cfg.KafkaTLSEnabled,
	}
}

type Params struct {
	fx.In

	Cfg *internal.Config
	L   *zap.Logger
	MW  metrics.Writer
	LC  fx.Lifecycle
}

// New provides the shared Kafka producer. When KAFKA_ENABLED is false it returns
// a no-op producer so callers fall back to their legacy inline path and
// downstream writes never depend on Kafka being present.
func New(params Params) (*pkgkafka.Producer, error) {
	l := params.L.Named("kafka")

	if !params.Cfg.KafkaEnabled {
		l.Info("kafka disabled; producer is a no-op")
		return pkgkafka.DisabledProducer(params.L, params.MW), nil
	}

	p, err := pkgkafka.NewProducer(ClientConfig(params.Cfg), params.L, params.MW)
	if err != nil {
		return nil, err
	}

	params.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := p.Ping(ctx); err != nil {
				l.Warn("kafka ping failed on startup; will retry on produce", zap.Error(err))
			} else {
				l.Info("kafka producer connected")
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := p.Flush(ctx); err != nil {
				l.Warn("kafka flush on shutdown failed", zap.Error(err))
			}
			p.Close()
			return nil
		},
	})

	return p, nil
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
