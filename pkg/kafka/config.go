package kafka

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

// maxMessageBytes — Azure Event Hubs hard-caps a single event at 1 MiB and it
// cannot be raised, so we never build a batch larger than that.
const maxMessageBytes = 1024 * 1024

const (
	securitySSL     = "SSL"
	securitySASLSSL = "SASL_SSL"
)

const (
	saslPlain     = "PLAIN"
	saslScram256  = "SCRAM-SHA-256"
	saslScram512  = "SCRAM-SHA-512"
	saslOAuth     = "OAUTHBEARER"
	saslAWSMSKIAM = "AWS_MSK_IAM"
)

// Config is the provider-agnostic Kafka client configuration shared by the
// producer and consumer.
type Config struct {
	Brokers          []string
	ClientID         string
	SecurityProtocol string
	SASLMechanism    string
	SASLUsername     string
	SASLPassword     string
	TLSEnabled       bool
}

// baseOpts builds the transport options shared by producers and consumers.
// Security is pluggable per cloud: local is PLAINTEXT; MSK IAM and GCP/Azure
// OAUTHBEARER are wired per-cloud later.
func (c Config) baseOpts() ([]kgo.Opt, error) {
	if len(c.Brokers) == 0 {
		return nil, fmt.Errorf("no brokers configured")
	}

	opts := []kgo.Opt{kgo.SeedBrokers(c.Brokers...)}
	if c.ClientID != "" {
		opts = append(opts, kgo.ClientID(c.ClientID))
	}

	if c.TLSEnabled || c.SecurityProtocol == securitySSL || c.SecurityProtocol == securitySASLSSL {
		opts = append(opts, kgo.DialTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}))
	}

	switch strings.ToUpper(strings.TrimSpace(c.SASLMechanism)) {
	case "":
	case saslPlain:
		opts = append(opts, kgo.SASL(plain.Auth{User: c.SASLUsername, Pass: c.SASLPassword}.AsMechanism()))
	case saslScram256:
		opts = append(opts, kgo.SASL(scram.Auth{User: c.SASLUsername, Pass: c.SASLPassword}.AsSha256Mechanism()))
	case saslScram512:
		opts = append(opts, kgo.SASL(scram.Auth{User: c.SASLUsername, Pass: c.SASLPassword}.AsSha512Mechanism()))
	case saslAWSMSKIAM:
		return nil, fmt.Errorf("SASL mechanism %q not yet implemented", c.SASLMechanism)
	case saslOAuth:
		return nil, fmt.Errorf("SASL mechanism %q not yet implemented", c.SASLMechanism)
	default:
		return nil, fmt.Errorf("unknown SASL mechanism %q", c.SASLMechanism)
	}

	return opts, nil
}
