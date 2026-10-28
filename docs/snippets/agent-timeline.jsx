export const AgentTimeline = ({ setup, operate }) => (
  <div className="not-prose nuon-tl">
    <div className="nuon-tl-bar">
      <div className="nuon-tl-seg nuon-tl-setup">
        <span className="nuon-tl-when">{setup.when}</span>
        <span className="nuon-tl-what">{setup.what}</span>
      </div>
      <div className="nuon-tl-seg nuon-tl-operate">
        <span className="nuon-tl-when">{operate.when}</span>
        <span className="nuon-tl-what">{operate.what}</span>
      </div>
    </div>
    <div className="nuon-tl-notes">
      <span className="nuon-tl-note-setup">{setup.note}</span>
      <span className="nuon-tl-note-operate">{operate.note}</span>
    </div>
  </div>
);
