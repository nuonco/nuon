import { expect, test } from '@playwright/test'

const STORY = '/?story=installs--editlabels--default&mode=preview'
const MANY_LABELS_STORY =
  '/?story=installs--editlabels--many-labels&mode=preview'

test.describe('EditLabels form behavior', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(STORY, { waitUntil: 'domcontentloaded' })
    await page.getByRole('button', { name: 'Open modal' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
  })

  test('save disables when a label key is cleared', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    const submit = dialog.getByRole('button', { name: 'Save labels' })

    await expect(submit).toBeEnabled()

    await dialog.getByLabel('Label 1 key').fill('')

    await expect(submit).toBeDisabled()
  })

  test('a label key errors only after it is touched', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    const key = dialog.getByLabel('Label 1 key')

    await expect(dialog.getByText('Key is required')).toBeHidden()

    await key.fill('')
    await key.blur()
    await expect(dialog.getByText('Key is required')).toBeVisible()

    await key.fill('env')
    await expect(dialog.getByText('Key is required')).toBeHidden()
  })
})

test('shows five labels before expanding the full list', async ({ page }) => {
  await page.goto(MANY_LABELS_STORY, { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Open modal with 20 labels' }).click()

  const dialog = page.getByRole('dialog')
  const toggle = dialog.getByRole('button', { name: 'Show 15 more labels' })

  await expect(dialog.getByRole('textbox')).toHaveCount(10)
  await expect(toggle).toHaveAttribute('aria-expanded', 'false')

  await toggle.click()
  await expect(dialog.getByRole('textbox')).toHaveCount(40)
  await expect(
    dialog.getByRole('button', { name: 'Show fewer labels' })
  ).toHaveAttribute('aria-expanded', 'true')

  await dialog.getByRole('button', { name: 'Show fewer labels' }).click()
  await expect(dialog.getByRole('textbox')).toHaveCount(10)
})

test('omits disclosure when five or fewer labels are present', async ({
  page,
}) => {
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Open modal' }).click()

  const dialog = page.getByRole('dialog')
  await expect(dialog.getByRole('textbox')).toHaveCount(6)
  await expect(
    dialog.getByRole('button', { name: /Show .* labels/ })
  ).toHaveCount(0)
})

test('expands the list when adding a label', async ({ page }) => {
  await page.goto(MANY_LABELS_STORY, { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Open modal with 20 labels' }).click()

  const dialog = page.getByRole('dialog')
  await dialog.getByRole('button', { name: 'Add label' }).click()

  await expect(dialog.getByLabel('Label 21 key')).toBeVisible()
  await expect(dialog.getByRole('textbox')).toHaveCount(42)
})

test('keeps invalid rows visible until their errors are fixed', async ({
  page,
}) => {
  await page.goto(MANY_LABELS_STORY, { waitUntil: 'domcontentloaded' })
  await page.getByRole('button', { name: 'Open modal with 20 labels' }).click()

  const dialog = page.getByRole('dialog')
  await dialog.getByRole('button', { name: 'Show 15 more labels' }).click()

  const hiddenRowKey = dialog.getByLabel('Label 6 key')
  await hiddenRowKey.fill('')
  await hiddenRowKey.blur()

  await expect(hiddenRowKey).toBeVisible()
  await expect(
    dialog.getByRole('button', { name: 'Show fewer labels' })
  ).toBeDisabled()
  await expect(
    dialog.getByRole('button', { name: 'Save labels' })
  ).toBeDisabled()

  await hiddenRowKey.fill('cost-center')
  await expect(
    dialog.getByRole('button', { name: 'Show fewer labels' })
  ).toBeEnabled()
})
