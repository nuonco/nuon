import { expect, test, type Page } from '@playwright/test'

const resourceCard = (page: Page, name: string) =>
  page.getByTestId('drift-resource-card').filter({
    has: page.getByRole('heading', { name, exact: true }),
  })

const openResources = async (page: Page, scenario = 'components-drifted') => {
  await page.goto(
    `/?story=views--installs--resources--${scenario}&mode=preview`,
    { waitUntil: 'domcontentloaded' }
  )
}

const approveFreshPlan = async (page: Page) => {
  await page
    .locator('#panel-root')
    .getByRole('button', { name: 'Approve plan', exact: true })
    .click()
  await page
    .getByRole('dialog')
    .getByRole('button', { name: 'Approve plan', exact: true })
    .click()
}

const expectInSync = async (page: Page, name: string, images = false) => {
  await expect(
    page
      .getByText(images ? 'Image synced to the registry' : 'Drift resolved', {
        exact: true,
      })
      .first()
  ).toBeVisible({ timeout: 10000 })
  await page.keyboard.press('Escape')
  await page.getByRole('link', { name: 'Resources', exact: true }).click()
  await page
    .getByRole('link', { name: images ? 'Images' : 'Components', exact: true })
    .click()
  // The Resources overview refreshes on the product's normal 20-second poll.
  await expect(
    page
      .getByRole('region', { name: images ? 'Synced' : 'In sync', exact: true })
      .getByRole('heading', { name, exact: true })
  ).toBeVisible({ timeout: 25000 })
  await expect(resourceCard(page, name)).toHaveCount(0)
}

test.describe('Resolve drift from Resources', () => {
  let errors: string[]
  let apiRequests: string[]

  test.beforeEach(({ page }) => {
    errors = []
    apiRequests = []
    page.on('pageerror', (error) => errors.push(error.message))
    page.on('request', (request) => {
      if (/^\/(v1|api)\//.test(new URL(request.url()).pathname))
        apiRequests.push(request.url())
    })
  })

  test.afterEach(() => {
    expect(errors).toEqual([])
    expect(apiRequests, 'Stories must not send requests to a real API').toEqual(
      []
    )
  })

  for (const [name, plan] of [
    ['api', 'Helm changes'],
    ['ledger', 'Resource changes'],
    ['webhooks', 'Kubernetes changes'],
  ]) {
    test(`plain Resolve reviews and applies a fresh ${name} plan`, async ({
      page,
    }) => {
      await openResources(page)
      await resourceCard(page, name)
        .getByRole('button', { name: 'Resolve drift', exact: true })
        .click()
      const panel = page.locator('#panel-root')
      await expect(panel.getByText(plan, { exact: true })).toBeVisible()
      await expect(
        panel.getByRole('button', { name: 'Approve plan', exact: true })
      ).toBeVisible()
      await approveFreshPlan(page)
      await expectInSync(page, name)
    })
  }

  for (const scenario of ['components-drifted', 'resolve-changed-plan']) {
    test(`${scenario} reuses approval only for an identical fresh plan`, async ({
      page,
    }) => {
      await openResources(page, scenario)
      await resourceCard(page, 'api')
        .getByRole('button', { name: 'View plan', exact: true })
        .click()
      const panel = page.getByTestId('drift-plan-panel')
      await panel
        .getByRole('button', {
          name: 'Approve plan and resolve drift',
          exact: true,
        })
        .click()
      if (scenario === 'resolve-changed-plan') {
        const freshPlan = page.locator('#panel-root')
        await expect(freshPlan.getByText(/replicas:.*4/)).toBeVisible()
        await approveFreshPlan(page)
      } else {
        await expect(
          page.getByRole('button', { name: 'Approve plan', exact: true })
        ).toHaveCount(0)
      }
      await expectInSync(page, 'api')
    })
  }

  test('respects install autoapproval without another click', async ({
    page,
  }) => {
    await openResources(page, 'resolve-auto-approved')
    await resourceCard(page, 'api')
      .getByRole('button', { name: 'Resolve drift', exact: true })
      .click()
    await expectInSync(page, 'api')
    await expect(
      page.getByRole('button', { name: 'Approve plan', exact: true })
    ).toHaveCount(0)
  })

  test('policy denial prevents apply even after approving the prior plan', async ({
    page,
  }) => {
    await openResources(page, 'resolve-policy-denied')
    await resourceCard(page, 'api')
      .getByRole('button', { name: 'View plan', exact: true })
      .click()
    await page
      .getByTestId('drift-plan-panel')
      .getByRole('button', {
        name: 'Approve plan and resolve drift',
        exact: true,
      })
      .click()
    await expect(
      page
        .getByText('Policy denied this deployment. The plan was not applied.', {
          exact: true,
        })
        .first()
    ).toBeVisible({ timeout: 10000 })
    await expect(page.getByText('Drift resolved', { exact: true })).toHaveCount(
      0
    )
    await page.keyboard.press('Escape')
    await page.getByRole('link', { name: 'Resources', exact: true }).click()
    await page.getByRole('link', { name: 'Components', exact: true }).click()
    await expect(
      resourceCard(page, 'api').getByRole('button', {
        name: 'Resolve drift',
        exact: true,
      })
    ).toBeVisible({ timeout: 25000 })
    await expect(
      resourceCard(page, 'api').getByText('In sync', { exact: true })
    ).toHaveCount(0)
  })

  test('image Resolve syncs the registry without a workload approval', async ({
    page,
  }) => {
    await openResources(page, 'images-pending-sync')
    await resourceCard(page, 'api-image')
      .getByRole('button', { name: 'Resolve drift', exact: true })
      .click()
    await expectInSync(page, 'api-image', true)
    await expect(
      page.getByRole('button', { name: 'Approve plan', exact: true })
    ).toHaveCount(0)
  })
})
