export const FlowSteps = ({ steps = [], groups = [], loop }) => {
  const groupStartingAt = (i) => groups.find((g) => g.from === i);

  return (
    <div className="not-prose nuon-flow">
      <ol className="nuon-flow-track">
        {steps.map((step, i) => {
          const group = groupStartingAt(i);
          return (
            <li key={step.title} className="nuon-flow-step">
              {group && <span className="nuon-flow-group">{group.label}</span>}
              <span className="nuon-flow-row">
                <span className="nuon-flow-num">{i + 1}</span>
                <span className="nuon-flow-text">
                  <span className="nuon-flow-title">
                    {step.href ? <a href={step.href}>{step.title}</a> : step.title}
                  </span>
                  {step.detail && <span className="nuon-flow-detail">{step.detail}</span>}
                </span>
              </span>
            </li>
          );
        })}
      </ol>
      {loop && (
        <p className="nuon-flow-loop">
          <svg viewBox="0 0 16 16" width="14" height="14" aria-hidden="true">
            <path
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
              strokeLinecap="round"
              d="M13 8a5 5 0 1 1-1.5-3.5M13 2.5V5h-2.5"
            />
          </svg>
          {loop}
        </p>
      )}
    </div>
  );
};
