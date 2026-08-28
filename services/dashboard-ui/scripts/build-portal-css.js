#!/usr/bin/env bun

import postcss from 'postcss'
import tailwindcss from '@tailwindcss/postcss'

const dashboardStyles = new URL('../client/styles.css', import.meta.url)
  .pathname
const portalStyles = new URL(
  '../../bundle-portal/client/src/styles.css',
  import.meta.url
).pathname
const output = new URL(
  '../../bundle-portal/client/dist/index.css',
  import.meta.url
).pathname
const scopePortalStyles = {
  postcssPlugin: 'scope-portal-styles',
  Rule(rule) {
    if (
      rule.parent?.type === 'atrule' &&
      rule.parent.name.endsWith('keyframes')
    )
      return
    rule.selectors = rule.selectors.map((selector) => {
      if (selector === ':root' || selector === 'html' || selector === 'body')
        return selector
      return selector === '*'
        ? '.offline-portal *'
        : `.offline-portal ${selector}`
    })
  },
}
const portalSource = await postcss([scopePortalStyles]).process(
  await Bun.file(portalStyles).text(),
  {
    from: portalStyles,
  }
)
const bundledComponentStyles = (await Bun.file(output).exists())
  ? await Bun.file(output).text()
  : ''
const dashboardSource = (await Bun.file(dashboardStyles).text()).replace(
  /^@import url\([^\n]+\);\n/,
  ''
)
const source = `${dashboardSource}
@source "./components/common";
@source "./components/layout";
@source "./components/approvals";
@source "./components/apps/bundles";
@source "./components/log-stream";
@source "./components/surfaces";
@source "./components/workflows";
@source "../../bundle-portal/client/src";
${bundledComponentStyles}
${portalSource.css}`
const result = await postcss([tailwindcss()]).process(source, {
  from: dashboardStyles,
  to: output,
})

await Bun.write(output, result.css)
