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

func TestSetupProjectMCPWritesStdioConfig(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	args := []string{"-C", "/tmp/stage.yml", "agents", "mcp"}
	require.NoError(t, setupProjectMCP("nuon", args, "cursor", "nuon-local"))
	require.NoError(t, setupProjectMCP("nuon", args, "claude", "nuon-stage"))
	require.NoError(t, setupProjectMCP("nuon", args, "amp", "nuon"))

	assertStdioServer(t, filepath.Join(dir, ".cursor", "mcp.json"), "mcpServers", "nuon-local", args)
	assertStdioServer(t, filepath.Join(dir, ".mcp.json"), "mcpServers", "nuon-stage", args)
	assertStdioServer(t, filepath.Join(dir, ".amp", "settings.json"), "amp.mcpServers", "nuon", args)
}

func TestStdioMCPArgsIncludesOverrides(t *testing.T) {
	t.Setenv("NUON_CONFIG_FILE", "")
	prev := ConfigFile
	ConfigFile = DefaultConfigFilePath
	t.Cleanup(func() { ConfigFile = prev })

	args, err := stdioMCPArgs(true, "http://localhost:8088/mcp", true, "nuon-local")
	require.NoError(t, err)
	require.Equal(t, []string{"agents", "mcp", "--url", "http://localhost:8088/mcp", "--name", "nuon-local"}, args)
}

func TestStdioMCPArgsIncludesConfigFile(t *testing.T) {
	t.Setenv("NUON_CONFIG_FILE", "")
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "stage.yml")
	require.NoError(t, os.WriteFile(cfgPath, []byte("api_url: https://api.stage.nuon.co\n"), 0o600))

	prev := ConfigFile
	ConfigFile = cfgPath
	t.Cleanup(func() { ConfigFile = prev })

	args, err := stdioMCPArgs(false, "", false, "")
	require.NoError(t, err)
	require.Equal(t, []string{"-C", cfgPath, "agents", "mcp"}, args)
}

func TestWriteMCPServersFilePreservesExistingServers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"other":true,"amp.mcpServers":{"datadog":{"url":"https://example"}}}`), 0o600))

	require.NoError(t, writeMCPServersFile(path, "amp.mcpServers", "nuon-stage", stdioMCPEntry("nuon", []string{"agents", "mcp"})))

	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	var root map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &root))
	require.JSONEq(t, `true`, string(root["other"]))

	servers := map[string]mcpServerEntry{}
	require.NoError(t, json.Unmarshal(root["amp.mcpServers"], &servers))
	require.Equal(t, "https://example", servers["datadog"].URL)
	require.Equal(t, "nuon", servers["nuon-stage"].Command)
	require.Equal(t, []string{"agents", "mcp"}, servers["nuon-stage"].Args)
}

func assertStdioServer(t *testing.T, path, key, name string, args []string) {
	t.Helper()

	raw, err := os.ReadFile(path)
	require.NoError(t, err)

	var root map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &root))

	servers := map[string]mcpServerEntry{}
	require.NoError(t, json.Unmarshal(root[key], &servers))
	require.Equal(t, "nuon", servers[name].Command)
	require.Equal(t, args, servers[name].Args)
	require.Empty(t, servers[name].URL)
	require.Empty(t, servers[name].Headers)
}
