package preflight

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	svcconfig "github.com/nuonco/nuon/pkg/services/config"
	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

// TestFieldNamesExistInConfig is the guard that replaces the `preflight:"..."`
// struct tag this package used to rely on. The tag kept check membership next
// to the field, but silently survived a rename and vanished entirely in a bad
// merge. Here a stale or misspelled key fails the build instead.
func TestFieldNamesExistInConfig(t *testing.T) {
	known := configKeys(reflect.TypeOf(internal.Config{}))
	require.NotEmpty(t, known, "no config keys discovered — reflection walk is broken")

	cfg := &internal.Config{}
	for _, check := range All() {
		for _, field := range check.Fields(cfg) {
			assert.Contains(t, known, field.Name,
				"check %q reads %q, which is not a config key on internal.Config",
				check.Name, field.Name)
		}
	}
}

// configKeys collects every `config:"..."` key on the struct, following
// embedded structs so squashed config (worker.Config) is included.
func configKeys(t reflect.Type) map[string]bool {
	keys := map[string]bool{}

	for i := range t.NumField() {
		field := t.Field(i)

		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			for k := range configKeys(field.Type) {
				keys[k] = true
			}
			continue
		}

		name, _, _ := strings.Cut(field.Tag.Get("config"), ",")
		if name == "" {
			name = strings.ToLower(field.Name)
		}
		keys[name] = true
	}

	return keys
}

func TestRegistryIsWellFormed(t *testing.T) {
	seen := map[string]bool{}

	for _, check := range All() {
		assert.NotEmpty(t, check.Name)
		assert.NotEmpty(t, check.Description, "check %q has no description", check.Name)
		assert.NotNil(t, check.Fields, "check %q has no Fields", check.Name)
		assert.NotNil(t, check.Probe, "check %q has no Probe", check.Name)

		assert.False(t, seen[check.Name], "duplicate check %q", check.Name)
		seen[check.Name] = true

		found, ok := Lookup(check.Name)
		require.True(t, ok, "check %q is not resolvable by name", check.Name)
		assert.Equal(t, check.Name, found.Name)
	}
}

func TestSkipPredicates(t *testing.T) {
	for _, tt := range []struct {
		name   string
		check  string
		cfg    *internal.Config
		skip   bool
		reason string
	}{
		{
			name:   "aws skips on gcp",
			check:  "aws",
			cfg:    &internal.Config{CloudProvider: "gcp"},
			skip:   true,
			reason: "cloud_provider=gcp",
		},
		{
			name:   "aws skips on azure",
			check:  "aws",
			cfg:    &internal.Config{CloudProvider: "azure"},
			skip:   true,
			reason: "cloud_provider=azure",
		},
		{
			// IsAWS treats an unset provider as AWS, matching NewConfig.
			name:  "aws runs when provider is unset",
			check: "aws",
			cfg:   &internal.Config{},
			skip:  false,
		},
		{
			name:   "kafka skips when disabled",
			check:  "kafka",
			cfg:    &internal.Config{},
			skip:   true,
			reason: "kafka_enabled=false",
		},
		{
			name:  "kafka runs when enabled",
			check: "kafka",
			cfg:   &internal.Config{KafkaEnabled: true},
			skip:  false,
		},
		{
			name:   "slack skips without a client id",
			check:  "slack",
			cfg:    &internal.Config{},
			skip:   true,
			reason: "slack app not configured",
		},
		{
			name:  "slack runs with a client id",
			check: "slack",
			cfg:   &internal.Config{SlackClientID: "abc"},
			skip:  false,
		},
		{
			name:   "nuon-auth skips when unconfigured",
			check:  "nuon-auth",
			cfg:    &internal.Config{},
			skip:   true,
			reason: "nuon auth not configured",
		},
		{
			name:  "nuon-auth runs with an issuer",
			check: "nuon-auth",
			cfg:   &internal.Config{NuonAuthIssuerURL: "https://example.com"},
			skip:  false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			check, ok := Lookup(tt.check)
			require.True(t, ok)
			require.NotNil(t, check.Skip)

			reason, skip := check.Skip(tt.cfg)
			assert.Equal(t, tt.skip, skip)
			if tt.skip {
				assert.Equal(t, tt.reason, reason)
			}
		})
	}
}

