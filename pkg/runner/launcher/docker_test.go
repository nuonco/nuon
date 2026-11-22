package launcher

import (
	"slices"
	"strings"
	"testing"
)

func TestContainerEnvArgs(t *testing.T) {
	t.Run("passes names on argv and values via env", func(t *testing.T) {
		args, vars, err := containerEnvArgs(map[string]string{
			"AWS_SESSION_TOKEN": "super-secret",
			"COLUMNS":           "500",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := []string{"--env", "AWS_SESSION_TOKEN", "--env", "COLUMNS"}
		if !slices.Equal(args, want) {
			t.Fatalf("expected %v, got %v", want, args)
		}

		// the secret must never reach argv, only the process env
		for _, a := range args {
			if strings.Contains(a, "super-secret") {
				t.Fatalf("secret value leaked into argv: %v", args)
			}
		}
		if !slices.Contains(vars, "AWS_SESSION_TOKEN=super-secret") {
			t.Fatalf("expected token in env vars, got %v", vars)
		}
	})

	t.Run("multiline value survives", func(t *testing.T) {
		pem := "-----BEGIN PRIVATE KEY-----\nabc\ndef\n-----END PRIVATE KEY-----"
		_, vars, err := containerEnvArgs(map[string]string{"TLS_KEY": pem})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !slices.Contains(vars, "TLS_KEY="+pem) {
			t.Fatalf("expected multiline value preserved, got %v", vars)
		}
	})

	t.Run("empty env", func(t *testing.T) {
		args, vars, err := containerEnvArgs(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(args) != 0 || len(vars) != 0 {
			t.Fatalf("expected no args or vars, got %v / %v", args, vars)
		}
	})

	t.Run("empty value is still passed", func(t *testing.T) {
		_, vars, err := containerEnvArgs(map[string]string{"EMPTY": ""})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !slices.Contains(vars, "EMPTY=") {
			t.Fatalf("expected EMPTY= in env vars, got %v", vars)
		}
	})

	t.Run("rejects bad names and values", func(t *testing.T) {
		for name, env := range map[string]map[string]string{
			"empty name":      {"": "v"},
			"equals in name":  {"A=B": "v"},
			"newline in name": {"A\nB": "v"},
			"space in name":   {"A B": "v"},
			"tab in name":     {"A\tB": "v"},
			"nul in name":     {"A\x00B": "v"},
			"docker override": {"DOCKER_HOST": "tcp://evil:2375"},
			"nul in value":    {"A": "v\x00w"},
		} {
			if _, _, err := containerEnvArgs(env); err == nil {
				t.Fatalf("%s: expected error, got none", name)
			}
		}
	})
}

func TestDockerEnvPrecedence(t *testing.T) {
	t.Setenv("NUON_TEST_PRECEDENCE", "host-value")

	env := dockerEnv("/tmp/docker-cfg", []string{
		"NUON_TEST_PRECEDENCE=container-value",
		"DOCKER_CONFIG=/tmp/hijacked",
	})

	// os/exec keeps the last duplicate, so container values must appear after
	// the host env, and DOCKER_CONFIG must appear after the container env.
	lastIndex := func(prefix string) int {
		idx := -1
		for i, e := range env {
			if strings.HasPrefix(e, prefix) {
				idx = i
			}
		}
		return idx
	}

	if got := env[lastIndex("NUON_TEST_PRECEDENCE=")]; got != "NUON_TEST_PRECEDENCE=container-value" {
		t.Fatalf("expected container value to win over host env, got %q", got)
	}
	if got := env[lastIndex("DOCKER_CONFIG=")]; got != "DOCKER_CONFIG=/tmp/docker-cfg" {
		t.Fatalf("expected runner DOCKER_CONFIG to win over container env, got %q", got)
	}
}
