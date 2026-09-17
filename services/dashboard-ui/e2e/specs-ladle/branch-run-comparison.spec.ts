import { expect, test } from '@playwright/test'

const WITH_BASELINE =
  '/?story=branches--branchruncomparisonruns--with-baseline&mode=preview'
const FIRST_RUN =
  '/?story=branches--branchruncomparisonruns--first-run-no-baseline&mode=preview'

test.describe('BranchRunComparisonRuns behavior', () => {
  test('reads previous run first, then current run', async ({ page }) => {
    await page.goto(WITH_BASELINE, { waitUntil: 'domcontentloaded' })

    const labels = page.getByText(/^(Previous|Current) run$/)
    await expect(labels).toHaveText(['Previous run', 'Current run'])
  })

  test('a first run shows only the current run', async ({ page }) => {
    await page.goto(FIRST_RUN, { waitUntil: 'domcontentloaded' })

    await expect(page.getByText('Current run')).toBeVisible()
    await expect(page.getByText('Previous run')).toHaveCount(0)
    await expect(
      page.getByText('First run on this branch — no previous baseline to compare against.')
    ).toBeVisible()
  })
})
