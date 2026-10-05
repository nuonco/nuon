import { expect, test, type Locator, type Page } from '@playwright/test'

const STORY =
  '/?story=views--installs--deployment-details--layout-controls&mode=preview'

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
    const resources = element.querySelector('.density-resources')!
    const action = element.querySelector('.density-action')!
    const header = element.querySelector('.density-header')!
    const bounds = element.getBoundingClientRect()
    const cells = [
      ...resources.querySelectorAll<HTMLSpanElement>('span[role="group"]'),
    ].map((cell) => {
      const label = cell.lastElementChild!
      const icon = cell.querySelector('svg')!
      const range = document.createRange()
      range.selectNodeContents(label)
      const text = [...range.getClientRects()].map((rect) => ({
        left: rect.left,
        right: rect.right,
        top: rect.top,
        bottom: rect.bottom,
      }))
      const fragmentedWords: string[] = []
      const walker = document.createTreeWalker(label, NodeFilter.SHOW_TEXT)
      while (walker.nextNode()) {
        const node = walker.currentNode
        for (const word of (node.textContent ?? '').matchAll(/\S+/g)) {
          range.setStart(node, word.index)
          range.setEnd(node, word.index + word[0].length)
          if (range.getClientRects().length > 1) fragmentedWords.push(word[0])
        }
      }
      const cellBounds = cell.getBoundingClientRect()
      const iconBounds = icon.getBoundingClientRect()
      return {
        font: parseFloat(getComputedStyle(label).fontSize),
        iconSize: iconBounds.width,
        left: cellBounds.left,
        right: cellBounds.right,
        top: cellBounds.top,
        bottom: cellBounds.bottom,
        iconRight: iconBounds.right,
        text,
        fragmentedWords,
      }
    })
    return {
      left: bounds.left,
      right: bounds.right,
      identityWidth: identity.getBoundingClientRect().width,
      identityRight: identity.getBoundingClientRect().right,
      resourcesWidth: resources.getBoundingClientRect().width,
      resourcesLeft: resources.getBoundingClientRect().left,
      resourcesTop: resources.getBoundingClientRect().top,
      resourcesBottom: resources.getBoundingClientRect().bottom,
      headerTop: header.getBoundingClientRect().top,
      headerBottom: header.getBoundingClientRect().bottom,
      currentStepInIdentity: !!identity.querySelector(
        '[aria-label="Current step"]'
      ),
      progressInIdentity: !!identity.querySelector(
        '[aria-label^="Workflow steps:"]'
      ),
      actionRight: action.getBoundingClientRect().right,
      dividerWidths: [resources, action].map((node) =>
        parseFloat(getComputedStyle(node).borderLeftWidth)
      ),
      cells,
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
    for (const [index, situation] of [
      'Success',
      'Failure before apply',
      'Partial rollout',
      'Running',
      'Pending approval',
    ].entries()) {
      await choose(page, 'Workflow situation', situation)
      const count = (index % 3) + 1
      await choose(
        page,
        'Resource categories',
        `${count} ${count === 1 ? 'category' : 'categories'}`
      )
      const row = page.getByRole('article')
      await expect(
        row
          .getByRole('group', { name: 'Resource outcomes', exact: true })
          .locator(':scope > span[role="group"]')
      ).toHaveCount(count)
      await page
        .getByRole('textbox', { name: 'Deployment title', exact: true })
        .fill(
          index % 2
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
      expect(metrics.dividerWidths).toEqual(width >= 1024 ? [1, 1] : [0, 0])
      if (width >= 1024) {
        expect(metrics.identityWidth).toBeGreaterThan(metrics.resourcesWidth)
        expect(metrics.identityRight).toBeLessThan(metrics.resourcesLeft)
      }
      const ongoing =
        situation === 'Running' || situation === 'Pending approval'
      expect(metrics.currentStepInIdentity).toBe(true)
      await expect(
        row.getByRole('group', { name: 'Next step', exact: true })
      ).toBeVisible()
      expect(metrics.progressInIdentity).toBe(ongoing)
      for (const [cellIndex, cell] of metrics.cells.entries()) {
        expect(cell.fragmentedWords).toEqual([])
        expect(cell.font).toBeGreaterThanOrEqual(14)
        expect(cell.font).toBeLessThanOrEqual(18)
        expect(cell.iconSize).toBeGreaterThanOrEqual(14)
        expect(cell.iconSize).toBeLessThanOrEqual(22)
        expect(cell.right).toBeLessThanOrEqual(metrics.right + 1)
        if (cellIndex) {
          const previous = metrics.cells[cellIndex - 1]
          expect(cell.left).toBeCloseTo(previous.left, 0)
          expect(cell.top).toBeGreaterThanOrEqual(previous.bottom + 11)
        }
        for (const rect of cell.text) {
          expect(rect.left).toBeGreaterThanOrEqual(cell.iconRight - 1)
          expect(rect.right).toBeLessThanOrEqual(cell.right + 1)
        }
      }
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
  expect(failure.resourcesLeft).toBeCloseTo(success.resourcesLeft, 0)
  expect(failure.resourcesWidth).toBeCloseTo(success.resourcesWidth, 0)
  for (const [index, cell] of failure.cells.entries())
    expect(cell.font).toEqual(success.cells[index].font)
  await page
    .getByRole('textbox', { name: 'Deployment title', exact: true })
    .fill(
      'A much longer deployment title that must not move the resource column'
    )
  await choose(page, 'Resource categories', '1 category')
  const single = await layoutMetrics(page)
  expect(single.identityWidth).toBeCloseTo(success.identityWidth, 0)
  expect(single.resourcesLeft).toBeCloseTo(success.resourcesLeft, 0)
  expect(single.resourcesWidth).toBeCloseTo(success.resourcesWidth, 0)
  await choose(page, 'Resource categories', '3 categories')
  await choose(page, 'Workflow situation', 'Success')
  const restored = await layoutMetrics(page)
  expect(restored.cells.map((cell) => cell.font)).toEqual(
    success.cells.map((cell) => cell.font)
  )
})

for (const width of [390, 768, 1024, 1440, 1920, 2560]) {
  test(`column balance preserves vertical outcomes and shares space proportionally at ${width}px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 1100 })
    await page.goto(STORY, { waitUntil: 'domcontentloaded' })
    const rows = page.getByRole('article')
    await expect(rows).toHaveCount(5)
    await expect(
      page.getByRole('combobox', { name: 'Column balance', exact: true })
    ).toHaveText('Resources 40% / Workflow 60%')
    const original = await layoutMetrics(page, rows.first())
    if (width >= 1024)
      expect(
        original.resourcesWidth /
          (original.identityWidth + original.resourcesWidth)
      ).toBeCloseTo(0.4, 3)
    for (const [label, resourceShare] of [
      ['Resources 35% / Workflow 65%', 0.35],
      ['Resources 40% / Workflow 60%', 0.4],
      ['Resources 45% / Workflow 55%', 0.45],
    ] as const) {
      await choose(page, 'Column balance', label)
      const first = await layoutMetrics(page, rows.first())
      for (const row of await rows.all()) {
        const metrics = await layoutMetrics(page, row)
        expect(metrics.identityWidth).toBeCloseTo(first.identityWidth, 0)
        expect(metrics.resourcesLeft).toBeCloseTo(first.resourcesLeft, 0)
        expect(metrics.resourcesWidth).toBeCloseTo(first.resourcesWidth, 0)
        expect(metrics.actionRight).toBeCloseTo(original.actionRight, 0)
        for (const [index, cell] of metrics.cells.entries()) {
          expect(cell.fragmentedWords).toEqual([])
          expect(cell.right).toBeLessThanOrEqual(metrics.right + 1)
          if (index) {
            expect(cell.left).toBeCloseTo(metrics.cells[index - 1].left, 0)
            expect(cell.top).toBeGreaterThan(metrics.cells[index - 1].bottom)
          }
        }
        if (width >= 1024) {
          expect(metrics.identityWidth).toBeGreaterThan(metrics.resourcesWidth)
          expect(
            metrics.resourcesWidth /
              (metrics.identityWidth + metrics.resourcesWidth)
          ).toBeCloseTo(resourceShare, 3)
          expect(metrics.dividerWidths).toEqual([1, 1])
          expect(metrics.resourcesLeft - metrics.identityRight).toBeCloseTo(
            24,
            0
          )
          expect(metrics.cells[0].left - metrics.resourcesLeft).toBeCloseTo(
            25,
            0
          )
          expect(metrics.resourcesTop).toBeCloseTo(metrics.headerTop, 0)
          expect(metrics.resourcesBottom).toBeCloseTo(metrics.headerBottom, 0)
        } else {
          expect(metrics.resourcesWidth).toBeCloseTo(original.resourcesWidth, 0)
        }
      }
      if (width >= 1440)
        expect(first.identityWidth + first.resourcesWidth).toBeCloseTo(
          original.identityWidth + original.resourcesWidth,
          0
        )
    }
    await page
      .getByRole('button', { name: 'Reset preview', exact: true })
      .click()
    await expect(
      page.getByRole('combobox', { name: 'Column balance', exact: true })
    ).toHaveText('Resources 40% / Workflow 60%')
    expect((await layoutMetrics(page, rows.first())).identityWidth).toBeCloseTo(
      original.identityWidth,
      0
    )
  })
}

test('width, sizing, text, theme and details controls work without API access', async ({
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
  expect((await layoutMetrics(page)).dividerWidths).toEqual([0, 0])
  await choose(page, 'Preview width', 'Available canvas width')
  await choose(page, 'Outcome sizing', 'Compact (14px / 14px icons)')
  expect((await layoutMetrics(page)).cells.map((cell) => cell.font)).toEqual([
    14, 14, 14,
  ])
  await choose(page, 'Outcome sizing', 'Comfortable (18px / 22px icons)')
  expect((await layoutMetrics(page)).cells.map((cell) => cell.font)).toEqual([
    18, 18, 18,
  ])
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
  await expect(
    page.getByRole('combobox', { name: 'Outcome sizing', exact: true })
  ).toHaveText('Automatic (fit each sector)')
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
      for (const cell of metrics.cells) expect(cell.fragmentedWords).toEqual([])
      if (width >= 1024) {
        expect(metrics.identityWidth).toBeGreaterThan(metrics.resourcesWidth)
        expect(metrics.identityWidth).toBeCloseTo(completed.identityWidth, 0)
        expect(metrics.resourcesLeft).toBeCloseTo(completed.resourcesLeft, 0)
        expect(metrics.resourcesWidth).toBeCloseTo(completed.resourcesWidth, 0)
        expect(metrics.dividerWidths).toEqual([1, 1])
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
