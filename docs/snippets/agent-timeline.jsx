export const AgentTimeline = ({ segments = [] }) => {
  const cols = segments.map((s) => `${s.weight || 1}fr`).join(" ");
  return (
    <div className="not-prose nuon-tl">
      <div className="nuon-tl-bar" style={{ gridTemplateColumns: cols }}>
        {segments.map((s) => (
          <div key={s.when} className="nuon-tl-seg" data-tone={s.tone}>
            <span className="nuon-tl-when">{s.when}</span>
            <span className="nuon-tl-what">{s.what}</span>
          </div>
        ))}
      </div>
      <div className="nuon-tl-notes" style={{ gridTemplateColumns: cols }}>
        {segments.map((s) => (
          <span key={s.when}>{s.note}</span>
        ))}
      </div>
    </div>
  );
};
