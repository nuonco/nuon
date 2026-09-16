import { expect, test } from '@playwright/test'

const MULTIPLE =
  '/?story=workflows--workflowdetails--multiple-failed-steps&mode=preview'
const SINGLE =
  '/?story=workflows--workflowdetails--single-failed-step&mode=preview'

test.describe('Workflow failed step banners', () => {
  test('summarises the error count and toggles the older errors', async ({
    page,
  }) => {
    await page.goto(MULTIPLE, { waitUntil: 'domcontentloaded' })

    await expect(page.getByText('3 steps failed')).toBeVisible()
    const toggle = page.getByRole('button', { name: /3 steps failed/ })
    await expect(toggle).toHaveAttribute('aria-expanded', 'false')
    await expect(toggle).toContainText('Show 2 more')

    await expect(page.getByText('Step deploy api-server failed')).toBeVisible()
    await expect(page.getByText('Step deploy database failed')).toBeHidden()

    await toggle.click()

    await expect(toggle).toHaveAttribute('aria-expanded', 'true')
    await expect(toggle).toContainText('Show less')
    await expect(page.getByText('Step deploy database failed')).toBeVisible()
    await expect(page.getByText('Step deploy cache failed')).toBeVisible()
  })

  test('shows no count summary for a single failed step', async ({ page }) => {
    await page.goto(SINGLE, { waitUntil: 'domcontentloaded' })

    await expect(page.getByText('Step deploy api-server failed')).toBeVisible()
    await expect(page.getByText(/steps failed/)).toBeHidden()
  })
})
