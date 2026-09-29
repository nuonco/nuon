package kafka

import (
	"fmt"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

// why: maxMessageBytes mirrors the topic-level max.message.bytes that
// images/kafka/create-topics.sh sets in every environment, so we never build a
// batch the broker will reject. Kept in sync by convention, not enforced —
// see infra/kafka/vars/defaults.yaml (mono) for the broker-side value.
const maxMessageBytes = 4 * 1024 * 1024

const (
	securityPlaintext = "PLAINTEXT"
	securitySSL       = "SSL"
)

type Config struct {
	Brokers          []string
	ClientID         string
	SecurityProtocol string

	TLSCAPath   string
	TLSCertPath string
	TLSKeyPath  string

	ProduceTimeout time.Duration
}

func (c Config) protocol() string {
	return strings.ToUpper(strings.TrimSpace(c.SecurityProtocol))
}

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
		// why: Reject rather than fall through to plaintext: a typo'd protocol would
		// otherwise silently drop TLS against a broker that requires it.
		return nil, fmt.Errorf("unsupported security protocol %q", c.SecurityProtocol)
	}

	return opts, nil
}
