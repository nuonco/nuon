import { watch, existsSync } from "fs";
import devHtml from "../client/dev.html";

const BFF_PORT = Number(process.env.HTTP_PORT || 4000);
const DEV_PORT = BFF_PORT + 1;
const BFF_ORIGIN = `http://localhost:${BFF_PORT}`;

// --- Build config matching Go BFF's clientConfig ---

function env(key, fallback = "") {
  return process.env[key.toUpperCase()] ?? fallback;
}

const clientConfig = {
  apiUrl: env("NUON_API_URL", "https://api.nuon.co"),
  temporalUiUrl: env("NUON_TEMPORAL_UI_URL") || undefined,
  authServiceUrl: env("NUON_AUTH_SERVICE_URL") || undefined,
  appUrl: env("NUON_APP_URL", `http://localhost:${BFF_PORT}`),
  githubAppName: env("GITHUB_APP_NAME", "nuon-connect"),
  pylonAppId: env("PYLON_APP_ID") || undefined,
  datadogEnv: env("DATADOG_ENV") || undefined,
  datadogApiKey: env("DATADOG_API_KEY") || undefined,
  datadogApplicationKey: env("DATADOG_APPLICATION_KEY") || undefined,
  datadogTraceDebug: env("DATADOG_TRACE_DEBUG") === "true" || undefined,
  datadogApiUrl: env("DATADOG_API_URL") || undefined,
  version: env("VERSION") || undefined,
  gitRef: env("GIT_REF") || undefined,
  isByoc: env("NUON_BYOC") === "true",
  sfTrialEndpoint: env("SF_TRIAL_ACCESS_ENDPOINT") || undefined,
  onboardingV2: env("NUON_ONBOARDING_V2") === "true" || undefined,
  adminDashboardUrl: env("NUON_ADMIN_DASHBOARD_URL") || undefined,
};

const configScript = `window.__NUON_CONFIG__=${JSON.stringify(clientConfig)};`;

// --- Wait for BFF ---

async function waitForBFF(maxAttempts = 60) {
  for (let i = 0; i < maxAttempts; i++) {
    try {
      await fetch(BFF_ORIGIN, {
        signal: AbortSignal.timeout(2000),
        redirect: "manual",
      });
      return;
    } catch {
      if (i % 5 === 0) console.log(`Waiting for BFF on :${BFF_PORT}...`);
      await Bun.sleep(1000);
    }
  }
  throw new Error(`BFF not reachable at ${BFF_ORIGIN} after ${maxAttempts}s`);
}

await waitForBFF();

// --- CSS-only reload via SSE ---

const cssClients = new Map();

if (existsSync(new URL("../dist/assets", import.meta.url).pathname)) {
  let debounce;
  watch(new URL("../dist/assets", import.meta.url).pathname, (event, filename) => {
    if (filename !== "styles.css") return;
    clearTimeout(debounce);
    debounce = setTimeout(() => {
      for (const [c, heartbeat] of cssClients) {
        try {
          c.enqueue("data: css\n\n");
        } catch {
          clearInterval(heartbeat);
          cssClients.delete(c);
        }
      }
    }, 200);
  });
}

// --- Proxy helper ---

async function proxyToBFF(req) {
  const url = new URL(req.url);
  const proxyUrl = new URL(url.pathname + url.search, BFF_ORIGIN);
  return fetch(proxyUrl, {
    method: req.method,
    headers: req.headers,
    body: req.body,
    redirect: "manual",
  });
}

// --- Static file helpers ---

const DIST_DIR = new URL("../dist", import.meta.url).pathname;
const PUBLIC_DIR = new URL("../public", import.meta.url).pathname;

function servePublic(req) {
  const url = new URL(req.url);
  return new Response(Bun.file(`${PUBLIC_DIR}${url.pathname}`));
}

// --- Dev server ---

Bun.serve({
  port: DEV_PORT,
  development: true,
  routes: {
    "/__dev/config.js": new Response(configScript, {
      headers: { "content-type": "application/javascript" },
    }),

    "/__dev/css-reload": (req) => {
      let controller;
      let heartbeat;
      const stream = new ReadableStream({
        start(c) {
          controller = c;
          heartbeat = setInterval(() => {
            try {
              c.enqueue(": keepalive\n\n");
            } catch {
              clearInterval(heartbeat);
              cssClients.delete(c);
            }
          }, 30_000);
          cssClients.set(c, heartbeat);
        },
        cancel() {
          clearInterval(heartbeat);
          cssClients.delete(controller);
        },
      });
      return new Response(stream, {
        headers: {
          "content-type": "text/event-stream",
          "cache-control": "no-cache",
        },
      });
    },

    "/assets/*": (req) => {
      const url = new URL(req.url);
      return new Response(Bun.file(`${DIST_DIR}${url.pathname}`));
    },

    "/fonts/*": servePublic,
    "/images/*": servePublic,
    "/empty-graphics/*": servePublic,
    "/empty-state/*": servePublic,
    "/login-graphics/*": servePublic,
    "/onboarding-graphics/*": servePublic,
    "/favicon.svg": servePublic,
    "/favicon.ico": servePublic,

    "/": async (req) => {
      const hasCookie = (req.headers.get("cookie") || "").includes("X-Nuon-Auth");
      if (!hasCookie) {
        return new Response(null, {
          status: 302,
          headers: { location: clientConfig.authServiceUrl + "/?url=" + encodeURIComponent(`http://localhost:${DEV_PORT}/`) },
        });
      }
      return proxyToBFF(req);
    },
    "/v1/*": (req) => proxyToBFF(req),
    "/api/*": (req) => proxyToBFF(req),
    "/admin/*": (req) => proxyToBFF(req),

    "/*": devHtml,
  },

  async fetch(req) {
    return proxyToBFF(req);
  },
});

console.log(
  `Dev server on http://localhost:${DEV_PORT} (HMR) → BFF http://localhost:${BFF_PORT}`,
);

// --- Self-terminate when parent dies or BFF disappears ---

const PARENT_PID = process.ppid;

const BFF_CHECK_INTERVAL = 5000;
let failCount = 0;
const MAX_FAILURES = 6;

setInterval(async () => {
  try {
    process.kill(PARENT_PID, 0);
  } catch {
    console.log("Parent process gone, shutting down dev server");
    process.exit(0);
  }

  try {
    await fetch(BFF_ORIGIN, {
      signal: AbortSignal.timeout(2000),
      redirect: "manual",
    });
    if (failCount > 0) {
      console.log(`BFF is back after ${failCount} failed checks`);
      failCount = 0;
    }
  } catch {
    failCount++;
    console.log(`BFF unreachable (${failCount}/${MAX_FAILURES})`);
    if (failCount >= MAX_FAILURES) {
      console.log("BFF unreachable for 30s, shutting down dev server");
      process.exit(0);
    }
  }
}, BFF_CHECK_INTERVAL);
