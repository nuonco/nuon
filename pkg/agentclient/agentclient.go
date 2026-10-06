// Package agentclient identifies a coding agent that launched the CLI.
// The detected name is safe to print and to send as the X-Nuon-Agent header.
package agentclient

import (
	"os"
	"strings"
)

const (
	// Header is the request header carrying the detected agent name.
	Header = "X-Nuon-Agent"
	// CommandHeader is the request header carrying the CLI command path, without arguments.
	CommandHeader = "X-Nuon-Command"
	// EnvVar overrides detection. "off" disables it. "cursor" and "claude" force a name.
	EnvVar = "NUON_AGENT_CLIENT"

	Cursor = "cursor"
	Claude = "claude"
)

// UserAgent is the HTTP User-Agent for an agent-driven CLI request.
// It contains "nuon-cli" so control-plane CLI detection still matches.
func UserAgent(version, agent string) string {
	return "nuon-cli/" + version + " (" + agent + ")"
}

// Client is a detected coding agent.
type Client struct {
	// Name is the attribution value: "cursor" or "claude".
	Name string
}

// Detect reports the coding agent that started this process, if any.
func Detect() (Client, bool) {
	return DetectEnv(os.Environ())
}

// DetectEnv reports the coding agent described by env. Entries are KEY=VALUE.
// An explicit NUON_AGENT_CLIENT wins over ambient agent variables.
func DetectEnv(env []string) (Client, bool) {
	vals := envValues(env)
	switch strings.ToLower(strings.TrimSpace(vals[EnvVar])) {
	case "off":
		return Client{}, false
	case Cursor:
		return Client{Name: Cursor}, true
	case Claude:
		return Client{Name: Claude}, true
	}

	if vals["CURSOR_AGENT"] == "1" || vals["CURSOR_INVOKED_AS"] == "agent" {
		return Client{Name: Cursor}, true
	}
	if vals["CLAUDECODE"] == "1" {
		return Client{Name: Claude}, true
	}
	return Client{}, false
}

func envValues(env []string) map[string]string {
	vals := make(map[string]string, len(env))
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		vals[key] = value
	}
	return vals
}
