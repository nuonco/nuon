export const AppBranchesDiagram = () => {
  const Arrow = () => (
    <svg className="nuon-ab-arrow" viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
      <path fill="none" stroke="currentColor" strokeWidth="2" d="M4 12h15m-5-5 5 5-5 5" />
    </svg>
  );
  const Branch = ({ name, groups }) => (
    <div className="nuon-ab-branch">
      <div className="nuon-ab-head">
        <a className="nuon-ab-kicker" href="/concepts/app-branches">App branch</a>
        <span className="nuon-ab-branch-name">{name}</span>
      </div>
      <div className="nuon-ab-head">
        <a className="nuon-ab-kicker" href="/concepts/app-branches#install-groups">Install groups</a>
        <div className="nuon-ab-groups">
          {groups.map((g, i) => (
            <span key={g} className="nuon-ab-group-wrap">
              {i > 0 && <Arrow />}
              <span className="nuon-ab-group">{g}</span>
            </span>
          ))}
        </div>
      </div>
    </div>
  );
  const App = ({ name, children }) => (
    <div className="nuon-infra-concept">
      <div className="nuon-ab-head">
        <a className="nuon-ab-kicker" href="/concepts/apps">App</a>
        <span className="nuon-ab-app">{name}</span>
      </div>
      {children}
    </div>
  );
  return (
    <div className="not-prose nuon-infra nuon-ab" role="img" aria-label="Each app, such as acme-aws or acme-azure, has one or more app branches. Each app branch splits its installs into install groups that deploy in order, such as canary then customers.">
      <div className="nuon-ab-row">
        <App name="acme-aws">
          <Branch name="main" groups={["canary", "customers"]} />
          <Branch name="big-customer" groups={["production"]} />
        </App>
        <App name="acme-azure">
          <Branch name="main" groups={["canary", "customers"]} />
        </App>
      </div>
    </div>
  );
};
