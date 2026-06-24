import { readdirSync, statSync } from "fs";
import { join } from "path";

const BFF_PORT = Number(process.env.HTTP_PORT || 4000);
const DEV_PORT = BFF_PORT + 1;
const BFF_ORIGIN = `http://localhost:${BFF_PORT}`;

const RELOAD_SCRIPT = `<script>(function(){var es,last=Date.now();function connect(){es=new EventSource("/__dev/reload");es.onmessage=function(e){last=Date.now();if(e.data==="reload")location.reload()};es.onerror=function(){try{es.close()}catch(_){}setTimeout(connect,1000)}}connect();setInterval(function(){if(Date.now()-last>50000){try{es.close()}catch(_){}connect()}},10000)})();</script>`;

const clients = new Set();

function broadcast(data) {
  for (const c of clients) {
    try {
      c.enqueue(data);
    } catch {
      clients.delete(c);
    }
  }
}

// fs.watch({recursive:true}) is unreliable on Linux (inotify) — it silently
// stops emitting after a while. dist/ is a handful of files, so we poll mtimes.
function distSignature() {
  const parts = [];
  const walk = (dir) => {
    let entries;
    try {
      entries = readdirSync(dir, { withFileTypes: true });
    } catch {
      return;
    }
    for (const e of entries) {
      const p = join(dir, e.name);
      if (e.isDirectory()) {
        walk(p);
      } else {
        try {
          const s = statSync(p);
          parts.push(`${p}:${s.mtimeMs}:${s.size}`);
        } catch {
          continue;
        }
      }
    }
  };
  walk("dist");
  return parts.sort().join("|");
}

let lastSig = distSignature();
let debounce;
setInterval(() => {
  const sig = distSignature();
  if (sig === lastSig) return;
  lastSig = sig;
  clearTimeout(debounce);
  debounce = setTimeout(() => broadcast("data: reload\n\n"), 300);
}, 250);

setInterval(() => broadcast(": ping\n\ndata: ping\n\n"), 20000);

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
          connection: "keep-alive",
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
