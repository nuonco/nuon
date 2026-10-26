package agentclient

import "strings"

// genericUserAgents are HTTP libraries and browsers. Their product token does
// not name the agent, so FromUserAgent ignores them.
var genericUserAgents = map[string]bool{
	"mozilla":         true,
	"curl":            true,
	"wget":            true,
	"node":            true,
	"node-fetch":      true,
	"undici":          true,
	"axios":           true,
	"go-http-client":  true,
	"python-requests": true,
	"python-httpx":    true,
	"python-urllib":   true,
	"aiohttp":         true,
	"okhttp":          true,
	"java":            true,
	"nuon-cli":        true,
}

// FromUserAgent returns the agent named by the first product token of a
// User-Agent header, without its version: "claude-code/2.1 (cli)" is
// "claude-code". Generic HTTP clients and browsers return false.
func FromUserAgent(ua string) (string, bool) {
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return "", false
	}
	token, _, _ := strings.Cut(ua, " ")
	name, _, _ := strings.Cut(token, "/")
	name = strings.ToLower(strings.TrimSpace(name))
	if genericUserAgents[name] || !validAgentName(name) {
		return "", false
	}
	return name, true
}

// Normalize lowercases a reported agent name so "Cursor" and "cursor" count
// as one agent.
func Normalize(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
