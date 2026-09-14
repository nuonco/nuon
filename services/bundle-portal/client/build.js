#!/usr/bin/env bun

import path from "node:path";

const root = import.meta.dir;
const sharedPackages = /^(react(?:\/.*)?|react-dom(?:\/.*)?|react-router)$/;
const sharedPackagePath = (packageName) => {
  if (packageName === "react-router") {
    return path.join(
      root,
      "node_modules/react-router/dist/development/index.mjs",
    );
  }
  if (packageName === "react" || packageName === "react-dom") {
    return path.join(root, "node_modules", packageName, "index.js");
  }
  return path.join(root, "node_modules", `${packageName}.js`);
};

const result = await Bun.build({
  entrypoints: [path.join(root, "src/index.tsx")],
  outdir: path.join(root, "dist"),
  target: "browser",
  minify: true,
  naming: "index.[ext]",
  tsconfig: path.join(root, "tsconfig.build.json"),
  plugins: [
    {
      name: "shared-react-runtime",
      setup(build) {
        build.onResolve(
          { filter: sharedPackages },
          ({ path: packageName }) => ({
            path: sharedPackagePath(packageName),
          }),
        );
      },
    },
  ],
});

if (!result.success) {
  for (const log of result.logs) console.error(log);
  process.exit(1);
}
