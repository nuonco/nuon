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
	// AIAgentEnvVar is the general agent name. The value is a slug, optionally
	// with @version: amp, claude-code, cursor-cli@1.2.3.
	AIAgentEnvVar = "AI_AGENT"

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
	// Name is the attribution value, such as "cursor", "claude", or "amp".
	Name string
}

// Detect reports the coding agent that started this process, if any.
func Detect() (Client, bool) {
	return DetectEnv(os.Environ())
}

// DetectEnv reports the coding agent described by env. Entries are KEY=VALUE.
// NUON_AGENT_CLIENT and the Cursor and Claude Code variables win. A named
// AI_AGENT is the fallback.
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
	if name, ok := aiAgentName(vals[AIAgentEnvVar]); ok {
		return Client{Name: name}, true
	}
	return Client{}, false
}

// aiAgentName returns the slug from AI_AGENT. A @version suffix is dropped.
// Values that are not a slug, including "1" and "true", are not names.
func aiAgentName(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	name, _, _ := strings.Cut(raw, "@")
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "1", "true", "yes", "on", "false", "no", "off":
		return "", false
	}
	if !validAgentName(name) {
		return "", false
	}
	return name, true
}

func validAgentName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
		case c == '-' && i > 0 && i < len(name)-1 && name[i-1] != '-':
		default:
			return false
		}
	}
	return true
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
