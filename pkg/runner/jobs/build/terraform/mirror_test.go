package terraform

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// gitMissingDiagnosticFixture is the literal byte sequence terraform get
// streamed to stderr for the apigateway-v2 module download failure
// captured in /Users/harsh/Downloads/Logs from Nuon.txt (lines 234-245).
// We keep the ANSI color escapes intact so the test exercises the same
// strip path the runner hits — tfexec gives us no clean way to pass
// -no-color for the build-vendoring subcommands.
const gitMissingDiagnosticFixture = "\x1b[31m\x1b[31m╷\x1b[0m\x1b[0m\n" +
	"\x1b[31m│\x1b[0m \x1b[0m\x1b[1m\x1b[31mError: \x1b[0m\x1b[0m\x1b[1mFailed to download module\x1b[0m\n" +
	"\x1b[31m│\x1b[0m \x1b[0m\n" +
	"\x1b[31m│\x1b[0m \x1b[0m\x1b[0m  on main.tf line 2:\n" +
	"\x1b[31m│\x1b[0m \x1b[0m   2: \x1b[4mmodule \"api_gateway\"\x1b[0m {\x1b[0m\n" +
	"\x1b[31m│\x1b[0m \x1b[0m\n" +
	"\x1b[31m│\x1b[0m \x1b[0mCould not download module \"api_gateway\" (main.tf:2) source code from\n" +
	"\x1b[31m│\x1b[0m \x1b[0m\"git::https://github.com/terraform-aws-modules/terraform-aws-apigateway-v2?ref=c62c315eeab078913c51d7d6a5eb722f4c1e82f5\":\n" +
	"\x1b[31m│\x1b[0m \x1b[0merror downloading\n" +
	"\x1b[31m│\x1b[0m \x1b[0m'https://github.com/terraform-aws-modules/terraform-aws-apigateway-v2?ref=c62c315eeab078913c51d7d6a5eb722f4c1e82f5':\n" +
	"\x1b[31m│\x1b[0m \x1b[0mgit must be available and on the PATH.\n" +
	"\x1b[31m╵\x1b[0m\x1b[0m\n"

func observedLogger(t *testing.T) (*zap.Logger, *observer.ObservedLogs) {
	t.Helper()
	core, obs := observer.New(zapcore.DebugLevel)
	return zap.New(core), obs
}

// TestRunTF_FailureWithDiagnostic locks in the real-world shape of a
// failed terraform command: one Warn entry whose `stderr` field carries
// the entire diagnostic block intact (envelope and all), with ANSI
// escapes stripped so the body is grep-friendly.
func TestRunTF_FailureWithDiagnostic(t *testing.T) {
	l, obs := observedLogger(t)
	wantErr := errors.New("exit status 1")

	err := runTF(l, "get", func(stdout, stderr io.Writer) error {
		_, _ = stderr.Write([]byte(gitMissingDiagnosticFixture))
		return wantErr
	})
	require.ErrorIs(t, err, wantErr)

	entries := obs.All()
	require.Len(t, entries, 1)
	got := entries[0]
	assert.Equal(t, zapcore.WarnLevel, got.Level)
	assert.Equal(t, "terraform get failed", got.Message)

	// Pull out the captured fields.
	fields := got.ContextMap()
	stderrField, ok := fields["stderr"].(string)
	require.True(t, ok, "expected stderr field on failure entry")
	assert.NotContains(t, fields, "stdout", "stdout should be omitted when empty")
	assert.Contains(t, fields, "error")

	// ANSI escapes must be stripped.
	assert.NotContains(t, stderrField, "\x1b[", "ANSI escapes should be stripped")

	// All substantive content from the diagnostic survives intact.
	assert.Contains(t, stderrField, "╷")
	assert.Contains(t, stderrField, "│ Error: Failed to download module")
	assert.Contains(t, stderrField, "on main.tf line 2:")
	assert.Contains(t, stderrField, "module \"api_gateway\"")
	assert.Contains(t, stderrField, "git::https://github.com/terraform-aws-modules/terraform-aws-apigateway-v2?ref=c62c315eeab078913c51d7d6a5eb722f4c1e82f5")
	assert.Contains(t, stderrField, "git must be available and on the PATH.")
	assert.Contains(t, stderrField, "╵")
}

