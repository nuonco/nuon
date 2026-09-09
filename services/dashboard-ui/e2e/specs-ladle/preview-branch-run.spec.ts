import { expect, test } from '@playwright/test'

const STORY =
  '/?story=branches--previewbranchrunmodal--branch-source&mode=preview'

test('searches branches when selecting a preview source', async ({ page }) => {
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Open preview run' }).click()

  const dialog = page.getByRole('dialog')
  await expect(dialog).toBeVisible()
  await expect(dialog).toBeFocused()

  const branchSelect = dialog.getByRole('combobox')
  await branchSelect.click()

  const search = page.getByPlaceholder('Search...')
  await expect(search).toBeFocused()
  await search.fill('gitops')

  await expect(page.getByRole('option')).toHaveCount(1)
  await expect(page.getByRole('option', { name: 'am/gitops-ui' })).toBeVisible()
  await expect(
    page.getByRole('option', { name: 'am/gcp-custom-root-domain' })
  ).toHaveCount(0)
})
