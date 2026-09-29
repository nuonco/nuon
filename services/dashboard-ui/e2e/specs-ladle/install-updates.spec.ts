import { expect, test } from '@playwright/test'

const story = (id: string) => `/?story=${id}&mode=preview`

const TIMELINE_DEFAULT = story(
  'features--installs--updates--updates-timeline--default'
)
const TIMELINE_EMPTY = story('features--installs--updates--updates-timeline--empty')
const TIMELINE_HAS_MORE = story(
  'features--installs--updates--updates-timeline--has-more'
)
const TIMELINE_IMPACTED = story(
  'features--installs--updates--updates-timeline--impacted-and-failed'
)
const DETAILS_DEFAULT = story('features--installs--updates--update-details--default')
const DETAILS_IMPACTED = story(
  'features--installs--updates--update-details--impacted-but-unchanged'
)
const DETAILS_INSTALL_CONFIG = story(
  'features--installs--updates--update-details--install-config'
)
const DETAILS_FAILED = story('features--installs--updates--update-details--failed')
const CURRENT_RUN_APPLIED = story(
  'features--installs--updates--current-app-branch-run--applied'
)
const CURRENT_RUN_EMPTY = story('features--installs--updates--current-app-branch-run--empty')

test.describe('InstallUpdatesTimeline', () => {
  test('renders every update discriminant', async ({ page }) => {
    await page.goto(TIMELINE_DEFAULT, { waitUntil: 'domcontentloaded' })

    for (const label of ['App config', 'Stack', 'Inputs', 'Install config']) {
      await expect(page.getByText(label, { exact: true })).toBeVisible()
    }
  })

  test('shows the commit subject as the update title', async ({ page }) => {
    await page.goto(TIMELINE_DEFAULT, { waitUntil: 'domcontentloaded' })

    await expect(
      page.getByRole('button', {
        name: 'Add a cache component to the deployment plan',
      })
    ).toBeVisible()
  })

  test('offers load more only when another page exists', async ({ page }) => {
    await page.goto(TIMELINE_DEFAULT, { waitUntil: 'domcontentloaded' })
    await expect(page.getByRole('button', { name: 'Load more' })).toHaveCount(0)

    await page.goto(TIMELINE_HAS_MORE, { waitUntil: 'domcontentloaded' })
    await expect(page.getByRole('button', { name: 'Load more' })).toBeVisible()
  })

  test('explains the empty state', async ({ page }) => {
    await page.goto(TIMELINE_EMPTY, { waitUntil: 'domcontentloaded' })

    await expect(page.getByText('No updates yet')).toBeVisible()
  })

  test('opens impact reasons from an update in the timeline', async ({
    page,
  }) => {
    await page.goto(TIMELINE_IMPACTED, { waitUntil: 'domcontentloaded' })
    await page.getByRole('button', { name: 'App config updated' }).click()

    const panel = page.getByRole('complementary')
    await expect(panel.getByText('frontend', { exact: true })).toBeVisible()
    await expect(panel.getByText('Impacted', { exact: true })).toBeVisible()
  })
})

const openPanel = async (page: import('@playwright/test').Page) => {
  await page.getByRole('button', { name: 'Open panel' }).click()
  return page.getByRole('complementary')
}

test.describe('InstallUpdateDetails', () => {
  test('surfaces stack impacts and component impact reasons', async ({
    page,
  }) => {
    await page.goto(DETAILS_DEFAULT, { waitUntil: 'domcontentloaded' })
    const panel = await openPanel(page)

    await expect(panel.getByText(/stack impacts/i)).toBeVisible()
    await expect(panel.getByText('Permissions', { exact: true })).toBeVisible()

    await expect(panel.getByText(/component impacts/i)).toBeVisible()
    await expect(panel.getByText('api', { exact: true })).toBeVisible()
    await expect(
      panel.getByText('role.maintenance', { exact: true }).first()
    ).toBeVisible()
    await expect(panel.getByText(/via Operation role/i)).toBeVisible()
  })

  test('labels a checksum-identical component as impacted, not changed', async ({
    page,
  }) => {
    await page.goto(DETAILS_IMPACTED, { waitUntil: 'domcontentloaded' })
    const panel = await openPanel(page)

    await expect(panel.getByText('frontend', { exact: true })).toBeVisible()
    await expect(panel.getByText('Impacted', { exact: true })).toBeVisible()
    await expect(panel.getByText('Changed', { exact: true })).toHaveCount(0)
    await expect(panel.getByText(/via Component ref/i)).toBeVisible()
  })

  test('renders an install config update without a diff section', async ({
    page,
  }) => {
    await page.goto(DETAILS_INSTALL_CONFIG, { waitUntil: 'domcontentloaded' })
    const panel = await openPanel(page)

    await expect(
      panel.getByText('Install config', { exact: true })
    ).toBeVisible()
    await expect(panel.getByText(/stack impacts/i)).toHaveCount(0)
    await expect(panel.getByText(/component impacts/i)).toHaveCount(0)
  })

  test('shows the failure status and description', async ({ page }) => {
    await page.goto(DETAILS_FAILED, { waitUntil: 'domcontentloaded' })
    const panel = await openPanel(page)

    await expect(panel.getByText('Error', { exact: true })).toBeVisible()
  })
})

test.describe('CurrentAppBranchRun', () => {
  test('shows the applied branch, commit, and preview badge', async ({
    page,
  }) => {
    await page.goto(CURRENT_RUN_APPLIED, { waitUntil: 'domcontentloaded' })

    await expect(page.getByText('Applied app branch')).toBeVisible()
    await expect(
      page.getByText('feat/add-cache', { exact: true })
    ).toBeVisible()
  })

  test('explains when nothing has been applied', async ({ page }) => {
    await page.goto(CURRENT_RUN_EMPTY, { waitUntil: 'domcontentloaded' })

    await expect(
      page.getByText('No app branch run has been applied.')
    ).toBeVisible()
  })
})
