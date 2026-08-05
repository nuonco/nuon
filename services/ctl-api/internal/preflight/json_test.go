package preflight

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

// The whole point of the DTO: no path through the encoder can publish a secret.
// Runs over every real check so a new one cannot opt out by accident.
func TestJSONNeverEmitsSecretValues(t *testing.T) {
	const sentinel = "SUPER-SECRET-SENTINEL"

	cfg := &internal.Config{
		DBPassword:            sentinel,
		ClickhouseDBPassword:  sentinel,
		GithubAppKey:          sentinel,
		NuonAuthClientSecret:  sentinel,
		NuonAuthSessionKey:    sentinel,
		NuonAuthProviderType:  "google",
		NuonAuthClientID:      "client-id",
		NuonAuthRedirectURL:   "https://example.com/cb",
		SlackClientID:         "client-id",
		SlackClientSecret:     sentinel,
		SlackSigningSecret:    sentinel,
		SlackStateJWTSecret:   sentinel,
		SlackOAuthRedirectURL: "https://example.com/slack",
	}

	var buf bytes.Buffer
	require.NoError(t, WriteJSONChecks(&buf, Describe(cfg, nil)))

	assert.NotContains(t, buf.String(), sentinel)

	// ...and the secrets really were present, so the assertion above means
	// something.
	var report jsonReport
	require.NoError(t, json.Unmarshal(buf.Bytes(), &report))

	var secretsSeen int
	for _, c := range report.Checks {
		for _, f := range c.Fields {
			if f.Secret {
				secretsSeen++
				assert.Empty(t, f.Value, "%s.%s published a value", c.Name, f.Name)
			}
		}
	}
	assert.Positive(t, secretsSeen, "no secret fields exercised")
}

// set distinguishes configured-but-hidden from absent, which is all a caller
// can learn about a secret.
func TestJSONSetFlagTracksPresenceForSecrets(t *testing.T) {
	fields := toJSONFields([]Field{
		{Name: "with_value", Value: "v", Secret: true},
		{Name: "without_value", Secret: true},
		{Name: "plain", Value: "visible"},
	})

	assert.True(t, fields[0].Set)
	assert.Empty(t, fields[0].Value)

	assert.False(t, fields[1].Set)
	assert.Empty(t, fields[1].Value)

	assert.True(t, fields[2].Set)
	assert.Equal(t, "visible", fields[2].Value)
}

// --list did not run anything, so it must not claim a status or a tally.
func TestJSONChecksOmitStatusAndSummary(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, WriteJSONChecks(&buf, Describe(&internal.Config{}, []string{"rds"})))

	var raw map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &raw))

	assert.NotContains(t, raw, "summary")

	checks := raw["checks"].([]any)
	require.Len(t, checks, 1)
	assert.NotContains(t, checks[0].(map[string]any), "status")
}

func TestJSONResultsCarrySummaryAndExitCode(t *testing.T) {
	var buf bytes.Buffer

	code := WriteJSONResults(&buf, []Result{
		{Name: "a", Status: StatusPass},
		{Name: "b", Status: StatusWarn},
		{Name: "c", Status: StatusSkipped},
	})
	assert.Equal(t, 0, code, "only a failure is non-zero")

	var report jsonReport
	require.NoError(t, json.Unmarshal(buf.Bytes(), &report))
	require.NotNil(t, report.Summary)
	assert.Equal(t, jsonSummary{Passed: 1, Warned: 1, Failed: 0, Skipped: 1}, *report.Summary)

	buf.Reset()
	assert.Equal(t, 1, WriteJSONResults(&buf, []Result{{Name: "a", Status: StatusFail}}))
}

// A run and a listing must describe a check identically, so docs generated from
// either agree.
func TestJSONSkippedCheckStillCarriesFields(t *testing.T) {
	cfg := &internal.Config{}

	ran := run(t.Context(), cfg, mustLookup(t, "kafka"))
	listed := Describe(cfg, []string{"kafka"})[0]

	assert.Equal(t, StatusSkipped, ran.Status)
	assert.NotEmpty(t, ran.Fields)
	assert.Equal(t, listed.Fields, ran.Fields)
}

func TestJSONIsStableAcrossRuns(t *testing.T) {
	cfg := &internal.Config{}

	var first, second bytes.Buffer
	require.NoError(t, WriteJSONChecks(&first, Describe(cfg, nil)))
	require.NoError(t, WriteJSONChecks(&second, Describe(cfg, nil)))

	assert.Equal(t, first.String(), second.String())
	assert.True(t, strings.HasSuffix(second.String(), "\n"))
}

func mustLookup(t *testing.T, name string) Check {
	t.Helper()

	check, ok := Lookup(name)
	require.True(t, ok, "check %q not registered", name)

	return check
}
