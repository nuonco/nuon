import { expect, test, type Locator, type Page } from '@playwright/test'

const STORY =
  '/?story=playground--installs--deployments--layout-controls&mode=preview'

async function choose(page: Page, name: string, value: string) {
  const control = page.getByRole('combobox', { name, exact: true })
  await control.scrollIntoViewIfNeeded()
  await control.focus()
  await page.evaluate(
    () =>
      new Promise<void>((resolve) => {
        requestAnimationFrame(() => requestAnimationFrame(() => resolve()))
      })
  )
  await control.click()
  await expect(control).toHaveAttribute('aria-expanded', 'true')
  await page.getByRole('option', { name: value, exact: true }).click()
  await expect(control).toHaveText(value)
}

async function layoutMetrics(page: Page, row?: Locator) {
  return (row ?? page.getByRole('article')).evaluate((element) => {
    const identity = element.querySelector('.density-identity')!
    const action = element.querySelector('.density-action')!
    const header = element.querySelector('.density-header')!
    const bounds = element.getBoundingClientRect()
    return {
      left: bounds.left,
      right: bounds.right,
      identityWidth: identity.getBoundingClientRect().width,
      identityRight: identity.getBoundingClientRect().right,
      headerTop: header.getBoundingClientRect().top,
      headerBottom: header.getBoundingClientRect().bottom,
      currentStepInIdentity: !!identity.querySelector(
        '[aria-label="Current step"]'
      ),
      progressInIdentity: !!identity.querySelector(
        '[aria-label^="Workflow steps:"]'
      ),
      actionRight: action.getBoundingClientRect().right,
      dividerWidths: [action].map((node) =>
        parseFloat(getComputedStyle(node).borderLeftWidth)
      ),
    }
  })
}