// Requirements that vary with other config are the thing struct tags could not
// express, so they get explicit coverage.
func TestConditionalRequirements(t *testing.T) {
	required := func(fields []Field, name string) bool {
		for _, f := range fields {
			if f.Name == name {
				return f.Required
			}
		}
		t.Fatalf("field %q not found", name)
		return false
	}

	rds, _ := Lookup("rds")
	assert.True(t, required(rds.Fields(&internal.Config{}), "db_password"),
		"db_password is required without IAM auth")
	assert.False(t, required(rds.Fields(&internal.Config{DBUseIAM: true}), "db_password"),
		"db_password is not required when the token is minted per connection")
	assert.True(t, required(rds.Fields(&internal.Config{DBUseIAM: true}), "db_region"),
		"db_region is required to mint an IAM token")

	kafka, _ := Lookup("kafka")
	plaintext := kafka.Fields(&internal.Config{KafkaSecurityProtocol: "PLAINTEXT"})
	assert.False(t, required(plaintext, "kafka_tls_cert_path"))
	ssl := kafka.Fields(&internal.Config{KafkaSecurityProtocol: "SSL"})
	assert.True(t, required(ssl, "kafka_tls_cert_path"))
}

// Google and GitHub carry fixed OAuth endpoints, so nuon_auth_issuer_url being
// empty is correct for them rather than a missing-config failure. Mirrors
// getDefaultIdentityProvider in app/auth/service/identity_providers.go.
func TestNuonAuthIssuerRequiredOnlyForOIDC(t *testing.T) {
	check, ok := Lookup("nuon-auth")
	require.True(t, ok)

	issuerRequired := func(providerType string) bool {
		fields := check.Fields(&internal.Config{NuonAuthProviderType: providerType})
		for _, f := range fields {
			if f.Name == "nuon_auth_issuer_url" {
				return f.Required
			}
		}
		t.Fatal("nuon_auth_issuer_url not found")
		return false
	}

	assert.False(t, issuerRequired("google"))
	assert.False(t, issuerRequired("github"))
	assert.True(t, issuerRequired("oidc"))
	assert.True(t, issuerRequired(""), "an unset type falls back to the generic OIDC requirement")
}

// A provider that cannot be confirmed over the network warns rather than fails,
// so an otherwise healthy google deployment still exits 0.
func TestNuonAuthGoogleWarnsWithoutIssuer(t *testing.T) {
	check, ok := Lookup("nuon-auth")
	require.True(t, ok)

	cfg := &internal.Config{
		NuonAuthProviderType: "google",
		NuonAuthClientID:     "client-id",
		NuonAuthClientSecret: "client-secret",
		NuonAuthRedirectURL:  "https://example.com/callback",
		NuonAuthSessionKey:   "session-key",
	}

	result := run(t.Context(), cfg, check)

	assert.Equal(t, StatusWarn, result.Status)
	assert.Contains(t, result.Detail, "google credentials present but unverified")
}

func TestNuonAuthRejectsUnknownProviderType(t *testing.T) {
	check, ok := Lookup("nuon-auth")
	require.True(t, ok)

	cfg := &internal.Config{
		NuonAuthProviderType: "okta",
		NuonAuthIssuerURL:    "https://example.com",
		NuonAuthClientID:     "client-id",
		NuonAuthClientSecret: "client-secret",
		NuonAuthRedirectURL:  "https://example.com/callback",
		NuonAuthSessionKey:   "session-key",
	}

	result := run(t.Context(), cfg, check)

	assert.Equal(t, StatusFail, result.Status)
	assert.Contains(t, result.Detail, "invalid nuon_auth_provider_type")
}

// Production overrides the registered slack defaults, so a value still equal to
// one there means signature verification is forgeable. Elsewhere the default is
// the intended value.
func TestSlackDevDefaultsFailOnlyInProduction(t *testing.T) {
	check, ok := Lookup("slack")
	require.True(t, ok)

	withEnv := func(env svcconfig.Env) *internal.Config {
		cfg := &internal.Config{
			SlackClientID:         "client-id",
			SlackClientSecret:     "client-secret",
			SlackOAuthRedirectURL: "https://example.com/slack/oauth/callback",
			SlackSigningSecret:    devSlackSigningSecret,
			SlackStateJWTSecret:   "real-state-secret",
		}
		cfg.Env = env

		return cfg
	}

	prod := run(t.Context(), withEnv(svcconfig.Production), check)
	assert.Equal(t, StatusFail, prod.Status)
	assert.Contains(t, prod.Detail, "insecure dev default in production for slack_signing_secret")

	for _, env := range []svcconfig.Env{svcconfig.Development, svcconfig.Stage} {
		result := run(t.Context(), withEnv(env), check)
		assert.Equal(t, StatusWarn, result.Status, "env %s", env)
		assert.Contains(t, result.Detail, "slack_signing_secret")
	}
}

