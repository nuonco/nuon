// Install status view — polls the proxy until the install reaches a terminal
// state. Creation is async (see references/lifecycle.md).

import { useEffect, useRef, useState } from "react";
import { getInstall } from "./installsApi";

const TERMINAL = new Set(["active", "healthy", "error", "failed"]);

export function InstallStatus({ installId }) {
  const [install, setInstall] = useState(null);
  const [error, setError] = useState(null);
  const timer = useRef(null);

  useEffect(() => {
    let cancelled = false;

    async function poll() {
      try {
        const data = await getInstall(installId);
        if (cancelled) return;
        setInstall(data);
        const done =
          TERMINAL.has(data.sandbox_status) && TERMINAL.has(data.runner_status);
        if (!done) timer.current = setTimeout(poll, 5000);
      } catch (err) {
        if (!cancelled) setError(err.message);
      }
    }

    poll();
    return () => {
      cancelled = true;
      if (timer.current) clearTimeout(timer.current);
    };
  }, [installId]);

  if (error) return <p role="alert">{error}</p>;
  if (!install) return <p>Loading…</p>;

  return (
    <div>
      <h2>{install.name}</h2>
      <dl>
        <dt>Sandbox</dt>
        <dd>{install.sandbox_status || "provisioning…"}</dd>
        <dt>Runner</dt>
        <dd>{install.runner_status || "provisioning…"}</dd>
      </dl>
    </div>
  );
}
