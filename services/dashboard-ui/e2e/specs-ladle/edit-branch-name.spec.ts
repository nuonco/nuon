import { expect, test } from '@playwright/test'

const EDIT_NAME_STORY =
  '/?story=branches--editbranchnamemodal--edit-name&mode=preview'

test('edits the branch name without source settings', async ({ page }) => {
  await page.goto(EDIT_NAME_STORY, { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Open modal' }).click()

  const dialog = page.getByRole('dialog')
  const save = dialog.getByRole('button', { name: 'Save changes' })

  await expect(dialog.getByText('Edit name', { exact: true })).toBeVisible()
  await expect(dialog.getByLabel('Branch name')).toHaveValue('production')
  await expect(dialog.getByLabel('Repository')).toHaveCount(0)
  await expect(save).toBeDisabled()

  await dialog.getByLabel('Branch name').fill('main')
  await expect(save).toBeEnabled()
})
