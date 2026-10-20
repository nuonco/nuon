export const AskYourAgent = ({ title, prompt, tools = [] }) => {
  const [open, setOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  const copy = () => {
    navigator.clipboard.writeText(prompt).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  return (
    <div className="not-prose nuon-ask-agent" data-open={open}>
      <div className="nuon-ask-agent-row">
        <button
          type="button"
          className="nuon-ask-agent-toggle"
          aria-expanded={open}
          onClick={() => setOpen(!open)}
        >
          <span className="nuon-ask-agent-badge">
            <svg viewBox="0 0 16 16" width="12" height="12" aria-hidden="true">
              <path
                fill="currentColor"
                d="M8 0l1.8 5.2L15 7l-5.2 1.8L8 14l-1.8-5.2L1 7l5.2-1.8z"
              />
            </svg>
            Ask your agent
          </span>
          <span className="nuon-ask-agent-title">{title}</span>
          <svg
            className="nuon-ask-agent-chevron"
            viewBox="0 0 16 16"
            width="12"
            height="12"
            aria-hidden="true"
          >
            <path fill="none" stroke="currentColor" strokeWidth="1.5" d="M4 6l4 4 4-4" />
          </svg>
        </button>
        <div className="nuon-ask-agent-actions">
          <button type="button" className="nuon-ask-agent-btn nuon-ask-agent-primary" onClick={copy}>
            {copied ? "Copied" : "Copy prompt"}
          </button>
        </div>
      </div>
      {open && (
        <div className="nuon-ask-agent-body">
          <pre className="nuon-ask-agent-prompt">{prompt}</pre>
          <p className="nuon-ask-agent-note">
            Read-only: the agent calls{" "}
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
