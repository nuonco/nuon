package urlpolicy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEntry(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		allowHTTP bool
		want      string
		wantErr   string
	}{
		{name: "origin and path prefix", raw: "https://acme-artifacts.s3.us-west-2.amazonaws.com/runner/", want: "https://acme-artifacts.s3.us-west-2.amazonaws.com/runner/"},
		{name: "bare host", raw: "example.com", want: "https://example.com/"},
		{name: "bare host with port and path", raw: "example.com:8443/templates", want: "https://example.com:8443/templates"},
		{name: "wildcard", raw: "https://*.example.com/templates/", want: "https://*.example.com/templates/"},
		{name: "bare wildcard", raw: "*.example.com", want: "https://*.example.com/"},
		{name: "host is lowercased", raw: "https://EXAMPLE.com", want: "https://example.com/"},
		{name: "default port is dropped", raw: "https://example.com:443/a", want: "https://example.com/a"},
		{name: "idna", raw: "https://bücher.example/", want: "https://xn--bcher-kva.example/"},
		{name: "http when allowed", raw: "http://localhost:8083", allowHTTP: true, want: "http://localhost:8083/"},
		{name: "http rejected", raw: "http://example.com", wantErr: "must use https"},
		{name: "other scheme", raw: "s3://bucket/key", wantErr: "must use https"},
		{name: "empty", raw: "  ", wantErr: "empty"},
		{name: "credentials", raw: "https://user:pass@example.com/", wantErr: "credentials"},
		{name: "query", raw: "https://example.com/a?b=c", wantErr: "query or fragment"},
		{name: "fragment", raw: "https://example.com/a#b", wantErr: "query or fragment"},
		{name: "inner wildcard", raw: "https://a.*.example.com/", wantErr: "wildcard"},
		{name: "wildcard ip", raw: "https://*.10.0.0.1/", wantErr: "wildcard"},
		{name: "dot dot", raw: "https://example.com/a/../b/", wantErr: "invalid path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseEntry(tt.raw, ParseOptions{AllowHTTP: tt.allowHTTP})
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, entry.String())
		})
	}
}

func TestPolicyCheck(t *testing.T) {
	policy, err := New(Config{
		Categories: map[Category]CategoryConfig{
			CategoryRunnerBootstrap: {
				Builtins: []string{"https://raw.githubusercontent.com/nuonco/runner/"},
				Entries: []string{
					"https://acme-artifacts.s3.us-west-2.amazonaws.com/runner",
					"https://*.example.com/scripts/",
					"https://example.org:8443/",
				},
			},
		},
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		url     string
		allowed bool
		reason  string
	}{
		{name: "empty is unset", url: "", allowed: true},
		{name: "builtin", url: "https://raw.githubusercontent.com/nuonco/runner/refs/heads/main/scripts/aws/init.sh", allowed: true},
		{name: "fragment ignored", url: "https://raw.githubusercontent.com/nuonco/runner/scripts/aws/init.sh#azure", allowed: true},
		{name: "query allowed", url: "https://acme-artifacts.s3.us-west-2.amazonaws.com/runner/init.sh?versionId=abc", allowed: true},
		{name: "prefix without slash matches itself", url: "https://acme-artifacts.s3.us-west-2.amazonaws.com/runner", allowed: true},
		{name: "prefix stops at a segment boundary", url: "https://acme-artifacts.s3.us-west-2.amazonaws.com/runner-evil/init.sh"},
		{name: "other owner on a shared host", url: "https://raw.githubusercontent.com/mallory/runner/init.sh"},
		{name: "path case matters", url: "https://raw.githubusercontent.com/NuonCo/runner/init.sh"},
		{name: "host case does not", url: "https://RAW.githubusercontent.com/nuonco/runner/init.sh", allowed: true},
		{name: "trailing dot host", url: "https://raw.githubusercontent.com./nuonco/runner/init.sh", allowed: true},
		{name: "explicit default port", url: "https://raw.githubusercontent.com:443/nuonco/runner/init.sh", allowed: true},
		{name: "wildcard subdomain", url: "https://cdn.example.com/scripts/init.sh", allowed: true},
		{name: "wildcard deeper subdomain", url: "https://a.b.example.com/scripts/init.sh", allowed: true},
		{name: "wildcard excludes apex", url: "https://example.com/scripts/init.sh"},
		{name: "wildcard excludes suffix lookalike", url: "https://evilexample.com/scripts/init.sh"},
		{name: "host suffix lookalike", url: "https://raw.githubusercontent.com.evil.test/nuonco/runner/init.sh"},
		{name: "port must match", url: "https://example.org/init.sh"},
		{name: "port matches", url: "https://example.org:8443/init.sh", allowed: true},
		{name: "http", url: "http://raw.githubusercontent.com/nuonco/runner/init.sh"},
		{name: "userinfo", url: "https://raw.githubusercontent.com@evil.test/nuonco/runner/init.sh"},
		{name: "userinfo on allowed host", url: "https://u:p@raw.githubusercontent.com/nuonco/runner/init.sh", reason: "credentials"},
		{name: "dot dot", url: "https://raw.githubusercontent.com/nuonco/runner/../../mallory/x.sh", reason: `".."`},
		{name: "encoded dot dot", url: "https://raw.githubusercontent.com/nuonco/runner/%2e%2e/%2e%2e/mallory/x.sh", reason: `".."`},
		{name: "encoded slash", url: "https://raw.githubusercontent.com/nuonco/runner/..%2f..%2fmallory/x.sh", reason: "encoded slash"},
		{name: "encoded backslash", url: "https://raw.githubusercontent.com/nuonco/runner/..%5Cx.sh", reason: "encoded slash"},
		{name: "git getter", url: "git::https://raw.githubusercontent.com/nuonco/runner//init.sh", reason: "scheme"},
		{name: "file", url: "file:///etc/passwd", reason: "scheme"},
		{name: "relative", url: "./scripts/init.sh", reason: "absolute https URL"},
		{name: "scheme relative", url: "//raw.githubusercontent.com/nuonco/runner/init.sh", reason: "absolute https URL"},
		{name: "template", url: "{{ .nuon.inputs.url }}", reason: "absolute https URL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := policy.Check(CategoryRunnerBootstrap, "runner.init_script_url", tt.url)
			if tt.allowed {
				require.Nil(t, violation)
				return
			}
			require.NotNil(t, violation)
			require.Equal(t, CategoryRunnerBootstrap, violation.Category)
			require.Contains(t, violation.Reason, tt.reason)
			require.Contains(t, violation.Error(), "allowed_hosts_runner_bootstrap")
		})
	}
}

