// Minimal Nuon ctl-api client. Server-side ONLY — holds the API token.
// Injects Bearer auth + X-Nuon-Org-ID and normalizes errors.

const NUON_API_URL = process.env.NUON_API_URL || "https://ctl.prod.nuon.co";
const NUON_API_TOKEN = process.env.NUON_API_TOKEN;
const NUON_ORG_ID = process.env.NUON_ORG_ID;

if (!NUON_API_TOKEN || !NUON_ORG_ID) {
  throw new Error("NUON_API_TOKEN and NUON_ORG_ID must be set (server-side).");
}

class NuonError extends Error {
  constructor(status, body) {
    super(`nuon ctl-api error: ${status}`);
    this.status = status;
    this.body = body;
  }
}

async function request(method, path, body) {
  const res = await fetch(`${NUON_API_URL}${path}`, {
    method,
    headers: {
      Authorization: `Bearer ${NUON_API_TOKEN}`,
      "X-Nuon-Org-ID": NUON_ORG_ID,
      ...(body ? { "Content-Type": "application/json" } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  const text = await res.text();
  const parsed = text ? JSON.parse(text) : undefined;
  if (!res.ok) throw new NuonError(res.status, parsed);
  return parsed;
}

const nuon = {
  getAppInputSchema: (appId) =>
    request("GET", `/v1/apps/${appId}/input-latest-config`),
  createInstall: (appId, payload) =>
    request("POST", `/v1/apps/${appId}/installs`, payload),
  getInstall: (installId) => request("GET", `/v1/installs/${installId}`),
};

module.exports = { nuon, NuonError };
