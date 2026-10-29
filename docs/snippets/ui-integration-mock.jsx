export const AcmeScreen = ({ screen }) => {
  const [cloud, setCloud] = useState("aws");
  const clouds = {
    aws: {
      tab: "AWS",
      account: "AWS account",
      where: "in your AWS console",
      button: "Open in AWS console",
      create: <><code>aws_account.region</code></>,
      link: <>The <code>quick_link_url</code> on the install's stack opens CloudFormation quick-create, ready to run.</>,
      guide: { href: "/platform-support/aws", label: "AWS platform guide" },
    },
    azure: {
      tab: "Azure",
      account: "Azure subscription",
      where: "in the Azure portal",
      button: "Deploy to Azure",
      create: <><code>azure_account.location</code></>,
      link: <>The <code>quick_link_url</code> opens Deploy to Azure. It's empty when the install is scoped to a resource group, so the customer deploys the Bicep template instead.</>,
      guide: { href: "/platform-support/azure", label: "Azure platform guide" },
    },
    gcp: {
      tab: "Google Cloud",
      account: "Google Cloud project",
      where: "with Terraform in your project",
      button: "View setup steps",
      create: <><code>gcp_account.project_id</code> and <code>gcp_account.region</code></>,
      link: <>There's no one-click link. The customer applies the generated Terraform, for example with Google Cloud's Infrastructure Manager.</>,
      guide: { href: "/platform-support/gcp", label: "Google Cloud platform guide" },
    },
  };
  const c = clouds[cloud];
  const Frame = ({ path, label, children }) => (
    <div className="not-prose nuon-infra nuon-um" role="img" aria-label={label}>
      <div className="nuon-um-bar">
        <span className="nuon-um-dot" />
        <span className="nuon-um-dot" />
        <span className="nuon-um-dot" />
        <span className="nuon-um-url">app.acme.example.com/settings/{path}</span>
        <span className="nuon-um-tag">Example UI</span>
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
  const screens = {
    deploy: (
      <Frame
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
      </Frame>
    ),
    start: (
      <Frame
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
      </Frame>
    ),
    connect: (
      <div className="nuon-um-picker">
        <Frame
          path="deployment/connect"
          label={`Acme Analytics asks the customer to connect their ${c.account}. It lists what the setup template creates: a network, permissions for Acme, and an agent that installs and updates Acme Analytics. A ${c.button} button starts it.`}
        >
          <div className="nuon-um-title">Connect your {c.account}</div>
          <div className="nuon-um-sub">Run our setup template {c.where}. It takes a few minutes.</div>
          <div className="nuon-um-box">
            <div className="nuon-um-label">The template creates</div>
            <ul className="nuon-um-list">
              <li>A private network for Acme Analytics</li>
              <li>Permissions that only cover installing and updating Acme Analytics</li>
              <li>An agent that runs those installs and updates from inside your account</li>
            </ul>
          </div>
          <div className="nuon-um-actions">
            <span className="nuon-um-button">{c.button}</span>
            <span className="nuon-um-status">
              <span className="nuon-um-pulse" />
              Waiting for your {c.account}
            </span>
          </div>
        </Frame>
        <div className="nuon-um-tabs" role="tablist" aria-label="Customer's cloud">
          {Object.keys(clouds).map((k) => (
            <button
              key={k}
              type="button"
              role="tab"
              aria-selected={cloud === k}
              className={cloud === k ? "nuon-um-tab nuon-um-tab-on" : "nuon-um-tab"}
              onClick={() => setCloud(k)}
            >
              {clouds[k].tab}
            </button>
          ))}
        </div>
        <div className="not-prose nuon-infra nuon-um-notes">
          <div>
            <span className="nuon-um-notes-k">Create the install with</span>
            <span>{c.create}</span>
          </div>
          <div>
            <span className="nuon-um-notes-k">Setup link</span>
            <span>{c.link}</span>
          </div>
          <div>
            <span className="nuon-um-notes-k">More</span>
            <span><a href={c.guide.href}>{c.guide.label}</a></span>
          </div>
        </div>
      </div>
    ),
    progress: (
      <Frame
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
      </Frame>
    ),
    approve: (
      <Frame
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
      </Frame>
    ),
  };
  return screens[screen] || null;
};
