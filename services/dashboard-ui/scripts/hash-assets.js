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

const stylesFile = Bun.file(`${dist}assets/styles.css`);
let stylesHashedName = null;
if (await stylesFile.exists()) {
  const content = await stylesFile.arrayBuffer();
  const hash = new Bun.CryptoHasher("md5")
    .update(content)
    .digest("hex")
    .slice(0, 8);
  stylesHashedName = `styles-${hash}.css`;
  await Bun.write(`${dist}assets/${stylesHashedName}`, stylesFile);
  const { unlinkSync } = require("fs");
  unlinkSync(`${dist}assets/styles.css`);
}

let html = await Bun.file(new URL("../client/index.html", import.meta.url).pathname).text();

if (jsBasename) {
  html = html.replace(`/assets/app.js`, `/assets/${jsBasename}`);
}

if (cssFromBundlerBasename) {
  html = html.replace(`/assets/app.css`, `/assets/${cssFromBundlerBasename}`);
} else {
  html = html.replace(
    /\s*<link\s+rel="stylesheet"\s+href="\/assets\/app\.css"\s*\/?\s*>\s*/,
    "\n",
  );
}

if (stylesHashedName) {
  html = html.replace(`/assets/styles.css`, `/assets/${stylesHashedName}`);
}

await Bun.write(`${dist}index.html`, html);

console.log("Asset hashing complete:");
if (jsBasename) console.log(`  JS:     /assets/${jsBasename}`);
if (cssFromBundlerBasename)
  console.log(`  CSS:    /assets/${cssFromBundlerBasename}`);
if (stylesHashedName)
  console.log(`  Styles: /assets/${stylesHashedName}`);
