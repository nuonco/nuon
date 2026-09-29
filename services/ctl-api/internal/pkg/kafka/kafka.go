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

type Producer = pkgkafka.Producer
type Message = pkgkafka.Message

const (
	TopicRunnerHeartBeats = "runner_heart_beats"
	TopicOtelLogRecords   = "otel_log_records"
	TopicOtelTraces       = "otel_traces"
	TopicDLQ              = "dlq"
)

const (
	TypeRunnerHeartBeat = "runner_heart_beat"
	TypeOtelLogRecord   = "otel_log_record"
	TypeOtelTrace       = "otel_trace"
	TypeDLQ             = "dlq_record"
)

func ClientID(cfg *internal.Config) string {
	if cfg.KafkaClientID != "" {
		return cfg.KafkaClientID
	}
	if cfg.ServiceDeployment == "" {
		return cfg.ServiceName + "/" + cfg.ServiceType
	}

	return cfg.ServiceName + "/" + cfg.ServiceType + "-" + cfg.ServiceDeployment
}

func ClientConfig(cfg *internal.Config) pkgkafka.Config {
	return pkgkafka.Config{
		Brokers:          splitBrokers(cfg.KafkaBrokers),
		ClientID:         ClientID(cfg),
		SecurityProtocol: cfg.KafkaSecurityProtocol,
		TLSCAPath:        cfg.KafkaTLSCAPath,
		TLSCertPath:      cfg.KafkaTLSCertPath,
		TLSKeyPath:       cfg.KafkaTLSKeyPath,
		ProduceTimeout:   cfg.KafkaProduceTimeout,
	}
}

// why: ConsumerGroup names the group for one consumer. Per-consumer rather than one
// shared group because a group with heterogeneous topic subscriptions rebalances
// every member whenever any member restarts — so a deploy of one consumer would
// stall the others for no reason.
//
// Changing a group's name makes it start from the earliest retained offset
// (pkgkafka sets ConsumeResetOffset to AtStart), which for a topic with history
// means replaying the whole retention window. Rename only while a topic is cold,
// or seed the new group from the old one's committed offsets first.
func ConsumerGroup(cfg *internal.Config, name string) string {
	return cfg.KafkaConsumerGroupPrefix + "-" + name
}

func ConsumerConfig(cfg *internal.Config, name, topic string) pkgkafka.ConsumerConfig {
	return pkgkafka.ConsumerConfig{
		Group:                  ConsumerGroup(cfg, name),
		Topics:                 []string{topic},
		FetchMaxWait:           cfg.KafkaConsumerFetchMaxWait,
		FetchMinBytes:          cfg.KafkaConsumerFetchMinBytes,
		FetchMaxBytes:          cfg.KafkaConsumerFetchMaxBytes,
		FetchMaxPartitionBytes: cfg.KafkaConsumerFetchMaxPartitionBytes,
		MaxConcurrentFetches:   cfg.KafkaConsumerMaxConcurrentFetches,
	}
}

type Params struct {
	fx.In

	Cfg *internal.Config
	L   *zap.Logger
	MW  metrics.Writer
	LC  fx.Lifecycle
}

// why: New provides the shared Kafka producer. When KAFKA_ENABLED is false it returns
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