// Slack is an optional integration: an absent app must never fail a run.
func TestSlackSkipsWhenUnconfigured(t *testing.T) {
	check, ok := Lookup("slack")
	require.True(t, ok)

	result := run(t.Context(), &internal.Config{}, check)

	assert.Equal(t, StatusSkipped, result.Status)
	assert.Equal(t, "slack app not configured", result.Detail)
}

func TestWarnIsNotAFailure(t *testing.T) {
	var buf bytes.Buffer

	code := PrintResults(&buf, []Result{
		{Name: "a", Status: StatusPass},
		{Name: "b", Status: StatusWarn, Detail: "unverified"},
	})

	assert.Equal(t, 0, code)
	assert.Contains(t, buf.String(), "1 passed, 1 warned, 0 failed, 0 skipped")
}

func TestFieldDisplayRedactsSecrets(t *testing.T) {
	assert.Equal(t, "******", Field{Name: "k", Value: "hunter2", Secret: true}.Display())
	assert.Equal(t, "(unset)", Field{Name: "k", Secret: true}.Display())
	assert.Equal(t, "plain", Field{Name: "k", Value: "plain"}.Display())
}

func TestResolveUnknownCheck(t *testing.T) {
	checks, unknown := resolve([]string{"rds", "nope"})

	require.Len(t, checks, 1)
	assert.Equal(t, "rds", checks[0].Name)
	assert.Equal(t, []string{"nope"}, unknown)
}

// resolve follows registry order so the table does not reshuffle between runs.
func TestResolvePreservesRegistryOrder(t *testing.T) {
	checks, _ := resolve([]string{"slack", "rds", "temporal"})

	names := make([]string, 0, len(checks))
	for _, c := range checks {
		names = append(names, c.Name)
	}
	assert.Equal(t, []string{"rds", "temporal", "slack"}, names)
}

// A missing required field short-circuits before the probe, so an unset host
// reports the config key rather than a connection-refused error.
func TestRunSkipsProbeWhenRequiredConfigIsMissing(t *testing.T) {
	probed := false
	check := Check{
		Name:        "example",
		Description: "example",
		Fields: func(*internal.Config) []Field {
			return []Field{
				{Name: "present_key", Value: "set", Required: true},
				{Name: "missing_key", Required: true},
			}
		},
		Probe: func(context.Context, *internal.Config) (string, error) {
			probed = true
			return "", nil
		},
	}

	result := run(t.Context(), &internal.Config{}, check)

	assert.False(t, probed, "probe ran despite missing config")
	assert.Equal(t, StatusFail, result.Status)
	assert.Equal(t, "missing required config: missing_key", result.Detail)
	assert.Len(t, result.Fields, 2, "fields are attached so the printer can expand them")
}

func TestRunSkipsBeforeValidatingFields(t *testing.T) {
	check := Check{
		Name:        "example",
		Description: "example",
		Skip:        func(*internal.Config) (string, bool) { return "not applicable", true },
		Fields: func(*internal.Config) []Field {
			return []Field{{Name: "missing_key", Required: true}}
		},
		Probe: func(context.Context, *internal.Config) (string, error) {
			t.Fatal("probe ran for a skipped check")
			return "", nil
		},
	}

	result := run(t.Context(), &internal.Config{}, check)

	assert.Equal(t, StatusSkipped, result.Status)
	assert.Equal(t, "not applicable", result.Detail)
}

func TestPrintResultsExitCode(t *testing.T) {
	var buf bytes.Buffer

	assert.Equal(t, 0, PrintResults(&buf, []Result{
		{Name: "a", Status: StatusPass},
		{Name: "b", Status: StatusSkipped},
	}), "a skipped check is not a failure")

	assert.Equal(t, 1, PrintResults(&buf, []Result{
		{Name: "a", Status: StatusPass},
		{Name: "b", Status: StatusFail},
	}))
}

func TestPrintResultsRedactsSecretsInExpandedFields(t *testing.T) {
	var buf bytes.Buffer

	PrintResults(&buf, []Result{{
		Name:   "example",
		Status: StatusFail,
		Detail: "boom",
		Fields: []Field{{Name: "some_secret", Value: "hunter2", Secret: true}},
	}})

	assert.Contains(t, buf.String(), "some_secret")
	assert.NotContains(t, buf.String(), "hunter2")
}
