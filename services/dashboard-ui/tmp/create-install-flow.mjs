import { chromium } from 'playwright-core'
import { mkdirSync } from 'node:fs'

const ADMIN = 'http://127.0.0.1:8082'
const PUB = 'http://127.0.0.1:8081'
const APP = 'http://127.0.0.1:4000'
const EMAIL = process.env.NUON_DEV_EMAIL ?? 'seed@nuon.co'
const ORG = 'orgkxwo8cgonbqez7me7x8qz7d'

const seed = await fetch(`${ADMIN}/v1/general/seed-user`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json', 'X-Nuon-Admin-Email': EMAIL },
  body: '{}',
})
const { api_token } = JSON.parse((await seed.text()).match(/^\{[^}]*\}/)[0])

mkdirSync('/tmp/flash', { recursive: true })

const browser = await chromium.launch()
const ctx = await browser.newContext({
  viewport: { width: 1280, height: 900 },
  deviceScaleFactor: 1,
})
await ctx.addCookies([
  { name: 'X-Nuon-Auth', value: api_token, domain: '127.0.0.1', path: '/', httpOnly: true, sameSite: 'Lax' },
])
const page = await ctx.newPage()

await page.goto(`${APP}/${ORG}/installs`, { waitUntil: 'domcontentloaded' })
await page.getByRole('button', { name: 'Create install' }).first().click()
await page.getByText('Select an app to create an install').waitFor()

const appRadio = page.getByRole('radio', { name: /httpbin/ })
await appRadio.waitFor({ timeout: 10000 })
await appRadio.click({ force: true })

const nameInput = page.getByPlaceholder('Enter install name')
await nameInput.waitFor({ timeout: 10000 })
await nameInput.fill(`e2e-flash-${Date.now()}`)

const awsSettings = page.getByRole('group', { name: /AWS settings/i })
if (await awsSettings.isVisible({ timeout: 5000 }).catch(() => false)) {
  await awsSettings.getByRole('combobox').click()
  const search = page.getByPlaceholder('Search...')
  await search.waitFor()
  await search.fill('us-west-2')
  await page.getByRole('option', { name: /us-west-2/ }).first().click()
}
await page.getByText('Auto-approve changes').click()

// Instant DOM probe (no auto-waiting)
const probe = () =>
  page.evaluate(() => {
    const d = document.querySelector('[role="dialog"]')
    const txt = (d?.textContent || '').replace(/\s+/g, ' ').trim()
    const rect = d?.getBoundingClientRect()
    const heading = d?.querySelector('h1,h2,h3,[class*="head"]')?.textContent?.trim() || ''
    return {
      url: location.pathname,
      dialog: !!d,
      h: rect ? Math.round(rect.height) : null,
      w: rect ? Math.round(rect.width) : null,
      finishing: txt.includes('Finishing up'),
      branches: txt.includes('Connect to app branches'),
      loadingBranches: txt.includes('Loading app branches'),
      hasForm: !!document.querySelector('[placeholder="Enter install name"]'),
      hasSpinnerOnly: txt.length > 0 && txt.length < 60,
      snippet: txt.slice(0, 80),
    }
  })

const timeline = []
timeline.push({ t: 'before', ...(await probe()) })

await page.getByRole('button', { name: 'Create install' }).last().click()

const start = Date.now()
let i = 0
let navAt = null
while (Date.now() - start < 12000) {
  const p = await probe()
  timeline.push({ t: `+${Date.now() - start}ms`, ...p })
  await page.screenshot({ path: `/tmp/flash/frame-${String(i).padStart(3, '0')}.png` }).catch(() => {})
  i++
  if (/\/workflows\//.test(p.url)) {
    if (navAt === null) navAt = Date.now() - start
    if (Date.now() - start - navAt > 500) break
  }
  await page.waitForTimeout(40)
}

console.log(JSON.stringify(timeline.filter((r, idx) => idx === 0 || idx === timeline.length - 1 || r.finishing || r.branches || r.loadingBranches || r.hasSpinnerOnly || !r.hasForm), null, 2))
console.log('frames:', i, 'navAt:', navAt)

await ctx.close()
await browser.close()
