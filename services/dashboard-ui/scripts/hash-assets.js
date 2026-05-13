const dist = new URL("../dist/", import.meta.url).pathname;

const meta = await Bun.file(`${dist}meta.json`).json();

const outputFiles = Object.keys(meta.outputs);
const jsFile = outputFiles.find((f) => f.endsWith(".js"));
const cssFromBundler = outputFiles.find((f) => f.endsWith(".css"));

const basename = (f) => f.split("/").pop();
const jsBasename = jsFile ? basename(jsFile) : null;
const cssFromBundlerBasename = cssFromBundler ? basename(cssFromBundler) : null;

const stylesFile = Bun.file(`${dist}styles.css`);
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
  unlinkSync(`${dist}styles.css`);
}

let html = await Bun.file(new URL("../client/index.html", import.meta.url).pathname).text();

if (jsBasename) {
  html = html.replace(`/app.js`, `/assets/${jsBasename}`);
}

if (cssFromBundlerBasename) {
  html = html.replace(`/app.css`, `/assets/${cssFromBundlerBasename}`);
} else {
  html = html.replace(
    /\s*<link\s+rel="stylesheet"\s+href="\/app\.css"\s*\/?\s*>\s*/,
    "\n",
  );
}

if (stylesHashedName) {
  html = html.replace(`/styles.css`, `/assets/${stylesHashedName}`);
}

await Bun.write(`${dist}index.html`, html);

console.log("Asset hashing complete:");
if (jsBasename) console.log(`  JS:     /assets/${jsBasename}`);
if (cssFromBundlerBasename)
  console.log(`  CSS:    /assets/${cssFromBundlerBasename}`);
if (stylesHashedName)
  console.log(`  Styles: /assets/${stylesHashedName}`);
