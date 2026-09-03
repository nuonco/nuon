import { expect, test } from '@playwright/test'

const STORY = '/?story=branches--previewconfigsection--edit-modal&mode=preview'

test.describe('PreviewConfig form behavior', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: 'domcontentloaded' })
    await page.getByRole('button', { name: 'Edit preview settings' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
  })

  test('save is disabled until a setting changes', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    const submit = dialog.getByRole('button', { name: 'Save changes' })

    await expect(submit).toBeDisabled()
    await dialog.getByLabel('Apply').check()
    await expect(submit).toBeEnabled()
  })

  test('build and validate mode hides the default install', async ({
    page,
  }) => {
    const dialog = page.getByRole('dialog')

    await expect(dialog.getByLabel('Default install')).toBeVisible()
    await dialog.getByLabel('Build and validate').check()
    await expect(dialog.getByLabel('Default install')).toBeHidden()
  })

  test('GitHub settings are keyboard-operable', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    const comments = dialog.getByLabel('Comment on pull request')

    await comments.focus()
    await page.keyboard.press('Space')
    await expect(comments).toBeChecked()
  })
})
