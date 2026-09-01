import { expect, test } from '@playwright/test'

const CONNECTED_STORY =
  '/?story=branches--branchsourcecard--connected&mode=preview'

test('puts source editing on the source card', async ({ page }) => {
  await page.goto(CONNECTED_STORY, { waitUntil: 'domcontentloaded' })

  await expect(page.getByRole('button', { name: 'Edit source' })).toBeVisible()
  await expect(page.getByText('Path filter', { exact: true })).toHaveCount(0)
  await expect(page.getByText('Repository', { exact: true })).toBeVisible()
  await expect(page.getByText('Directory', { exact: true })).toBeVisible()
})
