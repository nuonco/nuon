package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

// maxMessageBytes mirrors the topic-level max.message.bytes that
// images/kafka/create-topics.sh sets in every environment, so we never build a
// batch the broker will reject.
const maxMessageBytes = 1024 * 1024

const (
	securityPlaintext = "PLAINTEXT"
	securitySSL       = "SSL"
)

// Config is the Kafka client configuration shared by the producer and consumer.
type Config struct {
	Brokers          []string
	ClientID         string
	SecurityProtocol string

	// TLSCAPath is the bundle used to verify brokers; empty uses system roots.
	TLSCAPath string
	// TLSCertPath and TLSKeyPath enable mTLS and must be set together.
	TLSCertPath string
	TLSKeyPath  string

	// ProduceTimeout bounds a synchronous produce. franz-go retries a buffered
	// record effectively forever by default, so without a bound a broker outage
	// would block a caller waiting on the ack indefinitely. Only applies to the
	// sync path; async producers never wait.
	ProduceTimeout time.Duration
}

func (c Config) protocol() string {
	return strings.ToUpper(strings.TrimSpace(c.SecurityProtocol))
}

// baseOpts builds the transport options shared by producers and consumers.
// Only two protocols exist, matching how Kafka is deployed: PLAINTEXT locally,
// and SSL with a client certificate against Strimzi, whose KafkaUsers
// authenticate by mTLS. There is no SASL path.
func (c Config) baseOpts(l *zap.Logger) ([]kgo.Opt, error) {
	if len(c.Brokers) == 0 {
		return nil, fmt.Errorf("no brokers configured")
	}

	opts := []kgo.Opt{kgo.SeedBrokers(c.Brokers...)}
	if c.ClientID != "" {
		opts = append(opts, kgo.ClientID(c.ClientID))
	}

	switch c.protocol() {
	case "", securityPlaintext:
	case securitySSL:
		reloader, err := newTLSReloader(c, l)
		if err != nil {
			return nil, err
		}
		opts = append(opts, kgo.Dialer(reloader.dial))
	default:
		// Reject rather than fall through to plaintext: a typo'd protocol would
		// otherwise silently drop TLS against a broker that requires it.
		return nil, fmt.Errorf("unsupported security protocol %q", c.SecurityProtocol)
	}

	return opts, nil
}
