#!/usr/bin/env bun

import postcss from "postcss";
import tailwindcss from "@tailwindcss/postcss";
import { watch } from "fs";

const INPUT = new URL("../client/styles.css", import.meta.url).pathname;
const OUTPUT = new URL("../dist/styles.css", import.meta.url).pathname;

const isWatch = process.argv.includes("--watch");

async function build() {
  const source = await Bun.file(INPUT).text();
  const result = await postcss([tailwindcss()]).process(source, {
    from: INPUT,
    to: OUTPUT,
  });
  await Bun.write(OUTPUT, result.css);
}

await build();

if (isWatch) {
  console.log("Watching for CSS changes...");

  let debounce;
  watch(new URL("../client", import.meta.url).pathname, { recursive: true }, (event, filename) => {
    if (!filename?.match(/\.(css|tsx?|html)$/)) return;
    clearTimeout(debounce);
    debounce = setTimeout(async () => {
      try {
        await build();
      } catch (e) {
        console.error("CSS build error:", e.message);
      }
    }, 100);
  });
}
