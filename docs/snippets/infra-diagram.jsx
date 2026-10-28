export const InfraDiagram = () => (
  <div className="not-prose nuon-infra" role="img" aria-label="The Nuon control plane sits outside the customer's cloud account. Inside the account, the stack holds the runner. The runner reaches out to the control plane, then deploys the sandbox, where your components run.">
    <div className="nuon-infra-top">
      <div className="nuon-infra-top-cell">
        <div className="nuon-infra-cp">Nuon control plane</div>
        <div className="nuon-infra-link">
          <span className="nuon-infra-line" />
          <span className="nuon-infra-tag">Runner calls out. Outbound only.</span>
          <span className="nuon-infra-line" />
        </div>
      </div>
    </div>
    <div className="nuon-infra-account">
      <div className="nuon-infra-account-head">
        Customer cloud account <span>AWS, Azure, or GCP</span>
      </div>
      <div className="nuon-infra-row">
        <div className="nuon-infra-concept">
          <a className="nuon-infra-name" href="/concepts/stacks">Stack</a>
          <div className="nuon-infra-concept nuon-infra-inner">
            <a className="nuon-infra-name" href="/concepts/runners">Runner</a>
          </div>
        </div>
        <div className="nuon-infra-arrow" aria-hidden="true">
          <span>deploys</span>
          <svg viewBox="0 0 24 24" width="24" height="24">
            <path fill="none" stroke="currentColor" strokeWidth="1.75" d="M4 12h15m-5-5 5 5-5 5" />
          </svg>
        </div>
        <div className="nuon-infra-concept">
          <a className="nuon-infra-name" href="/concepts/sandboxes">Sandbox</a>
          <span className="nuon-infra-note">Kubernetes, VMs, or serverless</span>
          <a className="nuon-infra-components" href="/concepts/components">Your components</a>
        </div>
      </div>
    </div>
  </div>
);
