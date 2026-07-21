package kafka

import (
	"encoding/json"
	"testing"
)

func TestEnvelopeRoundTrip(t *testing.T) {
	type hb struct {
		ID       string `json:"id"`
		RunnerID string `json:"runner_id"`
	}
	in := hb{ID: "hb_123", RunnerID: "runner_abc"}

	b, err := Wrap(TypeRunnerHeartBeat, in)
	if err != nil {
		t.Fatalf("Wrap: %v", err)
	}

	env, err := Unwrap(b)
	if err != nil {
		t.Fatalf("Unwrap: %v", err)
	}

	if env.Version != EnvelopeVersion {
		t.Errorf("Version = %d, want %d", env.Version, EnvelopeVersion)
	}
	if env.Type != TypeRunnerHeartBeat {
		t.Errorf("Type = %q, want %q", env.Type, TypeRunnerHeartBeat)
	}
	if env.Source != envelopeSource {
		t.Errorf("Source = %q, want %q", env.Source, envelopeSource)
	}
	if env.ProducedAt.IsZero() {
		t.Error("ProducedAt is zero")
	}

	var out hb
	if err := json.Unmarshal(env.Payload, &out); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if out != in {
		t.Errorf("payload = %+v, want %+v", out, in)
	}
}

func TestBuildOpts(t *testing.T) {
	tests := []struct {
		name    string
		cfg     clientConfig
		wantErr bool
	}{
		{
			name:    "no brokers is an error",
			cfg:     clientConfig{},
			wantErr: true,
		},
		{
			name: "plaintext local",
			cfg:  clientConfig{Brokers: []string{"localhost:9092"}, ClientID: "ctl-api"},
		},
		{
			name: "plain sasl over ssl",
			cfg: clientConfig{
				Brokers:          []string{"b:9092"},
				SecurityProtocol: securitySASLSSL,
				SASLMechanism:    saslPlain,
				SASLUsername:     "u",
				SASLPassword:     "p",
			},
		},
		{
			name: "scram-512",
			cfg:  clientConfig{Brokers: []string{"b:9092"}, SASLMechanism: saslScram512, SASLUsername: "u", SASLPassword: "p"},
		},
		{
			name:    "msk iam not yet implemented",
			cfg:     clientConfig{Brokers: []string{"b:9092"}, SASLMechanism: saslAWSMSKIAM},
			wantErr: true,
		},
		{
			name:    "oauthbearer not yet implemented",
			cfg:     clientConfig{Brokers: []string{"b:9092"}, SASLMechanism: saslOAuth},
			wantErr: true,
		},
		{
			name:    "unknown mechanism",
			cfg:     clientConfig{Brokers: []string{"b:9092"}, SASLMechanism: "NOPE"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := tt.cfg.buildOpts()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("buildOpts: %v", err)
			}
			if len(opts) == 0 {
				t.Fatal("expected opts, got none")
			}
		})
	}
}
