import { rmSync, mkdirSync, copyFileSync } from "fs";

rmSync(new URL("../dist", import.meta.url).pathname, { recursive: true, force: true });
mkdirSync(new URL("../dist/assets", import.meta.url).pathname, { recursive: true });
copyFileSync(
  new URL("../client/index.html", import.meta.url).pathname,
  new URL("../dist/index.html", import.meta.url).pathname,
);
