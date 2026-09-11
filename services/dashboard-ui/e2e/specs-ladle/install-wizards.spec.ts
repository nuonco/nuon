import { expect, test } from '@playwright/test'

const SETUP_STORY =
  '/?story=installs--install-wizards--setup-customer-managed&mode=preview'
const DEPROVISION_STORY =
  '/?story=installs--install-wizards--deprovision-confirm&mode=preview'

test.describe('Install wizard behavior', () => {
  test('setup gates details and Nuon-managed stack access', async ({
    page,
  }) => {
    await page.goto(SETUP_STORY, { waitUntil: 'domcontentloaded' })

    await expect(page.getByRole('button', { name: 'Back' })).toBeDisabled()
    await page.getByRole('button', { name: 'Enter install details' }).click()

    const installName = page.getByLabel('Install name')
    const region = page.getByLabel('AWS region')
    const chooseStack = page.getByRole('button', {
      name: 'Choose stack management',
    })

    await installName.clear()
    await region.clear()
    await expect(chooseStack).toBeDisabled()

    await installName.fill('production')
    await region.fill('us-west-2')
    await expect(chooseStack).toBeEnabled()
    await chooseStack.click()

    await page.getByLabel('Nuon manages the stack').check()
    const provision = page.getByRole('button', {
      name: 'Provision install',
    })
    await expect(provision).toBeDisabled()

    await page
      .getByLabel('IAM role ARN')
      .fill('arn:aws:iam::123456789012:role/nuon-stack-manager')
    await expect(provision).toBeEnabled()
  })

  test('deprovision requires the exact install name', async ({ page }) => {
    await page.goto(DEPROVISION_STORY, { waitUntil: 'domcontentloaded' })

    const confirmation = page.getByLabel('Type production to confirm')
    const deprovision = page.getByRole('button', {
      name: 'Deprovision install',
    })

    await expect(deprovision).toBeDisabled()
    await confirmation.fill('Production')
    await expect(deprovision).toBeDisabled()
    await confirmation.fill('production')
    await expect(deprovision).toBeEnabled()
  })
})
