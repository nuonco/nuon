// Frontend API layer — calls the VENDOR proxy, never ctl-api directly.
// No Nuon credentials ever reach this code.

async function json(res) {
  const body = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(body.error || `request failed (${res.status})`);
  return body;
}

export function getInstallInputs() {
  return fetch("/api/install-inputs", { credentials: "include" }).then(json);
}

export function createInstall({ name, inputs }) {
  return fetch("/api/installs", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ name, inputs }),
  }).then(json);
}

export function getInstall(id) {
  return fetch(`/api/installs/${id}`, { credentials: "include" }).then(json);
}
