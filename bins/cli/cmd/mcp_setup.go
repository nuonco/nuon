package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/mcpserver"
)

func (c *cli) mcpSetupCmd() *cobra.Command {
	var (
		platform    string
		mcpURL      string
		name        string
		allowWrites bool
	)

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Write stdio MCP client config for Claude Code, Cursor, or Amp",
		Long: `Write MCP client configuration that runs "nuon agents mcp" over stdio.

Auth stays in the CLI config (~/.nuon or -C). The client never stores the API token.

Writes in the current directory:
  claude-code, claude    .mcp.json
  cursor                 .cursor/mcp.json
  amp                    .amp/settings.json

The server is registered as "nuon" and connects to https://mcp.nuon.co/mcp.
Override either with --url and --name; both are written into the generated
command. Pass --allow-writes to include it on the generated command so the
client exposes mutating tools. A non-default -C config path is copied in as well.

  nuon agents mcp setup --platform cursor --url https://mcp.example.com/mcp --name nuon-example
  nuon agents mcp setup --platform cursor --allow-writes`,
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       outputsAnnotation(OutputTable),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if name == "" {
				name = mcpserver.NameFromAPIURL(c.cfg.APIURL)
			}

			args, err := stdioMCPArgs(cmd.Flags().Changed("url"), mcpURL, cmd.Flags().Changed("name"), name, allowWrites)
			if err != nil {
				return err
			}

			return setupProjectMCP(mcpSetupCommand(), args, platform, name)
		}),
	}

	cmd.Flags().StringVar(&platform, "platform", "", "MCP client platform (claude-code, cursor, amp)")
	cmd.Flags().StringVar(&mcpURL, "url", "", "upstream MCP server URL passed through to nuon agents mcp (default https://mcp.nuon.co/mcp)")
	cmd.Flags().StringVar(&name, "name", "", "name in the client's MCP list (default nuon, derived from the configured API URL)")
	cmd.Flags().BoolVar(&allowWrites, "allow-writes", false, "pass --allow-writes through to nuon agents mcp")
	_ = cmd.MarkFlagRequired("platform")

	return cmd
}

type mcpServerEntry struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

func stdioMCPEntry(command string, args []string) mcpServerEntry {
	return mcpServerEntry{Command: command, Args: args}
}

func mcpSetupCommand() string {
	exe, err := os.Executable()
	if err != nil {
		return "nuon"
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return exe
}

func stdioMCPArgs(urlSet bool, url string, nameSet bool, name string, allowWrites bool) ([]string, error) {
	var args []string

	configArgs, err := mcpSetupConfigFlagArgs()
	if err != nil {
		return nil, err
	}
	args = append(args, configArgs...)
	args = append(args, "agents", "mcp")
	if urlSet {
		args = append(args, "--url", url)
	}
	if nameSet {
		args = append(args, "--name", name)
	}
	if allowWrites {
		args = append(args, "--allow-writes")
	}
	return args, nil
}

func mcpSetupConfigFlagArgs() ([]string, error) {
	path := ConfigFile
	if env := os.Getenv("NUON_CONFIG_FILE"); env != "" {
		path = env
	}
	if isDefaultNuonConfig(path) {
		return nil, nil
	}

	expanded, err := homedir.Expand(path)
	if err != nil {
		return nil, fmt.Errorf("unable to expand config path: %w", err)
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve config path: %w", err)
	}
	return []string{"-C", abs}, nil
}

func isDefaultNuonConfig(path string) bool {
	if path == "" || path == DefaultConfigFilePath {
		return true
	}
	left, err := homedir.Expand(path)
	if err != nil {
		return false
	}
	right, err := homedir.Expand(DefaultConfigFilePath)
	if err != nil {
		return false
	}
	left, err = filepath.Abs(left)
	if err != nil {
		return false
	}
	right, err = filepath.Abs(right)
	if err != nil {
		return false
	}
	return left == right
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

func setupProjectMCP(command string, args []string, platform, name string) error {
	platform, err := normalizeMCPPlatform(platform)
	if err != nil {
		return err
	}

	entry := stdioMCPEntry(command, args)
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
	fmt.Printf("  Command: %s %v\n", command, args)
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

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("unable to write %s: %w", configPath, err)
	}
	return nil
}
