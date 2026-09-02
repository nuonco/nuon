package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeMCPPlatform(t *testing.T) {
	got, err := normalizeMCPPlatform("claude")
	require.NoError(t, err)
	require.Equal(t, "claude-code", got)

	_, err = normalizeMCPPlatform("windsurf")
	require.Error(t, err)
}

func TestSetupProjectMCPWritesPlatformsAndOverrides(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	require.NoError(t, setupProjectMCP("http://localhost:8088/mcp", "tok_test", "org_test", "cursor", "nuon-local"))
	require.NoError(t, setupProjectMCP("https://mcp.stage.nuon.co/mcp", "tok_test", "org_test", "claude", "nuon-stage"))
	require.NoError(t, setupProjectMCP("https://mcp.nuon.co/mcp", "tok_test", "org_test", "amp", "nuon"))

	assertHTTPServer(t, filepath.Join(dir, ".cursor", "mcp.json"), "mcpServers", "nuon-local", "http://localhost:8088/mcp")
	assertHTTPServer(t, filepath.Join(dir, ".mcp.json"), "mcpServers", "nuon-stage", "https://mcp.stage.nuon.co/mcp")
	assertHTTPServer(t, filepath.Join(dir, ".amp", "settings.json"), "amp.mcpServers", "nuon", "https://mcp.nuon.co/mcp")
}

func TestWriteMCPServersFilePreservesExistingServers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"other":true,"amp.mcpServers":{"datadog":{"url":"https://example"}}}`), 0o600))

	require.NoError(t, writeMCPServersFile(path, "amp.mcpServers", "nuon-stage", nuonMCPEntry("https://mcp.stage.nuon.co/mcp", "tok", "org")))

	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	var root map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &root))
	require.JSONEq(t, `true`, string(root["other"]))

	servers := map[string]mcpServerEntry{}
	require.NoError(t, json.Unmarshal(root["amp.mcpServers"], &servers))
	require.Equal(t, "https://example", servers["datadog"].URL)
	require.Equal(t, "https://mcp.stage.nuon.co/mcp", servers["nuon-stage"].URL)
	require.Equal(t, "Bearer tok", servers["nuon-stage"].Headers["Authorization"])
}

func assertHTTPServer(t *testing.T, path, key, name, url string) {
	t.Helper()

	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	var root map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &root))

	servers := map[string]mcpServerEntry{}
	require.NoError(t, json.Unmarshal(root[key], &servers))
	require.Equal(t, url, servers[name].URL)
	require.Equal(t, "org_test", servers[name].Headers["X-Nuon-Org-ID"])
}
