export const AcmeScreen = ({ path, label, children }) => (
  <div className="not-prose nuon-infra nuon-um" role="img" aria-label={label}>
    <div className="nuon-um-bar">
      <span className="nuon-um-dot" />
      <span className="nuon-um-dot" />
      <span className="nuon-um-dot" />
      <span className="nuon-um-url">app.acme.example.com/settings/{path}</span>
    </div>
    <div className="nuon-um-body">
      <div className="nuon-um-side">
        <div className="nuon-um-brand">Acme Analytics</div>
        <span>General</span>
        <span>Members</span>
        <span>Billing</span>
        <span className="nuon-um-on">Deployment</span>
      </div>
      <div className="nuon-um-main">{children}</div>
    </div>
  </div>
);

export const UiIntegrationMock = () => (
  <AcmeScreen
    path="deployment"
    label="Example settings page in a product called Acme Analytics. Under Deployment, the customer chooses where to deploy: Acme Cloud, or their own cloud account on AWS, Azure, or Google Cloud. Your own cloud account on AWS is selected, with a Connect AWS account button."
  >
    <div className="nuon-um-title">Deployment</div>
    <div className="nuon-um-sub">Choose where Acme Analytics runs for your team.</div>
    <div className="nuon-um-label">Deploy to</div>
    <div className="nuon-um-options">
      <div className="nuon-um-option">
        <span className="nuon-um-radio" />
        <span>
          <b>Acme Cloud</b>
          <small>We host and run it for you.</small>
        </span>
      </div>
      <div className="nuon-um-option nuon-um-picked">
        <span className="nuon-um-radio" />
        <span>
          <b>Your cloud account</b>
          <small>Runs in your account. Your data stays with you.</small>
        </span>
      </div>
    </div>
    <div className="nuon-um-clouds">
      <span className="nuon-um-cloud nuon-um-cloud-on">AWS</span>
      <span className="nuon-um-cloud">Azure</span>
      <span className="nuon-um-cloud">Google Cloud</span>
    </div>
    <div className="nuon-um-actions">
      <span className="nuon-um-button">Connect AWS account</span>
    </div>
  </AcmeScreen>
);

export const AcmeStartScreen = () => (
  <AcmeScreen
    path="deployment/new"
    label="Acme Analytics setup form. The customer picks the AWS region us-west-2 and enters the domain analytics.globex.example.com, then clicks Start setup."
  >
    <div className="nuon-um-title">Deploy to your AWS account</div>
    <div className="nuon-um-sub">Tell us where Acme Analytics should run.</div>
    <div className="nuon-um-field">
      <span className="nuon-um-label">AWS region</span>
      <span className="nuon-um-input">us-west-2 (Oregon)</span>
    </div>
    <div className="nuon-um-field">
      <span className="nuon-um-label">Domain</span>
      <span className="nuon-um-input">analytics.globex.example.com</span>
    </div>
    <div className="nuon-um-actions">
      <span className="nuon-um-button">Start setup</span>
    </div>
  </AcmeScreen>
);

export const AcmeTemplateScreen = () => (
  <AcmeScreen
    path="deployment/connect"
    label="Acme Analytics asks the customer to connect their AWS account. It lists what the setup template creates: a network, permissions for Acme, and an agent that installs and updates Acme Analytics. An Open in AWS console button starts it, and the page shows that it is waiting for the AWS account."
  >
    <div className="nuon-um-title">Connect your AWS account</div>
    <div className="nuon-um-sub">Run our setup template in your AWS console. It takes a few minutes.</div>
    <div className="nuon-um-box">
      <div className="nuon-um-label">The template creates</div>
      <ul className="nuon-um-list">
        <li>A private network for Acme Analytics</li>
        <li>Permissions that only cover installing and updating Acme Analytics</li>
        <li>An agent that runs those installs and updates from inside your account</li>
      </ul>
    </div>
    <div className="nuon-um-actions">
      <span className="nuon-um-button">Open in AWS console</span>
      <span className="nuon-um-status">
        <span className="nuon-um-pulse" />
        Waiting for your AWS account
      </span>
    </div>
  </AcmeScreen>
);

export const AcmeProgressScreen = () => (
  <AcmeScreen
    path="deployment"
    label="Acme Analytics setup progress. AWS account connected and cluster ready are done. Deploying Acme Analytics is in progress, two of four parts done. Ready to use is still to come."
  >
    <div className="nuon-um-title">Setting up Acme Analytics</div>
    <div className="nuon-um-sub">You can leave this page. We'll email you when it's ready.</div>
    <ol className="nuon-um-steps">
      <li className="nuon-um-done">AWS account connected</li>
      <li className="nuon-um-done">Cluster ready</li>
      <li className="nuon-um-now">
        Deploying Acme Analytics <small>2 of 4 parts</small>
      </li>
      <li>Ready to use</li>
    </ol>
  </AcmeScreen>
);

export const AcmeApproveScreen = () => (
  <AcmeScreen
    path="deployment/review"
    label="Acme Analytics asks the customer to review changes before applying them to their AWS account: create a database, create two storage buckets, update the API service. Buttons: Approve and Not now."
  >
    <div className="nuon-um-title">Review changes to your AWS account</div>
    <div className="nuon-um-sub">Nothing changes until you approve.</div>
    <div className="nuon-um-box">
      <ul className="nuon-um-diff">
        <li><span className="nuon-um-add">+</span> Create a Postgres database</li>
        <li><span className="nuon-um-add">+</span> Create 2 storage buckets</li>
        <li><span className="nuon-um-chg">~</span> Update the API service</li>
      </ul>
    </div>
    <div className="nuon-um-actions">
      <span className="nuon-um-button">Approve</span>
      <span className="nuon-um-button nuon-um-ghost">Not now</span>
    </div>
  </AcmeScreen>
);
