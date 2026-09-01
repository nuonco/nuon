import { expect, test } from '@playwright/test'

const CREATE_STORY = '/?story=branches--branchform--create&mode=preview'
const EDIT_STORY = '/?story=branches--branchform--edit&mode=preview'

test.describe('BranchForm create behavior', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(CREATE_STORY, { waitUntil: 'domcontentloaded' })
    await page.getByRole('button', { name: 'Open modal' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
  })

  test('submit stays disabled until the branch name is valid', async ({
    page,
  }) => {
    const dialog = page.getByRole('dialog')
    const submit = dialog.getByRole('button', { name: 'Create branch' })

    await expect(submit).toBeDisabled()

    await dialog.getByLabel('Branch name').fill('production')

    await expect(submit).toBeEnabled()
  })

  test('the branch name errors only after it is touched', async ({ page }) => {
    const dialog = page.getByRole('dialog')
    const name = dialog.getByLabel('Branch name')

    await expect(dialog.getByText('Branch name is required')).toBeHidden()

    await name.click()
    await name.blur()
    await expect(dialog.getByText('Branch name is required')).toBeVisible()

    await name.fill('production')
    await expect(dialog.getByText('Branch name is required')).toBeHidden()
  })
})

test.describe('BranchForm edit behavior', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(EDIT_STORY, { waitUntil: 'domcontentloaded' })
    await page.getByRole('button', { name: 'Open modal' }).click()
    await expect(page.getByRole('dialog')).toBeVisible()
  })

  test('edit prefills the name and opens valid; clearing it disables save', async ({
    page,
  }) => {
    const dialog = page.getByRole('dialog')
    const submit = dialog.getByRole('button', { name: 'Save changes' })

    await expect(submit).toBeEnabled()

    await dialog.getByLabel('Branch name').fill('')
    await expect(submit).toBeDisabled()

    await dialog.getByLabel('Branch name').fill('production')
    await expect(submit).toBeEnabled()
  })

  test('ignore-all maps to an enabled GitHub changes toggle', async ({
    page,
  }) => {
    const dialog = page.getByRole('dialog')

    await expect(dialog.getByLabel('Ignore all GitHub changes')).toBeChecked()
    await expect(
      dialog.getByText(
        'Git push and pull request events still create a run, but the run is marked not attempted immediately.'
      )
    ).toBeVisible()
  })
})
