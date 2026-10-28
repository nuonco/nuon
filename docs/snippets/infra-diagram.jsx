export const InfraDiagram = () => (
  <div className="not-prose nuon-infra" role="img" aria-label="Nuon control plane outside the customer's cloud account. Inside the account, the stack creates the network, permissions, and runner. The runner polls the control plane for jobs, then provisions the sandbox and deploys your components into it.">
    <div className="nuon-infra-cp">
      <span className="nuon-infra-label">Nuon control plane</span>
      <span className="nuon-infra-sub">Your dashboard, CLI, and API</span>
    </div>
    <div className="nuon-infra-link">
      <span className="nuon-infra-line" />
      <span className="nuon-infra-tag">Runner polls for jobs. Outbound only.</span>
    </div>
    <div className="nuon-infra-account">
      <div className="nuon-infra-account-head">
        <span className="nuon-infra-label">Customer cloud account</span>
        <span className="nuon-infra-sub">AWS, Azure, or GCP</span>
      </div>
      <div className="nuon-infra-row">
        <div className="nuon-infra-box" data-kind="stack">
          <a className="nuon-infra-title" href="/concepts/stacks">Stack</a>
          <span className="nuon-infra-sub">Applied by the customer</span>
          <ul>
            <li>Network</li>
            <li>Permissions</li>
          </ul>
          <a className="nuon-infra-runner" href="/concepts/runners">Runner</a>
        </div>
        <div className="nuon-infra-arrow" aria-hidden="true">
          <span>provisions and deploys</span>
          <svg viewBox="0 0 24 24" width="22" height="22">
            <path fill="none" stroke="currentColor" strokeWidth="1.75" d="M4 12h15m-5-5 5 5-5 5" />
          </svg>
        </div>
        <div className="nuon-infra-box" data-kind="sandbox">
          <a className="nuon-infra-title" href="/concepts/sandboxes">Sandbox</a>
          <span className="nuon-infra-sub">Kubernetes, VMs, or serverless</span>
          <a className="nuon-infra-components" href="/concepts/components">Your components</a>
        </div>
      </div>
    </div>
  </div>
);
