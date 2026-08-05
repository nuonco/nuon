package preflight

import (
	"context"
	"fmt"
	"os"
	"strings"

	pkgkafka "github.com/nuonco/nuon/pkg/kafka"
	internal "github.com/nuonco/nuon/services/ctl-api/internal"
	ctlkafka "github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
)

const kafkaSSLProtocol = "SSL"

var kafkaCheck = Check{
	Name:        "kafka",
	Description: "kafka broker connectivity",

	Skip: func(cfg *internal.Config) (string, bool) {
		if !cfg.KafkaEnabled {
			return "kafka_enabled=false", true
		}

		return "", false
	},

	Fields: func(cfg *internal.Config) []Field {
		usesTLS := strings.EqualFold(cfg.KafkaSecurityProtocol, kafkaSSLProtocol)

		return []Field{
			{Name: "kafka_brokers", Value: cfg.KafkaBrokers, Required: true},
			{Name: "kafka_security_protocol", Value: cfg.KafkaSecurityProtocol, Required: true},
			{Name: "kafka_client_id", Value: ctlkafka.ClientID(cfg)},
			{Name: "kafka_tls_ca_path", Value: cfg.KafkaTLSCAPath, Required: usesTLS},
			{Name: "kafka_tls_cert_path", Value: cfg.KafkaTLSCertPath, Required: usesTLS},
			{Name: "kafka_tls_key_path", Value: cfg.KafkaTLSKeyPath, Required: usesTLS},
			{Name: "kafka_consumer_group_prefix", Value: cfg.KafkaConsumerGroupPrefix},
		}
	},

	Probe: func(ctx context.Context, cfg *internal.Config) (string, error) {
		if strings.EqualFold(cfg.KafkaSecurityProtocol, kafkaSSLProtocol) {
			// Checked before dialling: a missing mount surfaces as an opaque
			// handshake error otherwise.
			for _, path := range []string{cfg.KafkaTLSCAPath, cfg.KafkaTLSCertPath, cfg.KafkaTLSKeyPath} {
				if _, err := os.Stat(path); err != nil {
					return "", fmt.Errorf("TLS material unreadable: %w", err)
				}
			}
		}

		producer, err := pkgkafka.NewProducer(ctlkafka.ClientConfig(cfg), nopLogger(), nopMetrics())
		if err != nil {
			return "", fmt.Errorf("unable to build producer: %w", err)
		}
		defer producer.Close()

		if err := producer.Ping(ctx); err != nil {
			return "", fmt.Errorf("ping failed: %w", err)
		}

		return fmt.Sprintf("brokers reachable %s", summary(
			"brokers", cfg.KafkaBrokers,
			"protocol", cfg.KafkaSecurityProtocol,
			"client_id", ctlkafka.ClientID(cfg))), nil
	},
}
