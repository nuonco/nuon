export const UiIntegrationMock = () => (
  <div
    className="not-prose nuon-infra nuon-um"
    role="img"
    aria-label="Example settings page in a product called Acme Analytics. Under Deployment, the customer chooses where to deploy: Acme Cloud, or their own cloud account on AWS, Azure, or Google Cloud. Your own cloud account on AWS is selected, with a Connect AWS account button."
  >
    <div className="nuon-um-bar">
      <span className="nuon-um-dot" />
      <span className="nuon-um-dot" />
      <span className="nuon-um-dot" />
      <span className="nuon-um-url">app.acme.example.com/settings/deployment</span>
    </div>
    <div className="nuon-um-body">
      <div className="nuon-um-side">
        <div className="nuon-um-brand">Acme Analytics</div>
        <span>General</span>
        <span>Members</span>
        <span>Billing</span>
        <span className="nuon-um-on">Deployment</span>
      </div>
      <div className="nuon-um-main">
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
      </div>
    </div>
  </div>
);
