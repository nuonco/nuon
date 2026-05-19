import { watch } from "fs";

const BFF_PORT = Number(process.env.HTTP_PORT || 4000);
const DEV_PORT = BFF_PORT + 1;
const BFF_ORIGIN = `http://localhost:${BFF_PORT}`;

const RELOAD_SCRIPT = `<script>new EventSource("/__dev/reload").onmessage=()=>location.reload()</script>`;

const clients = new Set();

let debounce;
watch("dist", { recursive: true }, () => {
  clearTimeout(debounce);
  debounce = setTimeout(() => {
    for (const c of clients) c.enqueue("data: reload\n\n");
  }, 500);
});

Bun.serve({
  port: DEV_PORT,
  async fetch(req) {
    const url = new URL(req.url);

    if (url.pathname === "/__dev/reload") {
      let controller;
      const stream = new ReadableStream({
        start(c) {
          controller = c;
          clients.add(c);
        },
        cancel() {
          clients.delete(controller);
        },
      });
      return new Response(stream, {
        headers: {
          "content-type": "text/event-stream",
          "cache-control": "no-cache",
        },
      });
    }

    const proxyUrl = new URL(url.pathname + url.search, BFF_ORIGIN);
    const res = await fetch(proxyUrl, {
      method: req.method,
      headers: req.headers,
      body: req.body,
      redirect: "manual",
    });

    if ((res.headers.get("content-type") || "").includes("text/html")) {
      const html = await res.text();
      return new Response(
        html.replace("</body>", RELOAD_SCRIPT + "</body>"),
        {
          status: res.status,
          headers: res.headers,
        },
      );
    }

    return res;
  },
});

console.log(
  `Dev proxy on http://localhost:${DEV_PORT} → http://localhost:${BFF_PORT}`,
);

const BFF_CHECK_INTERVAL = 5000;
let failCount = 0;
const MAX_FAILURES = 12;

setInterval(async () => {
  try {
    await fetch(BFF_ORIGIN, { signal: AbortSignal.timeout(2000), redirect: "manual" });
    if (failCount > 0) {
      console.log(`BFF is back after ${failCount} failed checks`);
      failCount = 0;
    }
  } catch {
    failCount++;
    console.log(`BFF unreachable (${failCount}/${MAX_FAILURES})`);
    if (failCount >= MAX_FAILURES) {
      console.log("BFF unreachable for 1 minute, shutting down dev server");
      process.exit(0);
    }
  }
}, BFF_CHECK_INTERVAL);
