const fs = require("fs");
const path = require("path");
const crypto = require("crypto");

const dist = path.resolve(__dirname, "../dist");

const meta = JSON.parse(fs.readFileSync(path.join(dist, "meta.json"), "utf8"));

const outputFiles = Object.keys(meta.outputs);
const jsFile = outputFiles.find((f) => f.endsWith(".js"));
const cssFromEsbuild = outputFiles.find((f) => f.endsWith(".css"));

const jsBasename = jsFile ? path.basename(jsFile) : null;
const cssFromEsbuildBasename = cssFromEsbuild
  ? path.basename(cssFromEsbuild)
  : null;

const stylesPath = path.join(dist, "styles.css");
let stylesHashedName = null;
if (fs.existsSync(stylesPath)) {
  const content = fs.readFileSync(stylesPath);
  const hash = crypto.createHash("md5").update(content).digest("hex").slice(0, 8);
  stylesHashedName = `styles-${hash}.css`;
  fs.copyFileSync(stylesPath, path.join(dist, "assets", stylesHashedName));
  fs.unlinkSync(stylesPath);
}

let html = fs.readFileSync(
  path.resolve(__dirname, "../client/index.html"),
  "utf8",
);

if (jsBasename) {
  html = html.replace(`/app.js`, `/assets/${jsBasename}`);
}

if (cssFromEsbuildBasename) {
  html = html.replace(`/app.css`, `/assets/${cssFromEsbuildBasename}`);
} else {
  html = html.replace(
    /\s*<link\s+rel="stylesheet"\s+href="\/app\.css"\s*\/?\s*>\s*/,
    "\n",
  );
}

if (stylesHashedName) {
  html = html.replace(`/styles.css`, `/assets/${stylesHashedName}`);
}

fs.writeFileSync(path.join(dist, "index.html"), html);

console.log("Asset hashing complete:");
if (jsBasename) console.log(`  JS:     /assets/${jsBasename}`);
if (cssFromEsbuildBasename)
  console.log(`  CSS:    /assets/${cssFromEsbuildBasename}`);
if (stylesHashedName)
  console.log(`  Styles: /assets/${stylesHashedName}`);
