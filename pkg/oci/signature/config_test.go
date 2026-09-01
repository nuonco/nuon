package signature

import (
	"strings"
	"testing"
)

func TestVerificationValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Verification
		wantErr string
	}{
		{name: "unset"},
		{name: "disabled", config: &Verification{}},
		{
			name: "keyless exact subject",
			config: &Verification{RequireSignature: true, Authorities: []Authority{{
				Type:    AuthorityTypeKeyless,
				Issuer:  "https://issuer.example.com",
				Subject: "https://example.com/workflow",
			}}},
		},
		{
			name: "keyless subject regexp",
			config: &Verification{RequireSignature: true, Authorities: []Authority{{
				Type:          AuthorityTypeKeyless,
				Issuer:        "https://issuer.example.com",
				SubjectRegexp: `^https://example\.com/.+$`,
			}}},
		},
		{
			name: "public key",
			config: &Verification{RequireSignature: true, Authorities: []Authority{{
				Type:      AuthorityTypePublicKey,
				PublicKey: "public key contents",
			}}},
		},
		{
			name:    "missing authorities",
			config:  &Verification{RequireSignature: true},
			wantErr: "at least one authority",
		},
		{
			name: "keyless missing issuer",
			config: &Verification{RequireSignature: true, Authorities: []Authority{{
				Type:    AuthorityTypeKeyless,
				Subject: "subject",
			}}},
			wantErr: "issuer is required",
		},
		{
			name: "keyless ambiguous subject",
			config: &Verification{RequireSignature: true, Authorities: []Authority{{
				Type:          AuthorityTypeKeyless,
				Issuer:        "issuer",
				Subject:       "subject",
				SubjectRegexp: ".*",
			}}},
			wantErr: "exactly one",
		},
		{
			name: "invalid regexp",
			config: &Verification{RequireSignature: true, Authorities: []Authority{{
				Type:          AuthorityTypeKeyless,
				Issuer:        "issuer",
				SubjectRegexp: "[",
			}}},
			wantErr: "invalid subject_regexp",
		},
		{
			name: "key missing public key",
			config: &Verification{RequireSignature: true, Authorities: []Authority{{
				Type: AuthorityTypePublicKey,
			}}},
			wantErr: "public_key is required",
		},
		{
			name: "unknown authority",
			config: &Verification{RequireSignature: true, Authorities: []Authority{{
				Type: "unknown",
			}}},
			wantErr: "unsupported type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("Validate() error = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}
