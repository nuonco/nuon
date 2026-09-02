import { expect, test } from '@playwright/test'

const EDITOR_STORY = '/?story=branches--branchcisettings--editor&mode=preview'

test('edits branch CI trigger settings', async ({ page }) => {
  await page.goto(EDITOR_STORY, { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Open modal' }).click()

  const dialog = page.getByRole('dialog')
  const save = dialog.getByRole('button', { name: 'Save changes' })
  const regex = dialog.getByLabel('Ignored changes regex (optional)')

  await expect(regex).toHaveValue('^docs/')
  await expect(regex).toHaveAttribute('placeholder', '^(docs/|README\\.md)')
  await expect(dialog.getByLabel('Send status for ignored runs')).toBeChecked()
  await expect(
    dialog.getByText('Send status for ignored runs', { exact: true })
  ).toBeVisible()
  await expect(save).toBeDisabled()

  await dialog.getByRole('button', { name: 'Ignore all changes' }).click()
  await expect(regex).toHaveValue('.*')
  await expect(save).toBeEnabled()
})
