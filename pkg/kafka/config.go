package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/aws"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"go.uber.org/zap"
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

	// TLSCAPath is the bundle used to verify brokers; empty uses system roots.
	TLSCAPath string
	// TLSCertPath and TLSKeyPath enable mTLS and must be set together.
	TLSCertPath string
	TLSKeyPath  string
}

func (c Config) tlsRequired() bool {
	return c.TLSEnabled ||
		c.SecurityProtocol == securitySSL ||
		c.SecurityProtocol == securitySASLSSL ||
		strings.ToUpper(strings.TrimSpace(c.SASLMechanism)) == saslAWSMSKIAM
}

// baseOpts builds the transport options shared by producers and consumers.
// Security is pluggable per cloud: local is PLAINTEXT; MSK IAM and GCP/Azure
// OAUTHBEARER are wired per-cloud later.
func (c Config) baseOpts(l *zap.Logger) ([]kgo.Opt, error) {
	if len(c.Brokers) == 0 {
		return nil, fmt.Errorf("no brokers configured")
	}

	opts := []kgo.Opt{kgo.SeedBrokers(c.Brokers...)}
	if c.ClientID != "" {
		opts = append(opts, kgo.ClientID(c.ClientID))
	}

	mechanism := strings.ToUpper(strings.TrimSpace(c.SASLMechanism))

	if c.tlsRequired() {
		reloader, err := newTLSReloader(c, l)
		if err != nil {
			return nil, err
		}
		opts = append(opts, kgo.Dialer(reloader.dial))
	}

	switch mechanism {
	case "":
	case saslPlain:
		opts = append(opts, kgo.SASL(plain.Auth{User: c.SASLUsername, Pass: c.SASLPassword}.AsMechanism()))
	case saslScram256:
		opts = append(opts, kgo.SASL(scram.Auth{User: c.SASLUsername, Pass: c.SASLPassword}.AsSha256Mechanism()))
	case saslScram512:
		opts = append(opts, kgo.SASL(scram.Auth{User: c.SASLUsername, Pass: c.SASLPassword}.AsSha512Mechanism()))
	case saslAWSMSKIAM:
		opts = append(opts, kgo.SASL(aws.ManagedStreamingIAM(mskIAMAuth)))
	case saslOAuth:
		return nil, fmt.Errorf("SASL mechanism %q not yet implemented", c.SASLMechanism)
	default:
		return nil, fmt.Errorf("unknown SASL mechanism %q", c.SASLMechanism)
	}

	return opts, nil
}

// mskIAMAuth is invoked by franz-go on every authentication attempt, so
// credentials are loaded fresh each time rather than cached at startup —
// IRSA/STS-issued creds expire and must be refreshed.
// The provider chain is built once: cfg.Credentials is an SDK CredentialsCache,
// so retrieving per call refreshes expiring IRSA/STS credentials without
// re-reading the token file or re-calling STS on every authentication.
var loadAWSConfig = sync.OnceValues(func() (awssdk.Config, error) {
	return config.LoadDefaultConfig(context.Background())
})

// franz-go invokes this on every authentication.
func mskIAMAuth(ctx context.Context) (aws.Auth, error) {
	cfg, err := loadAWSConfig()
	if err != nil {
		return aws.Auth{}, fmt.Errorf("loading AWS config: %w", err)
	}

	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return aws.Auth{}, fmt.Errorf("retrieving AWS credentials: %w", err)
	}

	return aws.Auth{
		AccessKey:    creds.AccessKeyID,
		SecretKey:    creds.SecretAccessKey,
		SessionToken: creds.SessionToken,
	}, nil
}
