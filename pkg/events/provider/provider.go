package provider

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/nuonco/nuon/pkg/events/envelope"
	"github.com/nuonco/nuon/pkg/events/signature"
	"github.com/nuonco/nuon/pkg/events/sns"
)

type AuthType string

const (
	AuthNone         AuthType = "none"
	AuthHMAC         AuthType = "hmac"
	AuthAPIKey       AuthType = "api_key"
	AuthBasic        AuthType = "basic"
	AuthBearerJWT    AuthType = "bearer_jwt"
	AuthSNSSignature AuthType = "sns_signature"
)

type EnvelopeType string

const (
	EnvelopeNone        EnvelopeType = "none"
	EnvelopePubSubPush  EnvelopeType = "pubsub_push"
	EnvelopeCloudEvents EnvelopeType = "cloudevents"
	EnvelopeSNS         EnvelopeType = "sns"
)

var ErrUnsupportedEnvelope = errors.New("trigger envelope is not implemented")

type AuthConfig struct {
	Header          string   `json:"header,omitempty"`
	Prefix          string   `json:"prefix,omitempty"`
	Encoding        string   `json:"encoding,omitempty"`
	Algorithm       string   `json:"algorithm,omitempty"`
	Username        string   `json:"username,omitempty"`
	Issuer          string   `json:"issuer,omitempty"`
	Audience        []string `json:"audience,omitempty"`
	TopicARN        string   `json:"topic_arn,omitempty"`
	ExpectedEmail   string   `json:"expected_email,omitempty"`
	ExpectedSubject string   `json:"expected_subject,omitempty"`
}

type CallerField string

const (
	CallerFieldAudience        CallerField = "audience"
	CallerFieldExpectedEmail   CallerField = "expected_email"
	CallerFieldExpectedSubject CallerField = "expected_subject"
	CallerFieldTopicARN        CallerField = "topic_arn"
)

type Defaults struct {
	Auth           AuthType
	Envelope       EnvelopeType
	AuthConfig     AuthConfig
	TypeFrom       envelope.FieldSelector
	IDFrom         envelope.FieldSelector
	CallerFields   []CallerField
	NativeProtocol bool
	CallerSecret   bool
}

type Handshake struct {
	Status int
	Body   map[string]string
}

type Provider interface {
	Name() string
	Defaults() Defaults
	Decoder(envelopeType EnvelopeType) (envelope.Decoder, error)
	Verifier(auth AuthType, cfg AuthConfig) signature.Verifier
	Handshake(event *envelope.Event) (*Handshake, error)
	RejectStatus(err error) int
}

type Base struct {
	ProviderName     string
	ProviderDefaults Defaults
}

var _ Provider = Base{}

func (b Base) Name() string { return b.ProviderName }

func (b Base) Defaults() Defaults { return b.ProviderDefaults }

func (b Base) Decoder(envelopeType EnvelopeType) (envelope.Decoder, error) {
	switch envelopeType {
	case EnvelopeNone:
		return envelope.Raw{}, nil
	case EnvelopeCloudEvents:
		return envelope.CloudEvents{}, nil
	case EnvelopePubSubPush:
		return envelope.PubSubPush{}, nil
	case EnvelopeSNS:
		return sns.Decoder{}, nil
	default:
		return nil, ErrUnsupportedEnvelope
	}
}

func (b Base) Verifier(auth AuthType, cfg AuthConfig) signature.Verifier {
	switch auth {
	case AuthHMAC:
		if cfg.Header == "" {
			cfg = AuthConfig{Header: "X-Nuon-Signature", Prefix: "v1=", Algorithm: "sha256", Encoding: "hex"}
		}
		return signature.HMAC{Header: cfg.Header, Prefix: cfg.Prefix, Algorithm: cfg.Algorithm, Encoding: cfg.Encoding}
	case AuthAPIKey:
		return signature.APIKey{Header: cfg.Header, Prefix: cfg.Prefix}
	case AuthBasic:
		return signature.Basic{Username: cfg.Username}
	default:
		return nil
	}
}

func (b Base) Handshake(*envelope.Event) (*Handshake, error) { return nil, nil }

func (b Base) RejectStatus(error) int { return http.StatusAccepted }

var ErrConflictingAuthConfig = errors.New("conflicting auth_config")

func ApplyDefaults(defaults Defaults, provided AuthConfig) (AuthConfig, error) {
	desired := defaults.AuthConfig
	for _, field := range defaults.CallerFields {
		switch field {
		case CallerFieldAudience:
			desired.Audience = provided.Audience
		case CallerFieldExpectedEmail:
			desired.ExpectedEmail = provided.ExpectedEmail
		case CallerFieldExpectedSubject:
			desired.ExpectedSubject = provided.ExpectedSubject
		case CallerFieldTopicARN:
			desired.TopicARN = provided.TopicARN
		}
	}
	if provided.Header == "" {
		provided.Header = desired.Header
	}
	if provided.Prefix == "" {
		provided.Prefix = desired.Prefix
	}
	if provided.Encoding == "" {
		provided.Encoding = desired.Encoding
	}
	if provided.Algorithm == "" {
		provided.Algorithm = desired.Algorithm
	}
	if provided.Username == "" {
		provided.Username = desired.Username
	}
	if provided.Issuer == "" {
		provided.Issuer = desired.Issuer
	}
	if !reflect.DeepEqual(provided, desired) {
		return AuthConfig{}, ErrConflictingAuthConfig
	}
	return desired, nil
}
