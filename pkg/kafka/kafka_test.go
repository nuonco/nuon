package kafka

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
)

type testPayload struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestWrapUnwrapRoundTrip(t *testing.T) {
	orig := testPayload{ID: "abc123", Name: "widget"}

	b, err := Wrap("my-source", "widget.created", orig)
	require.NoError(t, err)
	require.NotEmpty(t, b)

	env, err := Unwrap(b)
	require.NoError(t, err)

	require.Equal(t, EnvelopeVersion, env.Version)
	require.Equal(t, "widget.created", env.Type)
	require.Equal(t, "my-source", env.Source)
	require.False(t, env.ProducedAt.IsZero())

	var got testPayload
	require.NoError(t, json.Unmarshal(env.Payload, &got))
	require.Equal(t, orig, got)
}

func TestWrapUnwrapRoundTrip_MapPayload(t *testing.T) {
	orig := map[string]any{"foo": "bar", "count": float64(3)}

	b, err := Wrap("map-source", "map.event", orig)
	require.NoError(t, err)

	env, err := Unwrap(b)
	require.NoError(t, err)

	require.Equal(t, EnvelopeVersion, env.Version)
	require.Equal(t, "map.event", env.Type)
	require.Equal(t, "map-source", env.Source)
	require.False(t, env.ProducedAt.IsZero())

	var got map[string]any
	require.NoError(t, json.Unmarshal(env.Payload, &got))
	require.Equal(t, orig, got)
}

func TestConfigBaseOpts(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "no brokers",
			cfg:     Config{},
			wantErr: true,
		},
		{
			name: "plaintext no sasl",
			cfg: Config{
				Brokers: []string{"localhost:9092"},
			},
			wantErr: false,
		},
		{
			name: "plain sasl_ssl",
			cfg: Config{
				Brokers:          []string{"broker:9092"},
				SecurityProtocol: securitySASLSSL,
				SASLMechanism:    saslPlain,
				SASLUsername:     "user",
				SASLPassword:     "pass",
			},
			wantErr: false,
		},
		{
			name: "scram-sha-256",
			cfg: Config{
				Brokers:          []string{"broker:9092"},
				SecurityProtocol: securitySASLSSL,
				SASLMechanism:    saslScram256,
				SASLUsername:     "user",
				SASLPassword:     "pass",
			},
			wantErr: false,
		},
		{
			name: "scram-sha-512",
			cfg: Config{
				Brokers:          []string{"broker:9092"},
				SecurityProtocol: securitySASLSSL,
				SASLMechanism:    saslScram512,
				SASLUsername:     "user",
				SASLPassword:     "pass",
			},
			wantErr: false,
		},
		{
			name: "aws msk iam not implemented",
			cfg: Config{
				Brokers:       []string{"broker:9092"},
				SASLMechanism: saslAWSMSKIAM,
			},
			wantErr: true,
		},
		{
			name: "oauthbearer not implemented",
			cfg: Config{
				Brokers:       []string{"broker:9092"},
				SASLMechanism: saslOAuth,
			},
			wantErr: true,
		},
		{
			name: "unknown mechanism",
			cfg: Config{
				Brokers:       []string{"broker:9092"},
				SASLMechanism: "NOPE",
			},
			wantErr: true,
		},
		{
			name: "lowercase plain",
			cfg: Config{
				Brokers:          []string{"broker:9092"},
				SecurityProtocol: securitySASLSSL,
				SASLMechanism:    "plain",
				SASLUsername:     "user",
				SASLPassword:     "pass",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := tt.cfg.baseOpts()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, opts)
		})
	}
}

func TestDedupToken(t *testing.T) {
	tests := []struct {
		name      string
		topic     string
		partition int32
		first     int64
		last      int64
		want      string
	}{
		{
			name:      "basic",
			topic:     "otel_log_records",
			partition: 3,
			first:     1000,
			last:      1999,
			want:      "otel_log_records:3:1000-1999",
		},
		{
			name:      "zero partition",
			topic:     "events",
			partition: 0,
			first:     0,
			last:      10,
			want:      "events:0:0-10",
		},
		{
			name:      "single record range",
			topic:     "spans",
			partition: 7,
			first:     42,
			last:      42,
			want:      "spans:7:42-42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DedupToken(tt.topic, tt.partition, tt.first, tt.last)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDisabledProducer(t *testing.T) {
	l := zap.NewNop()
	mw, err := metrics.New(validator.New(), metrics.WithDisable(true))
	require.NoError(t, err)

	p := DisabledProducer(l, mw)
	require.NotNil(t, p)
	require.False(t, p.Enabled())

	require.NoError(t, p.Ping(context.Background()))
	require.NoError(t, p.Flush(context.Background()))

	require.NotPanics(t, func() {
		p.Close()
	})
}
