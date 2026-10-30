export const TryWithYourAgent = ({ title, prompt, tools = [], writes = false }) => {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  const copy = () => {
    navigator.clipboard.writeText(prompt).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <div className="not-prose nuon-try-agent" data-open={open}>
      <div className="nuon-try-agent-row">
        <button
          type="button"
          className="nuon-try-agent-toggle"
          aria-expanded={open}
          onClick={() => setOpen(!open)}
        >
          <span className="nuon-try-agent-badge">
            <svg viewBox="0 0 16 16" width="12" height="12" aria-hidden="true">
              <path
                fill="currentColor"
                d="M8 0l1.8 5.2L15 7l-5.2 1.8L8 14l-1.8-5.2L1 7l5.2-1.8z"
              />
            </svg>
            Try with your agent
          </span>
          <span className="nuon-try-agent-title">{title}</span>
          <svg
            className="nuon-try-agent-chevron"
            viewBox="0 0 16 16"
            width="12"
            height="12"
            aria-hidden="true"
          >
            <path fill="none" stroke="currentColor" strokeWidth="1.5" d="M4 6l4 4 4-4" />
          </svg>
        </button>
        <div className="nuon-try-agent-actions">
          <button type="button" className="nuon-try-agent-btn nuon-try-agent-primary" onClick={copy}>
            {copied ? "Copied" : "Copy prompt"}
          </button>
        </div>
      </div>
      {open && (
        <div className="nuon-try-agent-body">
          <pre className="nuon-try-agent-prompt">{prompt}</pre>
          <p className="nuon-try-agent-note">
            {writes ? "Needs write access: the agent calls" : "Read-only: the agent calls"}{" "}
            {tools.map((t, i) => (
              <span key={t}>
                {i > 0 && (i === tools.length - 1 ? " and " : ", ")}
                <code>{t}</code>
              </span>
            ))}
            .
            <a href="/guides/agents/setup">Set up the Nuon MCP server →</a>
          </p>
        </div>
      )}
    </div>
  );
};