for (const width of [390, 768, 1024, 1440, 1920, 2560]) {
  test(`layout contract holds across content variations at ${width}px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 1100 })
    const errors: string[] = []
    const apiRequests: string[] = []
    page.on('pageerror', (error) => errors.push(error.message))
    page.on('request', (request) => {
      if (/^\/(v1|api)\//.test(new URL(request.url()).pathname))
        apiRequests.push(request.url())
    })
    await page.goto(STORY, { waitUntil: 'domcontentloaded' })
    await expect(
      page.getByRole('heading', { name: 'Deployment row layout' })
    ).toBeVisible()
    await choose(page, 'Workflow list', 'Single workflow')
    for (const situation of [
      'Success',
      'Failure before apply',
      'Partial rollout',
      'Running',
      'Pending approval',
    ]) {
      await choose(page, 'Workflow situation', situation)
      const row = page.getByRole('article')
      await expect(
        row.getByRole('group', { name: 'Resource outcomes', exact: true })
      ).toHaveCount(0)
      await page
        .getByRole('textbox', { name: 'Deployment title', exact: true })
        .fill(
          situation === 'Success'
            ? 'Provision the production install and update all requested resources safely'
            : 'Provisioned install'
        )
      if (
        situation === 'Failure before apply' ||
        situation === 'Partial rollout'
      ) {
        const reason =
          'The workflow stopped because the cloud API could not be reached. No further resources were updated. Restore connectivity and retry the failed step.'
        await page
          .getByRole('textbox', {
            name: 'Workflow description / failure reason',
            exact: true,
          })
          .fill(reason)
        await expect(row.getByText(reason, { exact: true })).toBeVisible()
        await expect(
          row.getByRole('group', { name: /Workflow steps:/ })
        ).toHaveCount(0)
      }
      const metrics = await layoutMetrics(page)
      expect(metrics.right).toBeLessThanOrEqual(width)
      expect(metrics.actionRight).toBeLessThanOrEqual(metrics.right + 1)
      expect(metrics.dividerWidths).toEqual(width >= 1024 ? [1] : [0])
      if (width >= 1024) {
        expect(metrics.identityRight).toBeLessThan(metrics.actionRight)
      }
      const ongoing =
        situation === 'Running' || situation === 'Pending approval'
      expect(metrics.currentStepInIdentity).toBe(true)
      await expect(
        row.getByRole('group', { name: 'Next step', exact: true })
      ).toBeVisible()
      expect(metrics.progressInIdentity).toBe(ongoing)
    }
    expect(errors).toEqual([])
    expect(apiRequests).toEqual([])
  })
}

test('column widths stay fixed across states and long content', async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 1100 })
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  await choose(page, 'Workflow list', 'Single workflow')
  await choose(page, 'Workflow situation', 'Success')
  await expect(page.getByRole('article')).toBeVisible()
  const success = await layoutMetrics(page)
  await choose(page, 'Workflow situation', 'Failure before apply')
  await page
    .getByRole('textbox', {
      name: 'Workflow description / failure reason',
      exact: true,
    })
    .fill(
      'The runner could not connect to the cloud API while validating the updated policy. Restore network connectivity before retrying this workflow.'
    )
  const failure = await layoutMetrics(page)
  expect(failure.identityWidth).toBeCloseTo(success.identityWidth, 0)
  await page
    .getByRole('textbox', { name: 'Deployment title', exact: true })
    .fill(
      'A much longer deployment title that must stay inside the workflow row'
    )
  const titled = await layoutMetrics(page)
  expect(titled.identityWidth).toBeCloseTo(success.identityWidth, 0)
  await choose(page, 'Workflow situation', 'Success')
  const restored = await layoutMetrics(page)
  expect(restored.identityWidth).toBeCloseTo(success.identityWidth, 0)
})

test('width, text, theme and details controls work without API access', async ({
  page,
}) => {
  await page.setViewportSize({ width: 1920, height: 1100 })
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  await choose(page, 'Workflow list', 'Single workflow')
  await expect(page.getByRole('article')).toBeVisible()
  await choose(page, 'Preview width', '390px')
  await expect(page.getByRole('status')).toContainText(
    'Actual preview width: 390px'
  )
  expect((await layoutMetrics(page)).dividerWidths).toEqual([0])
  await choose(page, 'Preview width', 'Available canvas width')
  for (const theme of ['Light', 'Dark']) {
    await choose(page, 'Theme', theme)
    await expect(page.locator('html')).toHaveAttribute(
      'data-theme',
      theme.toLowerCase()
    )
  }
  await page
    .getByRole('textbox', { name: 'Deployment title', exact: true })
    .fill('Custom deployment')
  const row = page.getByRole('article', {
    name: 'Custom deployment',
    exact: true,
  })
  await row.getByRole('button', { name: 'View details', exact: true }).focus()
  await page.keyboard.press('Enter')
  const panel = page.getByRole('complementary', {
    name: 'Layout preview details',
  })
  await expect(panel).toBeVisible()
  await expect(panel).toContainText('Fixture data only')
  await page.keyboard.press('Escape')
  await expect(panel).not.toBeVisible()
  await page.getByRole('button', { name: 'Reset preview', exact: true }).click()
  await expect(
    page.getByRole('article', { name: 'Provisioned install', exact: true })
  ).toBeVisible()
  await expect(page.getByRole('article')).toHaveCount(5)
})

for (const width of [390, 1024, 1920]) {
  test(`mixed deployment list preserves the layout hierarchy at ${width}px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 1100 })
    await page.goto(STORY, { waitUntil: 'domcontentloaded' })
    const preview = page.getByRole('main', {
      name: 'Deployments layout preview',
    })
    await expect(
      preview.getByRole('heading', { name: 'Deployments', exact: true })
    ).toBeVisible()
    await expect(preview.getByRole('article')).toHaveCount(5)
    await expect(preview.getByText(/Images/)).toHaveCount(0)
    for (const label of ['Status', 'Type', 'Resource', 'Date'])
      await expect(
        preview.getByRole('button', { name: label, exact: true })
      ).toBeVisible()
    const completed = await layoutMetrics(
      page,
      preview.getByRole('article', {
        name: 'Roll out template v12',
        exact: true,
      })
    )
    for (const row of await preview.getByRole('article').all()) {
      const metrics = await layoutMetrics(page, row)
      expect(metrics.right).toBeLessThanOrEqual(width)
      expect(metrics.actionRight).toBeLessThanOrEqual(metrics.right + 1)
      if (width >= 1024) {
        expect(metrics.identityWidth).toBeCloseTo(completed.identityWidth, 0)
        expect(metrics.dividerWidths).toEqual([1])
      }
    }
    await choose(page, 'Workflow list', '3 workflows')
    await expect(preview.getByRole('article')).toHaveCount(3)
    await choose(page, 'Workflow list', 'Single workflow')
    await expect(preview.getByRole('article')).toHaveCount(1)
  })
}

