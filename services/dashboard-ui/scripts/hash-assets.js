import { unlinkSync } from "fs";

const dist = new URL("../dist/", import.meta.url).pathname;

const meta = await Bun.file(`${dist}meta.json`).json();

const outputFiles = Object.entries(meta.outputs);
const jsFile = outputFiles.find(
  ([f, info]) => f.endsWith(".js") && info.entryPoint === "client/index.tsx",
)?.[0];
const cssFromBundler = outputFiles.find(
  ([f, info]) => f.endsWith(".css") && info.entryPoint === "client/index.tsx",
)?.[0];

const basename = (f) => f.split("/").pop();
const jsBasename = jsFile ? basename(jsFile) : null;
const cssFromBundlerBasename = cssFromBundler ? basename(cssFromBundler) : null;

async function hashStylesheet(name) {
  const file = Bun.file(`${dist}assets/${name}.css`);
  if (!(await file.exists())) return null;

  const content = await file.arrayBuffer();
  const hash = new Bun.CryptoHasher("md5")
    .update(content)
    .digest("hex")
    .slice(0, 8);
  const hashedName = `${name}-${hash}.css`;
  await Bun.write(`${dist}assets/${hashedName}`, file);
  unlinkSync(`${dist}assets/${name}.css`);
  return hashedName;
}

const stylesHashedName = await hashStylesheet("styles");
const liteHashedName = await hashStylesheet("lite");

let html = await Bun.file(new URL("../client/index.html", import.meta.url).pathname).text();

if (jsBasename) {
  html = html.replace(`/assets/app.js`, `/assets/${jsBasename}`);
}

if (cssFromBundlerBasename) {
  html = html.replace(`/assets/app.css`, `/assets/${cssFromBundlerBasename}`);
} else {
  html = html.replace(
    /[\t ]*<link[^>]*href="\/assets\/app\.css"[^>]*>\n?/,
    "",
  );
}

if (stylesHashedName) {
  html = html.replace(`/assets/styles.css`, `/assets/${stylesHashedName}`);
}

if (liteHashedName) {
  html = html.replace(`/assets/lite.css`, `/assets/${liteHashedName}`);
}

await Bun.write(`${dist}index.html`, html);

console.log("Asset hashing complete:");
if (jsBasename) console.log(`  JS:     /assets/${jsBasename}`);
if (cssFromBundlerBasename)
  console.log(`  CSS:    /assets/${cssFromBundlerBasename}`);
if (stylesHashedName) console.log(`  Styles: /assets/${stylesHashedName}`);
if (liteHashedName) console.log(`  Lite:   /assets/${liteHashedName}`);
