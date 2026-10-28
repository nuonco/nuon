export const ConceptMap = ({ columns = [], arrows = false }) => (
  <div className="not-prose nuon-cmap" data-cols={columns.length} data-arrows={arrows}>
    {columns.map((col, i) => (
      <div key={col.title} className="nuon-cmap-cell">
        {arrows && i > 0 && (
          <div className="nuon-cmap-arrow" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="20" height="20">
              <path fill="none" stroke="currentColor" strokeWidth="1.75" d="M4 12h15m-5-5 5 5-5 5" />
            </svg>
          </div>
        )}
        <div className="nuon-cmap-col" data-tone={col.tone}>
          <div className="nuon-cmap-title">{col.title}</div>
          {col.caption && <div className="nuon-cmap-caption">{col.caption}</div>}
          <ul className="nuon-cmap-list">
            {col.items.map((item) => (
              <li key={item.label}>
                <a href={item.href}>{item.label}</a>
                {item.note && <span className="nuon-cmap-note">{item.note}</span>}
              </li>
            ))}
          </ul>
        </div>
      </div>
    ))}
  </div>
);