test('deployment search and all four filters operate on the mixed fixtures', async ({
  page,
}) => {
  const errors: string[] = []
  const apiRequests: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  page.on('request', (request) => {
    if (/^\/(v1|api)\//.test(new URL(request.url()).pathname))
      apiRequests.push(request.url())
  })
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  const preview = page.getByRole('main', { name: 'Deployments layout preview' })
  const rows = preview.getByRole('article')
  await expect(rows).toHaveCount(5)
  const clear = preview.getByRole('button', {
    name: 'Clear filters',
    exact: true,
  })
  await preview
    .getByRole('textbox', { name: 'Search deployments' })
    .fill('readiness')
  await expect(rows).toHaveCount(1)
  await expect(rows).toHaveAttribute('aria-label', 'Roll out template v13')
  await clear.click()

  await preview.getByRole('button', { name: 'Status', exact: true }).click()
  await page.getByRole('button', { name: 'Failed Only', exact: true }).click()
  await expect(rows).toHaveCount(2)
  await preview
    .getByRole('heading', { name: 'Deployments', exact: true })
    .click()
  await clear.click()

  await preview.getByRole('button', { name: 'Type', exact: true }).click()
  await page
    .getByRole('button', { name: 'Stack update Only', exact: true })
    .click()
  await expect(rows).toHaveCount(1)
  await expect(rows).toHaveAttribute('aria-label', 'Update stack permissions')
  await preview
    .getByRole('heading', { name: 'Deployments', exact: true })
    .click()
  await clear.click()

  await preview.getByRole('button', { name: 'Resource', exact: true }).click()
  await page.getByRole('radio', { name: 'components', exact: true }).check()
  await expect(rows).toHaveCount(4)
  await preview
    .getByRole('heading', { name: 'Deployments', exact: true })
    .click()
  await clear.click()

  await preview.getByRole('button', { name: 'Date', exact: true }).click()
  await page.getByRole('radio', { name: 'Last 24 hours', exact: true }).check()
  await expect(rows).toHaveCount(3)
  await preview
    .getByRole('heading', { name: 'Deployments', exact: true })
    .click()
  await preview
    .getByRole('textbox', { name: 'Search deployments' })
    .fill('not a deployment')
  await expect(rows).toHaveCount(0)
  await expect(
    preview.getByText('No deployments found', { exact: true })
  ).toBeVisible()
  await clear.click()
  await expect(rows).toHaveCount(5)
  const selected = rows.first()
  await selected
    .getByRole('button', { name: 'View details', exact: true })
    .click()
  const panel = page.getByRole('complementary', {
    name: 'Layout preview details',
  })
  await expect(panel).toContainText('Deploy components')
  await page.keyboard.press('Escape')
  await expect(panel).not.toBeVisible()
  expect(errors).toEqual([])
  expect(apiRequests).toEqual([])
})

for (const width of [390, 768, 1024, 1440, 1920, 2560]) {
  test(`layout A uses the first column for steps and progress at ${width}px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 1100 })
    await page.goto(STORY, { waitUntil: 'domcontentloaded' })
    await choose(page, 'Workflow list', 'Single workflow')
    const row = page.getByRole('article')
    for (const situation of ['Running', 'Pending approval']) {
      await choose(page, 'Workflow situation', situation)
      const current = row.getByRole('group', {
        name: 'Current step',
        exact: true,
      })
      const next = row.getByRole('group', { name: 'Next step', exact: true })
      await expect(
        current.getByText('Current step', { exact: true })
      ).toBeVisible()
      await expect(next.getByText('Up next', { exact: true })).toBeVisible()
      await expect(
        row.getByText('1/3 steps complete', { exact: true })
      ).toBeVisible()
      const metadata = row.locator('.density-run-meta')
      await expect(
        metadata.getByText('In progress', { exact: true })
      ).toHaveCount(0)
      await expect(
        metadata.getByText(
          situation === 'Running' ? 'Running now' : 'Pending approval',
          { exact: true }
        )
      ).toHaveCount(1)
      const geometry = await row.evaluate((element) => {
        const box = (node: Element) => {
          const { left, right, top, bottom, width } =
            node.getBoundingClientRect()
          return { left, right, top, bottom, width }
        }
        const identity = element.querySelector('.density-identity')!
        const progress = identity.querySelector(
          '[aria-label^="Workflow steps:"]'
        )!
        const current = identity.querySelector('[aria-label="Current step"]')!
        const next = identity.querySelector('[aria-label="Next step"]')!
        return {
          identity: box(identity),
          current: box(current),
          next: box(next),
          nextDivider: parseFloat(getComputedStyle(next).borderLeftWidth),
          title: box(identity.querySelector('.density-run-heading > span')!),
          metadata: box(identity.querySelector('.density-run-meta')!),
          bar: box(progress.firstElementChild!),
          count: box(progress.lastElementChild!),
          segments: progress.firstElementChild!.children.length,
        }
      })
      if (geometry.identity.width >= 384) {
        expect(geometry.next.left).toBeGreaterThan(geometry.current.right)
        expect(geometry.next.top).toBeCloseTo(geometry.current.top, 0)
        expect(geometry.nextDivider).toBe(1)
      } else {
        expect(geometry.next.top).toBeGreaterThan(geometry.current.bottom)
        expect(geometry.next.left).toBeCloseTo(geometry.current.left, 0)
        expect(geometry.nextDivider).toBe(0)
      }
      if (geometry.identity.width >= 512) {
        expect(geometry.metadata.left).toBeGreaterThan(geometry.title.right)
        expect(geometry.metadata.right).toBeCloseTo(geometry.identity.right, 0)
      } else {
        expect(geometry.metadata.top).toBeGreaterThan(geometry.title.bottom)
      }
      expect(geometry.segments).toBe(3)
      expect(geometry.bar.left).toBeCloseTo(geometry.identity.left, 0)
      expect(geometry.count.right).toBeCloseTo(geometry.identity.right, 0)
      expect(geometry.count.left - geometry.bar.right).toBeCloseTo(12, 0)
      expect(geometry.bar.top).toBeGreaterThanOrEqual(
        Math.max(geometry.current.bottom, geometry.next.bottom)
      )
    }
    for (const [situation, currentName, nextName] of [
      ['Failure before apply', 'Plan stack permissions', 'Apply stack policy'],
      ['Partial rollout', 'Verify component readiness', 'Finalize deployment'],
      ['Success', 'Verify install health', 'No remaining steps'],
    ]) {
      await choose(page, 'Workflow situation', situation)
      const current = row.getByRole('group', {
        name: 'Current step',
        exact: true,
      })
      const next = row.getByRole('group', { name: 'Next step', exact: true })
      await expect(
        current.getByText('Current step', { exact: true })
      ).toBeVisible()
      await expect(
        current.getByText(currentName, { exact: true })
      ).toBeVisible()
      await expect(next.getByText('Up next', { exact: true })).toBeVisible()
      await expect(next.getByText(nextName, { exact: true })).toBeVisible()
      await expect(
        row.getByRole('group', { name: /Workflow steps:/ })
      ).toHaveCount(0)
      if (situation !== 'Success') {
        await expect(current.getByText('Failed', { exact: true })).toBeVisible()
        await expect(
          next.getByText('Not started — deployment stopped', { exact: true })
        ).toBeVisible()
        await expect(
          row.locator('.density-identity > span.text-base')
        ).toBeVisible()
      }
      const identityBox = await row.locator('.density-identity').boundingBox()
      const currentBox = await current.boundingBox()
      const nextBox = await next.boundingBox()
      expect(nextBox!.x + nextBox!.width).toBeLessThanOrEqual(
        identityBox!.x + identityBox!.width + 1
      )
      if (identityBox!.width >= 384) {
        expect(nextBox!.x).toBeGreaterThan(currentBox!.x + currentBox!.width)
        expect(nextBox!.y).toBeCloseTo(currentBox!.y, 0)
      } else {
        expect(nextBox!.y).toBeGreaterThan(currentBox!.y + currentBox!.height)
      }
    }
  })
}
