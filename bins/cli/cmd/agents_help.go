package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/mcpserver"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

// mcpClientJSON renders a client config block. Keys differ by client
// (mcpServers, amp.mcpServers, and others); pass the key the client uses.
func mcpClientJSON(key, indent string) string {
	block := `{
  "` + key + `": {
    "nuon": {
      "command": "nuon",
      "args": ["agents", "mcp", "--allow-writes"]
    }
  }
}`

	lines := strings.Split(block, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}

	return strings.Join(lines, "\n")
}

// agentsStatus is the live CLI state woven into the setup guide. It is nil
// when the guide backs static help text, which is built before flags have
// selected a config file.
type agentsStatus struct {
	APIURL    string
	SignedIn  bool
	OrgID     string
	MCPURL    string
	MCPURLErr error
}

func (c *cli) agentsStatus() *agentsStatus {
	st := &agentsStatus{APIURL: "https://api.nuon.co"}
	if c.cfg != nil {
		if c.cfg.APIURL != "" {
			st.APIURL = c.cfg.APIURL
		}
		st.SignedIn = c.cfg.APIToken != ""
		st.OrgID = c.cfg.OrgID
	}
	st.MCPURL, st.MCPURLErr = mcpserver.EndpointFromAPIURL(st.APIURL)

	return st
}

// agentsSetupGuide is the single source for agent setup instructions: it backs
// both "nuon agents help" and the "nuon agents" help text, so the two never
// disagree about a client's config file or a flag.
func agentsSetupGuide(st *agentsStatus) string {
	var b strings.Builder
	p := func(lines ...string) {
		for _, line := range lines {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	p(
		"Drive Nuon with an LLM agent.",
		"",
		"  nuon agents help      For humans: get set up. This guide.",
		"  nuon agents context   For your agent: orientation, tool catalog, sample",
		"                        queries, current org/app/install selection.",
		"  nuon agents mcp       The stdio MCP proxy a client runs. Registered below.",
		"",
		"1. Sign in and select an org",
		"",
	)

	if st != nil {
		if st.SignedIn {
			p("  " + styles.TextSuccess.Render("✓") + " Signed in to " + st.APIURL)
		} else {
			p("  " + styles.TextError.Render("✗") + " Not signed in")
		}
		if st.OrgID != "" {
			p("  " + styles.TextSuccess.Render("✓") + " Org " + st.OrgID)
		} else {
			p("  " + styles.TextError.Render("✗") + " No org selected")
		}
		p("")
	}

	p(
		"  Both are required:",
		"",
		"    nuon auth login    writes api_token to ~/.nuon",
		"    nuon orgs select   writes org_id to ~/.nuon",
		"",
		"  Everything below reads both from ~/.nuon, so no token or org ID goes",
		"  into your agent's config.",
		"",
		"2. Register the MCP server with your agent",
		"",
		"  Each client has its own config file, JSON key, and (sometimes) an add",
		"  command. Check that client's MCP docs rather than assuming a shared",
		"  schema. The process to spawn is nuon with arguments agents, mcp, and",
		"  --allow-writes.",
		"",
		"  --allow-writes exposes the mutating tools (descriptions start with \"WRITE",
		"  OPERATION:\"). That flag only lists them; the identity still needs org",
		"  Admin (org_admin), not Read-only. See",
		"  https://docs.nuon.co/concepts/access-control. Leave the flag off for a",
		"  read-only proxy.",
		"",
		"  Claude Code",
		"",
		"    claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes",
		"",
		"    Or as .mcp.json in a project:",
		"",
		mcpClientJSON("mcpServers", "      "),
		"",
		"  Cursor",
		"",
		"    Cursor has no add command. Save this as ~/.cursor/mcp.json (all",
		"    projects) or .cursor/mcp.json (one project):",
		"",
		mcpClientJSON("mcpServers", "      "),
		"",
		"    Then enable it:",
		"",
		"      agent mcp enable nuon",
		"",
		"  Amp",
		"",
		"    amp mcp add nuon -- nuon agents mcp --allow-writes",
		"",
		"    Add --workspace to that command to scope it to one workspace. By file,",
		"    Amp's key is amp.mcpServers, in ~/.config/amp/settings.json (you) or",
		"    .amp/settings.json (a workspace):",
		"",
		mcpClientJSON("amp.mcpServers", "      "),
		"",
		"  Other clients",
		"",
		"    Use that client's MCP docs for the file path and JSON key. Point it at",
		"    the nuon binary with arguments agents, mcp, and --allow-writes.",
		"",
		"3. Check it works",
		"",
		"    nuon agents context",
		"",
		"  Then ask your agent to run the whoami tool.",
		"",
		"Overriding the MCP URL",
		"",
		"  The proxy derives its MCP URL from api_url in ~/.nuon, turning api.<host>",
		"  into mcp.<host>/mcp. Pass --url when the MCP URL does not follow from the",
		"  API URL, for example a self-hosted or Nuon BYOC control plane, or any",
		"  deployment where the two hostnames differ. Pass --name to rename the",
		"  server in your client's list. Both go on the command the client runs:",
		"",
		"    claude mcp add --transport stdio nuon -- nuon agents mcp \\",
		"      --allow-writes --url https://mcp.example.nuon.co/mcp --name nuon-example",
		"",
	)

	if st != nil {
		if st.MCPURLErr != nil {
			p(
				"  "+styles.TextError.Render("Your api_url ("+st.APIURL+") does not follow that"),
				"  "+styles.TextError.Render("pattern, so --url is required."),
			)
		} else {
			p("  Yours resolves to " + st.MCPURL + ".")
		}
		p("")
	}

	p(
		"  A non-default CLI config carries its own token, org, and api_url: put -C",
		"  on the registered command too (Cursor and Amp: inside args).",
		"",
		"Docs: https://docs.nuon.co/guides/agents",
	)

	return b.String()
}

func (c *cli) agentsHelpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "Set up an agent to work with Nuon (for humans)",
		Long: `Print the agent setup guide: signing in, selecting an org, registering the
MCP server with Claude Code, Cursor, or Amp, and overriding the derived MCP URL.

Same guide as "nuon agents --help", plus your own sign-in, org, and resolved
MCP URL. For the document to hand your agent, use "nuon agents context".`,
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       annotations(skipAuthAnnotation(), outputsAnnotation(OutputTable)),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			// cobra's cmd.Print* writes to stderr; this guide is the output.
			fmt.Fprintln(cmd.OutOrStdout(), agentsSetupGuide(c.agentsStatus()))
			return nil
		}),
	}
}
