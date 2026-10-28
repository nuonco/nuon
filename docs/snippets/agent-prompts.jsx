export const AgentPrompts = ({ items = [] }) => (
  <div className="not-prose nuon-prompts">
    {items.map((item) => (
      <a key={item.title} className="nuon-prompts-row" href={item.href}>
        <span className="nuon-prompts-title">{item.title}</span>
        <span className="nuon-prompts-quote">{item.prompt}</span>
        <span className="nuon-prompts-where">In {item.where} →</span>
      </a>
    ))}
  </div>
);
