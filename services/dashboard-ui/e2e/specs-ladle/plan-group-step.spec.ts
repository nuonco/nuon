import { expect, test } from '@playwright/test'

const WITH_LABELS =
  '/?story=branches--workflowstepdetail--plangroupstep--with-labels&mode=preview'

test.describe('PlanGroupStep behavior', () => {
  test('links each install name to its install page', async ({ page }) => {
    await page.goto(WITH_LABELS, { waitUntil: 'domcontentloaded' })

    await expect(page.getByRole('link', { name: 'acme-prod' })).toHaveAttribute(
      'href',
      '/org123/installs/inlacmeprod'
    )
    await expect(
      page.getByRole('link', { name: 'acme-staging' })
    ).toHaveAttribute('href', '/org123/installs/inlacmestg')
  })

  test('an install with changes still expands its diff', async ({ page }) => {
    await page.goto(WITH_LABELS, { waitUntil: 'domcontentloaded' })

    const toggle = page.getByRole('button', {
      name: 'Show plan changes for acme-prod',
    })
    await expect(toggle).toHaveAttribute('aria-expanded', 'false')

    await toggle.click()
    await expect(toggle).toHaveAttribute('aria-expanded', 'true')
    await expect(page.getByText('redis').first()).toBeVisible()

    await expect(page.getByRole('link', { name: 'acme-prod' })).toBeVisible()
  })
})
