export const McpInstallPicker = () => {
  const json = (key) =>
    `{
  "${key}": {
    "nuon": {
      "command": "nuon",
      "args": ["agents", "mcp", "--allow-writes"]
    }
  }
}`;

  const clients = [
    {
      id: "claude-code",
      label: "Claude Code",
      blocks: [
        {
          code: `claude mcp add --transport stdio nuon -- nuon agents mcp --allow-writes`,
        },
      ],
    },
    {
      id: "codex",
      label: "Codex",
      blocks: [
        {
          code: `codex mcp add nuon -- nuon agents mcp --allow-writes`,
        },
      ],
    },
    {
      id: "cursor",
      label: "Cursor",
      blocks: [
        { caption: "Save as ~/.cursor/mcp.json", code: json("mcpServers") },
        { caption: "Then enable it", code: "agent mcp enable nuon" },
      ],
    },
    {
      id: "amp",
      label: "Amp",
      blocks: [
        {
          code: `amp mcp add nuon -- nuon agents mcp --allow-writes`,
        },
      ],
    },
    {
      id: "other",
      label: "Other",
      blocks: [
        {
          caption: "Add to your client's MCP config. The file path and JSON key vary by client.",
          code: json("mcpServers"),
        },
      ],
    },
  ];

  const [active, setActive] = useState(clients[0].id);
  const [copied, setCopied] = useState(-1);
  const client = clients.find((c) => c.id === active);

  const copy = (text, i) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(i);
      setTimeout(() => setCopied(-1), 2000);
    });
  };

  return (
    <div className="not-prose nuon-mcp-picker">
      <div className="nuon-mcp-picker-tabs" role="tablist" aria-label="MCP client">
        {clients.map((c) => (
          <button
            key={c.id}
            type="button"
            role="tab"
            aria-selected={c.id === active}
            className="nuon-mcp-picker-tab"
            onClick={() => {
              setActive(c.id);
              setCopied(-1);
            }}
          >
            {c.label}
          </button>
        ))}
      </div>
      {client.blocks.map((b, i) => (
        <div key={`${active}-${i}`} className="nuon-mcp-picker-block">
          {b.caption && <p className="nuon-mcp-picker-caption">{b.caption}</p>}
          <div className="nuon-mcp-picker-code">
            <pre>{b.code}</pre>
            <button
              type="button"
              className="nuon-try-agent-btn nuon-try-agent-primary nuon-mcp-picker-copy"
              onClick={() => copy(b.code, i)}
            >
              {copied === i ? "Copied" : "Copy"}
            </button>
          </div>
        </div>
      ))}
      <p className="nuon-mcp-picker-foot">
        <a href="/guides/agents/setup">Other clients, project config and custom MCP URLs →</a>
      </p>
    </div>
  );
};