func TestPolicyUnrestricted(t *testing.T) {
	var nilPolicy *Policy
	require.Nil(t, nilPolicy.Check(CategoryRunnerAPI, "runner.runner_api_url", "https://evil.test"))
	require.False(t, nilPolicy.AnyRestricted())
	require.Equal(t, ModeEnforce, nilPolicy.Mode())

	// Built-ins alone never restrict a category.
	policy, err := New(Config{
		Categories: map[Category]CategoryConfig{
			CategoryRunnerAPI:       {Builtins: []string{"https://api.example.com"}},
			CategoryStackTemplates:  {Entries: []string{"", "  "}},
			CategoryRunnerBootstrap: {Entries: []string{"https://example.com/"}},
		},
	})
	require.NoError(t, err)
	require.False(t, policy.Restricted(CategoryRunnerAPI))
	require.False(t, policy.Restricted(CategoryStackTemplates))
	require.True(t, policy.Restricted(CategoryRunnerBootstrap))
	require.True(t, policy.AnyRestricted())
	require.Nil(t, policy.Check(CategoryRunnerAPI, "runner.runner_api_url", "git::whatever"))
	require.Empty(t, policy.Entries(CategoryRunnerAPI))
}

func TestPolicyBuiltinsAreAddedOnceRestricted(t *testing.T) {
	policy, err := New(Config{
		Categories: map[Category]CategoryConfig{
			CategoryRunnerAPI: {
				Builtins: []string{"https://runner.example.com"},
				// Restating a built-in is how an operator locks a category to the built-ins.
				Entries: []string{"https://runner.example.com/"},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"https://runner.example.com/"}, policy.Entries(CategoryRunnerAPI))
	require.Nil(t, policy.Check(CategoryRunnerAPI, "runner.runner_api_url", "https://runner.example.com"))
	require.NotNil(t, policy.Check(CategoryRunnerAPI, "runner.runner_api_url", "https://evil.test"))
}

func TestNewRejectsBadInput(t *testing.T) {
	_, err := New(Config{Mode: "audit"})
	require.ErrorContains(t, err, "allowed_hosts_mode")

	_, err = New(Config{Categories: map[Category]CategoryConfig{
		CategoryStackTemplates: {Entries: []string{"http://example.com"}},
	}})
	require.ErrorContains(t, err, "allowed_hosts_stack_templates")
	require.ErrorContains(t, err, "must use https")

	_, err = New(Config{Categories: map[Category]CategoryConfig{"nope": {Entries: []string{"example.com"}}}})
	require.ErrorContains(t, err, "unknown")

	policy, err := New(Config{AllowHTTP: true, Mode: ModeWarn, Categories: map[Category]CategoryConfig{
		CategoryRunnerAPI: {Entries: []string{"http://localhost:8083"}},
	}})
	require.NoError(t, err)
	require.Equal(t, ModeWarn, policy.Mode())
	require.Nil(t, policy.Check(CategoryRunnerAPI, "f", "http://localhost:8083/v1"))
	require.NotNil(t, policy.Check(CategoryRunnerAPI, "f", "https://localhost:8083/v1"))
}

func TestRunnerEnvVars(t *testing.T) {
	restricted, err := New(Config{Categories: map[Category]CategoryConfig{
		CategoryRunnerAPI: {Entries: []string{"https://runner.example.com"}},
	}})
	require.NoError(t, err)
	unrestricted, err := New(Config{})
	require.NoError(t, err)

	for _, name := range []string{
		"HTTPS_PROXY", "https_proxy", "NO_PROXY", "ALL_PROXY", "RUNNER_BINARY_VERSION", "NUON_API_URL",
		"AWS_ENDPOINT_URL", "AWS_ENDPOINT_URL_STS", "AWS_ACCESS_KEY_ID", "AWS_CONTAINER_CREDENTIALS_FULL_URI",
	} {
		require.NotNil(t, restricted.CheckRunnerEnvVar("runner.env_vars."+name, name), name)
		require.Nil(t, unrestricted.CheckRunnerEnvVar("runner.env_vars."+name, name), name)
	}
	for _, name := range []string{"DEBUG", "LOG_LEVEL", "AWS_REGION", "PROXY_MODE", "RUNNER_AUTH_METHOD"} {
		require.Nil(t, restricted.CheckRunnerEnvVar("runner.env_vars."+name, name), name)
	}
}

func TestOnWarn(t *testing.T) {
	policy, err := New(Config{Mode: ModeWarn})
	require.NoError(t, err)

	var got Violations
	policy.OnWarn(func(v Violations) { got = v })
	policy.ReportWarnings(nil)
	require.Empty(t, got)
	policy.ReportWarnings(Violations{{Field: "a"}})
	require.Len(t, got, 1)
}

func TestFromEffective(t *testing.T) {
	policy, err := FromEffective(ModeWarn, map[string][]string{
		"runner_api":       {"http://localhost:8083/"},
		"stack_templates":  {},
		"a_later_category": {"https://example.com/"},
	})
	require.NoError(t, err)
	require.Equal(t, ModeWarn, policy.Mode())
	require.True(t, policy.Restricted(CategoryRunnerAPI))
	require.False(t, policy.Restricted(CategoryStackTemplates))
	require.Nil(t, policy.Check(CategoryRunnerAPI, "f", "http://localhost:8083/v1"))
	require.NotNil(t, policy.Check(CategoryRunnerAPI, "f", "https://evil.test"))

	_, err = FromEffective("audit", nil)
	require.Error(t, err)
}

func TestNew_HTTPBuiltinDoesNotFailARestrictedCategory(t *testing.T) {
	policy, err := New(Config{Categories: map[Category]CategoryConfig{
		CategoryRunnerAPI: {
			Entries:  []string{"https://runner.example.com"},
			Builtins: []string{"http://ctl-api.internal:8083"},
		},
	}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if v := policy.Check(CategoryRunnerAPI, "runner.runner_api_url", "http://ctl-api.internal:8083"); v != nil {
		t.Fatalf("built-in origin rejected: %v", v)
	}
	if v := policy.Check(CategoryRunnerAPI, "runner.runner_api_url", "http://runner.example.com"); v == nil {
		t.Fatal("http URL outside the built-in was allowed")
	}

	_, err = New(Config{Categories: map[Category]CategoryConfig{
		CategoryRunnerAPI: {Entries: []string{"http://runner.example.com"}},
	}})
	if err == nil {
		t.Fatal("http operator entry was accepted without AllowHTTP")
	}
}
