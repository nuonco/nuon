export const PlatformDiagram = () => {
  const Arrow = ({ d }) => (
    <svg viewBox="0 0 28 24" width="28" height="24" aria-hidden="true">
      <path fill="none" stroke="currentColor" strokeWidth="1.75" strokeDasharray={d} d="M2 12h22m-5-5 5 5-5 5" />
    </svg>
  );
  const Mark = () => (
    <path d="M24.272 0L15.525 5.05134V13.2295L8.44422 9.13842H8.44018L-0.000244141 14.0122V40.8304L8.43614 45.7042H8.44018L17.4979 40.4713V32.6521L24.272 36.5618L33.0191 31.5105V5.05134L24.272 0ZM1.97672 15.154L8.43614 11.4261H8.44018L15.5209 15.5131V29.2269L1.97672 21.4077V15.154ZM15.5209 39.3295L8.43614 43.4166L1.97672 39.6886V23.6913L15.5209 31.5105V39.3295ZM31.0422 30.3645L24.272 34.2742L17.5019 30.3687V16.6549L31.0461 24.474V30.3645H31.0422ZM31.0422 22.1904L17.4979 14.3713V6.19314L24.272 2.28359L31.0422 6.19314V22.1904Z" />
  );
  const Wordmark = () => (
    <svg className="nuon-pd-logo" viewBox="0 0 110 46" fill="currentColor" role="img" aria-label="Nuon">
      <Mark />
      <path d="M58.3241 27.3831V36.582H54.3539V28.3192C54.3539 25.8983 53.2243 24.8654 51.2878 24.8654C49.2542 24.8654 47.8019 26.0598 47.8019 29.2875V36.582H43.864V21.6701H47.7696V23.9295H47.834C48.8347 22.2187 50.513 21.186 52.708 21.186C55.9356 21.186 58.3241 22.9935 58.3241 27.3831Z" />
      <path d="M60.8943 30.8038V21.6694H64.9612V30.0938C64.9612 32.4176 66.0587 33.4182 67.7692 33.4182C69.7383 33.4182 71.0615 32.1272 71.0615 28.8349V21.6694H75.0317V36.5814H71.1261V34.2897H71.0615C70.1579 35.9035 68.7375 37.0656 66.4137 37.0656C63.3797 37.0656 60.8943 35.2259 60.8943 30.8038Z" />
      <path d="M93.5588 29.1262C93.5588 33.8063 90.0405 37.0663 85.2958 37.0663C80.5512 37.0663 77.033 33.8063 77.033 29.1262C77.033 24.446 80.5512 21.186 85.2958 21.186C90.0405 21.186 93.5588 24.446 93.5588 29.1262ZM81.0998 29.1262C81.0998 31.6759 82.875 33.3221 85.2958 33.3221C87.7165 33.3221 89.4596 31.6759 89.4596 29.1262C89.4596 26.5762 87.7165 24.93 85.2958 24.93C82.875 24.93 81.0998 26.5762 81.0998 29.1262Z" />
      <path d="M110 27.3831V36.582H106.03V28.3192C106.03 25.8983 104.9 24.8654 102.963 24.8654C100.93 24.8654 99.4775 26.0598 99.4775 29.2875V36.582H95.5398V21.6701H99.4453V23.9295H99.5098C100.51 22.2187 102.189 21.186 104.384 21.186C107.611 21.186 110 22.9935 110 27.3831Z" />
    </svg>
  );
  const rowRef = useRef(null);
  const cpRef = useRef(null);
  const runnerRef = useRef(null);
  const midRef = useRef(null);
  const [link, setLink] = useState(null);
  useEffect(() => {
    const row = rowRef.current;
    if (!row) return;
    const update = () => {
      const r = row.getBoundingClientRect();
      const c = cpRef.current.getBoundingClientRect();
      const a = runnerRef.current.getBoundingClientRect();
      const m = midRef.current.getBoundingClientRect();
      if (m.width === 0 || m.left < c.right) {
        setLink(null);
        return;
      }
      const sx = a.left - r.left;
      const sy = a.top + a.height / 2 - r.top;
      const vx = m.left - r.left + 10;
      const ex = c.right - r.left + 8;
      const ey = c.top + c.height / 2 - r.top;
      setLink({ sx, sy, vx, ex, ey, w: r.width, h: r.height });
    };
    update();
    const ro = new ResizeObserver(update);
    ro.observe(row);
    return () => ro.disconnect();
  }, []);
  return (
    <div
      className="not-prose nuon-infra nuon-pd"
      role="img"
      aria-label="Your app template and the Nuon control plane sit on the Nuon side. A git push rolls changes out to installs in each customer's cloud account. Each install's runner calls out to the control plane, outbound only."
    >
      <div className="nuon-pd-row" ref={rowRef}>
        <div className="nuon-pd-zone nuon-pd-yours">
          <div className="nuon-pd-zlabel nuon-pd-cp">
            <Wordmark />
            <span ref={cpRef}>control plane</span>
          </div>
          <div className="nuon-infra-concept nuon-pd-template">
            <div className="nuon-ab-head">
              <a className="nuon-ab-kicker" href="/concepts/apps">App template</a>
              <span className="nuon-pd-where">your git repo</span>
            </div>
            <div className="nuon-pd-template-name">acme</div>
            <div className="nuon-pd-template-sub">Your product, described as TOML</div>
            <div className="nuon-infra-components">
              <a className="nuon-infra-comps-title" href="/concepts/components">Components</a>
              <div className="nuon-infra-chips">
                <span className="nuon-infra-chip">API</span>
                <span className="nuon-infra-chip">Database</span>
                <span className="nuon-infra-chip">Storage bucket</span>
              </div>
            </div>
            <div className="nuon-infra-components">
              <span className="nuon-infra-comps-title">Per customer</span>
              <div className="nuon-infra-chips">
                <a className="nuon-infra-chip" href="/concepts/app-inputs">Inputs</a>
                <a className="nuon-infra-chip" href="/concepts/app-secrets">Secrets</a>
              </div>
            </div>
          </div>
        </div>
        <div className="nuon-pd-flows" ref={midRef} aria-hidden="true">
          <div className="nuon-pd-flow">
            <span>git push<br />rolls out</span>
            <Arrow />
          </div>
          <div className="nuon-pd-flow nuon-pd-flow-back">
            <Arrow d="3 3" />
            <span>runner calls out<br />outbound only</span>
          </div>
        </div>
        {link && (
          <>
            <svg className="nuon-pd-link" width={link.w} height={link.h} aria-hidden="true">
              <path
                d={`M${link.sx} ${link.sy}H${link.vx}V${link.ey}H${link.ex}`}
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeDasharray="4 3"
              />
              <path
                d={`M${link.ex + 7} ${link.ey - 4.5}L${link.ex} ${link.ey}L${link.ex + 7} ${link.ey + 4.5}`}
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
              />
            </svg>
            <span className="nuon-pd-link-label" style={{ right: link.w - link.vx + 38, top: link.ey }}>
              runner calls out, outbound only
            </span>
          </>
        )}
        <div className="nuon-pd-zone nuon-pd-theirs">
          <div className="nuon-pd-zlabel">
            Each customer's cloud account <span>AWS, Azure, or GCP</span>
          </div>
          <div className="nuon-infra-concept">
            <div className="nuon-ab-head">
              <a className="nuon-ab-kicker" href="/concepts/installs">Install</a>
              <span className="nuon-ab-branch-name">globex</span>
            </div>
            <div className="nuon-infra-components">
              <div className="nuon-ab-head">
                <a className="nuon-infra-comps-title" href="/concepts/stacks">Stack</a>
                <span className="nuon-pd-where">applied by the customer</span>
              </div>
              <a className="nuon-pd-runner" href="/concepts/runners" ref={runnerRef}>
                <svg className="nuon-pd-runner-mark" viewBox="0 0 34 46" fill="currentColor" aria-hidden="true">
                  <Mark />
                </svg>
                <span className="nuon-pd-runner-text">
                  <span className="nuon-pd-runner-name">Runner</span>
                  <span className="nuon-pd-runner-note">Managed by Nuon</span>
                </span>
              </a>
            </div>
            <div className="nuon-infra-components">
              <a className="nuon-infra-comps-title" href="/concepts/sandboxes">Sandbox</a>
              <div className="nuon-infra-chips">
                <span className="nuon-infra-chip">API</span>
                <span className="nuon-infra-chip">Database</span>
                <span className="nuon-infra-chip">Storage bucket</span>
              </div>
            </div>
          </div>
          <div className="nuon-infra-concept nuon-pd-faded">
            <div className="nuon-ab-head">
              <a className="nuon-ab-kicker" href="/concepts/installs">Install</a>
              <span className="nuon-ab-branch-name">initech</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export const RelationshipTree = () => {
  const Card = ({ name, file, order, isDefault, mono }) => (
    <div className="nuon-pt-card">
      <span className={mono ? "nuon-ab-branch-name" : "nuon-pt-name"}>
        {order && <span className="nuon-pt-num">{order}</span>}
        {name}
        {isDefault && <span className="nuon-pt-default">default</span>}
      </span>
      {file && <span className="nuon-pt-file">{file}</span>}
    </div>
  );
  const Installs = ({ names }) => (
    <div className="nuon-pt-installs">
      {names.map((n) => (
        <span key={n} className="nuon-pt-install">{n}</span>
      ))}
    </div>
  );
  return (
    <div className="not-prose nuon-infra nuon-pt">
      <div className="nuon-pt-legend">
        <span><i className="nuon-pt-key-repo" />Defined in your repo</span>
        <span><i className="nuon-pt-key-cust" />Runs in a customer's account</span>
      </div>
      <div className="nuon-pt-scroll">
        <div
          className="nuon-pt-tree"
          role="img"
          aria-label="App template acme has two app branches. main has install groups canary (order 1) with acme-internal and acme-staging, then customers (order 2, default) with globex, initech, and umbrella. big-customer has one install group, production, with hooli."
        >
          <div className="nuon-pt-heads">
            <a className="nuon-ab-kicker nuon-pt-h1" href="/concepts/apps">App template</a>
            <a className="nuon-ab-kicker nuon-pt-h2" href="/concepts/app-branches">App branches</a>
            <a className="nuon-ab-kicker nuon-pt-h2" href="/concepts/app-branches#install-groups">Install groups</a>
            <a className="nuon-ab-kicker" href="/concepts/installs">Installs</a>
          </div>
          <div className="nuon-pt-item nuon-pt-l1">
            <Card name="acme" file="metadata.toml" />
            <div className="nuon-pt-children">
              <div className="nuon-pt-item nuon-pt-l2">
                <Card name="main" file="branches/main.toml" mono />
                <div className="nuon-pt-children">
                  <div className="nuon-pt-item nuon-pt-l3">
                    <Card name="canary" file='tier = "canary"' order="1" />
                    <div className="nuon-pt-children">
                      <div className="nuon-pt-item"><Installs names={["acme-internal", "acme-staging"]} /></div>
                    </div>
                  </div>
                  <div className="nuon-pt-item nuon-pt-l3">
                    <Card name="customers" file="every other install" order="2" isDefault />
                    <div className="nuon-pt-children">
                      <div className="nuon-pt-item"><Installs names={["globex", "initech", "umbrella"]} /></div>
                    </div>
                  </div>
                </div>
              </div>
              <div className="nuon-pt-item nuon-pt-l2">
                <Card name="big-customer" file="branches/big-customer.toml" mono />
                <div className="nuon-pt-children">
                  <div className="nuon-pt-item nuon-pt-l3">
                    <Card name="production" order="1" isDefault />
                    <div className="nuon-pt-children">
                      <div className="nuon-pt-item"><Installs names={["hooli"]} /></div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
