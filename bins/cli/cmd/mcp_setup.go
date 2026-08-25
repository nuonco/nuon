package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/mcpserver"
)

func (c *cli) mcpSetupCmd() *cobra.Command {
	var (
		platform string
		mcpURL   string
	)

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure an MCP client to connect to the Nuon control plane",
		Long: `Write MCP client configuration for your platform so your AI agent can
interact with the Nuon control plane over HTTP.

Reads the API token and org ID from your current nuon auth context (~/.nuon).
Run "nuon auth login" first if you haven't authenticated.

Supported platforms:
  claude-code    Claude Code (~/.claude.json)
  cursor         Cursor (.cursor/mcp.json in current directory)`,
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       outputsAnnotation(OutputTable),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if c.cfg.APIToken == "" {
				return fmt.Errorf("no API token found — run \"nuon auth login\" first")
			}
			if c.cfg.OrgID == "" {
				return fmt.Errorf("no org selected — run \"nuon orgs select\" first")
			}

			if mcpURL == "" {
				mcpURL = defaultMCPURL(c.cfg.APIURL)
			}

			switch platform {
			case "claude-code", "cursor":
				return setupProjectMCP(mcpURL, c.cfg.APIToken, c.cfg.OrgID, platform)
			default:
				return fmt.Errorf("unsupported platform %q — supported: claude-code, cursor", platform)
			}
		}),
	}

	cmd.Flags().StringVar(&platform, "platform", "", "MCP client platform (claude-code, cursor)")
	cmd.Flags().StringVar(&mcpURL, "url", "", "MCP server URL (defaults based on environment)")
	_ = cmd.MarkFlagRequired("platform")

	return cmd
}

func defaultMCPURL(apiURL string) string {
	return mcpserver.EndpointFromAPIURL(apiURL)
}

type mcpClientConfig struct {
	MCPServers map[string]mcpServerEntry `json:"mcpServers"`
}

type mcpServerEntry struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

func nuonMCPEntry(mcpURL, apiToken, orgID string) mcpServerEntry {
	return mcpServerEntry{
		URL: mcpURL,
		Headers: map[string]string{
			"Authorization": "Bearer " + apiToken,
			"X-Nuon-Org-ID": orgID,
		},
	}
}

func setupProjectMCP(mcpURL, apiToken, orgID, platform string) error {
	var configPath string
	switch platform {
	case "claude-code":
		configPath = ".mcp.json"
	case "cursor":
		configPath = filepath.Join(".cursor", "mcp.json")
		if err := os.MkdirAll(".cursor", 0755); err != nil {
			return fmt.Errorf("unable to create .cursor directory: %w", err)
		}
	}

	cfg := mcpClientConfig{MCPServers: map[string]mcpServerEntry{}}

	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}
	if cfg.MCPServers == nil {
		cfg.MCPServers = map[string]mcpServerEntry{}
	}

	cfg.MCPServers["nuon-api"] = nuonMCPEntry(mcpURL, apiToken, orgID)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("unable to write %s: %w", configPath, err)
	}

	fmt.Printf("Configured %s MCP at %s\n", platform, configPath)
	fmt.Printf("  URL: %s\n", mcpURL)
	fmt.Printf("  Org: %s\n", orgID)
	return nil
}
