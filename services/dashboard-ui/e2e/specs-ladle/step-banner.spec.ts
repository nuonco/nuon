import { expect, test } from '@playwright/test'

const LONG_DESCRIPTION =
  '/?story=workflows--stepbanner--error-with-long-unbroken-description&mode=preview'

test('a long unbroken error description stays inside the banner', async ({
  page,
}) => {
  await page.setViewportSize({ width: 1024, height: 700 })
  await page.goto(LONG_DESCRIPTION, { waitUntil: 'domcontentloaded' })
  await page.getByText(/Step teardown apply plan acme-widgets/).waitFor()

  const overflow = await page.evaluate(
    () => document.body.scrollWidth - document.body.clientWidth
  )
  expect(overflow).toBe(0)

  await expect(page.getByRole('button', { name: 'Retry step' })).toBeVisible()
})
