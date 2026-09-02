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
		name     string
	)

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Write MCP client config for Claude Code, Cursor, or Amp",
		Long: `Write MCP client configuration so an AI agent can reach the Nuon
control-plane HTTP MCP server.

Reads the API token and org ID from the current CLI config (~/.nuon or -C).
Run "nuon auth login" and "nuon orgs select" first.

Writes in the current directory:
  claude-code, claude    .mcp.json
  cursor                 .cursor/mcp.json
  amp                    .amp/settings.json

--url overrides the MCP HTTP URL (defaults from the configured API URL).
--name overrides the client list name (nuon, nuon-stage, or nuon-local).`,
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
				mcpURL = mcpserver.EndpointFromAPIURL(c.cfg.APIURL)
			}
			if name == "" {
				name = mcpserver.NameFromAPIURL(c.cfg.APIURL)
			}

			return setupProjectMCP(mcpURL, c.cfg.APIToken, c.cfg.OrgID, platform, name)
		}),
	}

	cmd.Flags().StringVar(&platform, "platform", "", "MCP client platform (claude-code, cursor, amp)")
	cmd.Flags().StringVar(&mcpURL, "url", "", "MCP server URL (defaults based on the configured API URL)")
	cmd.Flags().StringVar(&name, "name", "", "name in the client's MCP list (defaults based on the configured API URL)")
	_ = cmd.MarkFlagRequired("platform")

	return cmd
}

type mcpServerEntry struct {
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
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

func normalizeMCPPlatform(platform string) (string, error) {
	switch platform {
	case "claude-code", "claude":
		return "claude-code", nil
	case "cursor":
		return "cursor", nil
	case "amp":
		return "amp", nil
	default:
		return "", fmt.Errorf("unsupported platform %q — supported: claude-code, cursor, amp", platform)
	}
}

func setupProjectMCP(mcpURL, apiToken, orgID, platform, name string) error {
	platform, err := normalizeMCPPlatform(platform)
	if err != nil {
		return err
	}

	entry := nuonMCPEntry(mcpURL, apiToken, orgID)
	var configPath string
	switch platform {
	case "claude-code":
		configPath = ".mcp.json"
		if err := writeMCPServersFile(configPath, "mcpServers", name, entry); err != nil {
			return err
		}
	case "cursor":
		configPath = filepath.Join(".cursor", "mcp.json")
		if err := os.MkdirAll(".cursor", 0o755); err != nil {
			return fmt.Errorf("unable to create .cursor directory: %w", err)
		}
		if err := writeMCPServersFile(configPath, "mcpServers", name, entry); err != nil {
			return err
		}
	case "amp":
		configPath = filepath.Join(".amp", "settings.json")
		if err := os.MkdirAll(".amp", 0o755); err != nil {
			return fmt.Errorf("unable to create .amp directory: %w", err)
		}
		if err := writeMCPServersFile(configPath, "amp.mcpServers", name, entry); err != nil {
			return err
		}
	}

	fmt.Printf("Configured %s MCP at %s\n", platform, configPath)
	fmt.Printf("  Name: %s\n", name)
	fmt.Printf("  URL: %s\n", mcpURL)
	fmt.Printf("  Org: %s\n", orgID)
	return nil
}

func writeMCPServersFile(configPath, serversKey, name string, entry mcpServerEntry) error {
	root := map[string]json.RawMessage{}
	if data, err := os.ReadFile(configPath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &root); err != nil {
			return fmt.Errorf("unable to parse %s: %w", configPath, err)
		}
	}

	servers := map[string]mcpServerEntry{}
	if raw, ok := root[serversKey]; ok {
		if err := json.Unmarshal(raw, &servers); err != nil {
			return fmt.Errorf("unable to parse %s %s: %w", configPath, serversKey, err)
		}
	}
	if servers == nil {
		servers = map[string]mcpServerEntry{}
	}
	servers[name] = entry

	encoded, err := json.Marshal(servers)
	if err != nil {
		return fmt.Errorf("unable to marshal MCP servers: %w", err)
	}
	root[serversKey] = encoded

	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("unable to write %s: %w", configPath, err)
	}
	return nil
}
