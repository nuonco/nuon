#!/usr/bin/env bun

import postcss from "postcss";
import tailwindcss from "@tailwindcss/postcss";
import { watch } from "fs";

const ENTRIES = [
  { input: "../client/styles.css", output: "../dist/assets/styles.css" },
  { input: "../client/lite/styles.css", output: "../dist/assets/lite.css" },
].map(({ input, output }) => ({
  input: new URL(input, import.meta.url).pathname,
  output: new URL(output, import.meta.url).pathname,
}));

const isWatch = process.argv.includes("--watch");

async function buildEntry({ input, output }) {
  const source = await Bun.file(input).text();
  const result = await postcss([tailwindcss()]).process(source, {
    from: input,
    to: output,
  });
  await Bun.write(output, result.css);
}

async function build() {
  await Promise.all(ENTRIES.map(buildEntry));
}

await build();

const HTML_INPUT = new URL("../client/index.html", import.meta.url).pathname;
const HTML_OUTPUT = new URL("../dist/index.html", import.meta.url).pathname;

async function copyHTML() {
  await Bun.write(HTML_OUTPUT, Bun.file(HTML_INPUT));
}

if (isWatch) {
  console.log("Watching for CSS changes...");

  let debounce;
  watch(new URL("../client", import.meta.url).pathname, { recursive: true }, (event, filename) => {
    if (!filename?.match(/\.(css|tsx?|html)$/)) return;
    clearTimeout(debounce);
    debounce = setTimeout(async () => {
      try {
        await build();
        if (filename.endsWith(".html")) await copyHTML();
      } catch (e) {
        console.error("CSS build error:", e.message);
      }
    }, 100);
  });
}
