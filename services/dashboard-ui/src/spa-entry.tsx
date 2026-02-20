import React from "react";
import { createRoot } from "react-dom/client";

// SPA entry point — used when DASHBOARD_MODE=go.
// This will be expanded with react-router-dom and the full provider tree
// once the RSC → client component migration is done (Phase 6).
// For now, this is a minimal bootstrap to validate the Vite SPA build pipeline.

function App() {
  return (
    <div id="ui-version" className="hidden">
      Version: {process.env.VERSION || "development"}
    </div>
  );
}

const container = document.getElementById("root");
if (container) {
  const root = createRoot(container);
  root.render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  );
}