// TestRunTF_SuccessWithStdout confirms the happy path: one Info entry,
// stdout captured as a field, no stderr field, no error.
func TestRunTF_SuccessWithStdout(t *testing.T) {
	l, obs := observedLogger(t)

	err := runTF(l, "get", func(stdout, stderr io.Writer) error {
		_, _ = stdout.Write([]byte("Downloading registry.terraform.io/terraform-aws-modules/apigateway-v2/aws 5.3.1 for api_gateway...\n"))
		return nil
	})
	require.NoError(t, err)

	entries := obs.All()
	require.Len(t, entries, 1)
	got := entries[0]
	assert.Equal(t, zapcore.InfoLevel, got.Level)
	assert.Equal(t, "terraform get completed", got.Message)

	fields := got.ContextMap()
	stdoutField, ok := fields["stdout"].(string)
	require.True(t, ok, "expected stdout field on success entry")
	assert.Equal(t, "Downloading registry.terraform.io/terraform-aws-modules/apigateway-v2/aws 5.3.1 for api_gateway...", stdoutField)
	assert.NotContains(t, fields, "stderr", "stderr should be omitted when empty")
	assert.NotContains(t, fields, "error")
}

// TestRunTF_SilentSuccess confirms commands that emit no output produce
// a single Info entry with no stdout/stderr/error fields.
func TestRunTF_SilentSuccess(t *testing.T) {
	l, obs := observedLogger(t)

	err := runTF(l, "providers mirror", func(stdout, stderr io.Writer) error {
		return nil
	})
	require.NoError(t, err)

	entries := obs.All()
	require.Len(t, entries, 1)
	assert.Equal(t, zapcore.InfoLevel, entries[0].Level)
	assert.Equal(t, "terraform providers mirror completed", entries[0].Message)
	fields := entries[0].ContextMap()
	assert.Empty(t, fields, "no fields expected for silent success")
}

// TestStripANSI is a focused unit test for the ANSI escape regex —
// terraform's diagnostics use SGR sequences exclusively, so we don't
// need to handle other CSI categories.
func TestStripANSI(t *testing.T) {
	cases := map[string]string{
		"":                                      "",
		"plain text":                            "plain text",
		"\x1b[31mred\x1b[0m":                    "red",
		"\x1b[1m\x1b[31mbold red\x1b[0m\x1b[0m": "bold red",
		"prefix \x1b[4munderline\x1b[0m suffix": "prefix underline suffix",
		// Multi-line preserved.
		"line1\n\x1b[31mline2\x1b[0m\nline3": "line1\nline2\nline3",
	}
	for in, want := range cases {
		got := stripANSI(in)
		assert.Equal(t, want, got, "stripANSI(%q)", in)
	}
}

// TestScrubbedEnvMap confirms the env scrubber drops the TF redirection
// vars and preserves everything else.
func TestScrubbedEnvMap(t *testing.T) {
	in := []string{
		"PATH=/usr/bin:/bin",
		"TF_CLI_CONFIG_FILE=/host/.terraformrc",
		"TF_PLUGIN_CACHE_DIR=/host/cache",
		"AWS_REGION=us-west-2",
		"NOEQ_INVALID", // no '=' — should be dropped silently
	}
	got := scrubbedEnvMap(in)
	assert.Equal(t, "/usr/bin:/bin", got["PATH"])
	assert.Equal(t, "us-west-2", got["AWS_REGION"])
	assert.NotContains(t, got, "TF_CLI_CONFIG_FILE")
	assert.NotContains(t, got, "TF_PLUGIN_CACHE_DIR")
	assert.NotContains(t, got, "NOEQ_INVALID")

	// Ensure no value in the scrubbed map carries the prohibited
	// TF_CLI_ARGS prefix that tfexec.SetEnv would reject.
	for k := range got {
		assert.False(t, strings.HasPrefix(k, "TF_CLI_ARGS"), "TF_CLI_ARGS leaked through scrubber")
	}
}
