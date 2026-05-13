import { readdirSync, readFileSync, writeFileSync } from "fs";
import { join } from "path";
import { gzipSync } from "zlib";

const assetsDir = new URL("../dist/assets", import.meta.url).pathname;

const files = readdirSync(assetsDir).filter(
  (f) => f.endsWith(".js") || f.endsWith(".css"),
);

for (const file of files) {
  const filePath = join(assetsDir, file);
  const content = readFileSync(filePath);
  const gz = gzipSync(content, { level: 9 });
  writeFileSync(`${filePath}.gz`, gz);

  const pct = ((1 - gz.length / content.length) * 100).toFixed(0);
  const sizeKB = (content.length / 1024).toFixed(0);
  const gzKB = (gz.length / 1024).toFixed(0);
  console.log(`  ${file}: ${sizeKB}KB → ${gzKB}KB (${pct}% smaller)`);
}
