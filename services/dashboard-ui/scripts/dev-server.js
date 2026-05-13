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
