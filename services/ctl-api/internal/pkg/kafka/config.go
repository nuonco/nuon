package kafka

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
)

// max message size — Azure Event Hubs hard-caps a single event at 1 MiB and it
// cannot be raised, so we never build a batch larger than that.
const maxMessageBytes = 1024 * 1024

// security protocols
const (
	securitySSL     = "SSL"
	securitySASLSSL = "SASL_SSL"
)

// SASL mechanisms
const (
	saslPlain     = "PLAIN"
	saslScram256  = "SCRAM-SHA-256"
	saslScram512  = "SCRAM-SHA-512"
	saslOAuth     = "OAUTHBEARER"
	saslAWSMSKIAM = "AWS_MSK_IAM"
)

// clientConfig is the resolved, provider-agnostic Kafka client configuration.
type clientConfig struct {
	Brokers          []string
	ClientID         string
	SecurityProtocol string
	SASLMechanism    string
	SASLUsername     string
	SASLPassword     string
	TLSEnabled       bool
}

// buildOpts translates config into franz-go client options. Security is
// pluggable per cloud: local is PLAINTEXT; MSK IAM and GCP/Azure OAUTHBEARER are
// wired per-cloud later.
func (c clientConfig) buildOpts() ([]kgo.Opt, error) {
	if len(c.Brokers) == 0 {
		return nil, fmt.Errorf("no brokers configured")
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(c.Brokers...),
		kgo.ClientID(c.ClientID),
		// idempotent producer is on by default with acks=all
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerBatchCompression(kgo.Lz4Compression()),
		kgo.ProducerBatchMaxBytes(maxMessageBytes),
	}

	if c.TLSEnabled || c.SecurityProtocol == securitySSL || c.SecurityProtocol == securitySASLSSL {
		opts = append(opts, kgo.DialTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}))
	}

	switch strings.ToUpper(strings.TrimSpace(c.SASLMechanism)) {
	case "":
		// no SASL — local PLAINTEXT
	case saslPlain:
		opts = append(opts, kgo.SASL(plain.Auth{User: c.SASLUsername, Pass: c.SASLPassword}.AsMechanism()))
	case saslScram256:
		opts = append(opts, kgo.SASL(scram.Auth{User: c.SASLUsername, Pass: c.SASLPassword}.AsSha256Mechanism()))
	case saslScram512:
		opts = append(opts, kgo.SASL(scram.Auth{User: c.SASLUsername, Pass: c.SASLPassword}.AsSha512Mechanism()))
	case saslAWSMSKIAM:
		// TODO(kafka): wire AWS MSK IAM (franz-go pkg/sasl/aws) when we target MSK.
		return nil, fmt.Errorf("SASL mechanism %q not yet implemented", c.SASLMechanism)
	case saslOAuth:
		// TODO(kafka): wire OAUTHBEARER token source per cloud (GCP/Azure).
		return nil, fmt.Errorf("SASL mechanism %q not yet implemented", c.SASLMechanism)
	default:
		return nil, fmt.Errorf("unknown SASL mechanism %q", c.SASLMechanism)
	}

	return opts, nil
}
