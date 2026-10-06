import { expect, test } from '@playwright/test'

const STORY =
  '/?story=playground--installs--deployments--interactive-sandbox&mode=preview'

test('deployment stories separate routed views, feature scenarios and playground controls', async ({
  request,
}) => {
  const response = await request.get('/meta.json')
  expect(response.ok()).toBe(true)
  const {
    stories,
  }: {
    stories: Record<string, { filePath: string; meta: Record<string, unknown> }>
  } = await response.json()

  for (const [prefix, names, filePath] of [
    [
      'features--installs--deployment-details',
      [
        'running-components',
        'components-not-started',
        'pending-approval',
        'partial-failure',
        'stack-recovery',
        'completed',
      ],
      'client/components/installs/DeploymentDetail/DeploymentDetail.stories.tsx',
    ],
    [
      'playground--installs--deployments',
      ['interactive-sandbox', 'layout-controls'],
      'client/components/playground/installs/Deployments/Deployments.stories.tsx',
    ],
    [
      'views--installs--deployment-details',
      ['in-progress', 'awaiting-approval', 'succeeded', 'failed'],
      'client/views/install/DeploymentDetail.stories.tsx',
    ],
    [
      'views--installs--deployments',
      ['loading', 'empty', 'results', 'filtered'],
      'client/views/install/Deployments.stories.tsx',
    ],
  ] as const) {
    const ids = Object.keys(stories).filter((id) =>
      id.startsWith(`${prefix}--`)
    )
    expect(ids.sort()).toEqual(names.map((name) => `${prefix}--${name}`).sort())
    for (const id of ids) {
      expect(stories[id].filePath).toBe(filePath)
      expect(stories[id].meta).toEqual(
        prefix.startsWith('views--')
          ? { fullBleed: true, installViews: true }
          : {}
      )
    }
  }
})

test('focused scenarios have shareable URLs and preserve their scope through details and filters', async ({
  page,
}) => {
  const scenarios = [
    {
      story: 'running-components',
      label: 'Running components',
      titles: ['Deploy api + worker'],
    },
    {
      story: 'components-not-started',
      label: 'Components pending',
      titles: ['Roll out template v14'],
    },
    {
      story: 'pending-approval',
      label: 'Pending approval',
      titles: ['Roll out template v15'],
    },
    {
      story: 'partial-failure',
      label: 'Partial failure',
      titles: ['Roll out template v13'],
    },
    {
      story: 'stack-recovery',
      label: 'Stack failure and recovery',
      titles: ['Update stack permissions', 'Update stack permissions'],
    },
    {
      story: 'completed',
      label: 'Completed rollout',
      titles: ['Roll out template v12'],
    },
  ]
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  for (const { story, label, titles } of scenarios) {
    await page.getByRole('combobox', { name: 'Preview scenario' }).click()
    await page.getByRole('option', { name: label, exact: true }).click()
    await expect(page).toHaveURL(
      (url) =>
        url.searchParams.get('story') ===
          `features--installs--deployment-details--${story}` &&
        url.searchParams.get('mode') === 'preview'
    )
    await expect(
      page.getByRole('combobox', { name: 'Preview scenario' })
    ).toHaveText(label)
    await expect(page.getByRole('article')).toHaveCount(titles.length)
    for (const [index, title] of titles.entries()) {
      await expect(page.getByRole('article').nth(index)).toHaveAttribute(
        'aria-label',
        title
      )
    }
    const search = page.getByRole('textbox', { name: 'Search deployments' })
    await search.fill('does not exist')
    await page
      .getByRole('button', { name: 'Clear filters', exact: true })
      .click()
    await expect(page.getByRole('article')).toHaveCount(titles.length)
    await page
      .getByRole('article')
      .first()
      .getByRole('button', { name: 'View details' })
      .click()
    const panel = page.getByRole('complementary', {
      name: 'Deployment details',
    })
    await panel.getByRole('link', { name: 'Open full page' }).click()
    const fullPage = page.getByRole('main', { name: 'Deployment full page' })
    await expect(
      fullPage.getByRole('heading', { name: titles[0], exact: true })
    ).toBeVisible()
    await page.getByRole('link', { name: 'Back to deployments' }).click()
    await expect(page.getByRole('article')).toHaveCount(titles.length)
    await page
      .getByRole('link', { name: 'Preview an existing full-page link' })
      .click()
    await expect(
      fullPage.getByRole('heading', { name: titles[0], exact: true })
    ).toBeVisible()
    await page.getByRole('link', { name: 'Back to deployments' }).click()
  }
  await page.reload({ waitUntil: 'domcontentloaded' })
  await expect(page.getByRole('article')).toHaveCount(1)
  await expect(page.getByRole('article')).toHaveAttribute(
    'aria-label',
    'Roll out template v12'
  )
  await page.getByRole('combobox', { name: 'Preview scenario' }).click()
  await page
    .getByRole('option', { name: 'All deployments', exact: true })
    .click()
  await expect(page).toHaveURL(
    (url) =>
      url.searchParams.get('story') ===
        'playground--installs--deployments--interactive-sandbox' &&
      url.searchParams.get('mode') === 'preview'
  )
  await expect(page.getByRole('article')).toHaveCount(7)
  expect(errors).toEqual([])
})

for (const width of [1440, 1920, 2560, 390]) {
  test(`shared deployment rows use aligned sectors across the available width at ${width}px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 1000 })
    await page.goto(STORY, { waitUntil: 'domcontentloaded' })
    const row = page.getByRole('article', {
      name: 'Roll out template v12',
      exact: true,
    })
    await expect(row).toBeVisible()
    const availableWidth = await row.evaluate((element) => {
      const section = element.closest('section')!
      const style = getComputedStyle(section)
      return (
        section.getBoundingClientRect().width -
        parseFloat(style.paddingLeft) -
        parseFloat(style.paddingRight)
      )
    })
    const rowBox = (await row.boundingBox())!
    const identityBox = (await row.locator('.density-identity').boundingBox())!
    const resourcesBox = (await row
      .locator('.density-resources')
      .boundingBox())!
    const outcomesBox = (await row
      .getByRole('group', { name: 'Resource outcomes', exact: true })
      .boundingBox())!
    const actionBox = (await row
      .getByRole('button', { name: 'View details' })
      .boundingBox())!
    const dividerWidths = await row.evaluate((element) => {
      const outcomes = element.querySelector(
        '[aria-label="Resource outcomes"]'
      )!
      const action = element.querySelector('button')!
      return [outcomes, action].map((node) =>
        parseFloat(getComputedStyle(node.parentElement!).borderLeftWidth)
      )
    })
    const details = row.getByRole('button', { name: 'View details' })
    expect(
      await details.evaluate(
        (element) => getComputedStyle(element).backgroundColor
      )
    ).not.toBe('rgba(0, 0, 0, 0)')
    await expect(details.locator('svg')).toHaveCount(1)

    expect(Math.abs(rowBox.width - availableWidth)).toBeLessThan(2)
    expect(actionBox.x + actionBox.width).toBeLessThanOrEqual(width)
    if (width >= 1440) {
      expect(dividerWidths).toEqual([1, 1])
      expect(
        resourcesBox.width / (identityBox.width + resourcesBox.width)
      ).toBeCloseTo(0.4, 3)
      expect(resourcesBox.x - (identityBox.x + identityBox.width)).toBeCloseTo(
        24,
        0
      )
      expect(outcomesBox.x - resourcesBox.x).toBeCloseTo(25, 0)
      expect(actionBox.x).toBeGreaterThanOrEqual(
        outcomesBox.x + outcomesBox.width
      )
    } else {
      expect(dividerWidths).toEqual([0, 0])
      expect(outcomesBox.y).toBeGreaterThanOrEqual(
        identityBox.y + identityBox.height
      )
      expect(actionBox.y).toBeGreaterThanOrEqual(
        outcomesBox.y + outcomesBox.height
      )
      expect(outcomesBox.x + outcomesBox.width).toBeLessThanOrEqual(width)
    }
    const categories = row
      .getByRole('group', { name: 'Resource outcomes', exact: true })
      .locator(':scope > span[role="group"]')
    for (let index = 1; index < (await categories.count()); index++) {
      const previous = (await categories.nth(index - 1).boundingBox())!
      const current = (await categories.nth(index).boundingBox())!
      expect(current.x).toBeCloseTo(previous.x, 0)
      expect(current.y).toBeGreaterThanOrEqual(
        previous.y + previous.height + 11
      )
    }
    for (const label of ['Current step', 'Next step']) {
      await expect(
        row.getByRole('group', { name: label, exact: true })
      ).toBeVisible()
    }
    const failed = page
      .getByRole('article', {
        name: 'Update stack permissions',
        exact: true,
      })
      .last()
    const failureReason = failed.getByText(
      'Validate stack policy failed — invalid resource ARN',
      { exact: true }
    )
    await expect(failureReason).toHaveCount(1)
    const failureBox = (await failureReason.boundingBox())!
    const failedIdentityBox = (await failed
      .locator('.density-identity')
      .boundingBox())!
    expect(failureBox.x).toBeCloseTo(failedIdentityBox.x, 0)
    expect(failureBox.x + failureBox.width).toBeLessThanOrEqual(
      failedIdentityBox.x + failedIdentityBox.width + 1
    )
    await expect(
      failed.getByRole('group', { name: 'Current step', exact: true })
    ).toContainText('Validate stack policy')
    await expect(
      failed.getByRole('group', { name: 'Next step', exact: true })
    ).toContainText('No remaining steps')
    if (width >= 1440) {
      const failedResourcesBox = (await failed
        .locator('.density-resources')
        .boundingBox())!
      expect(failedResourcesBox.x).toBeCloseTo(resourcesBox.x, 0)
      expect(failedResourcesBox.width).toBeCloseTo(resourcesBox.width, 0)
    }
    await details.focus()
    await page.keyboard.press('Enter')
    await expect(
      page.getByRole('complementary', { name: 'Deployment details' })
    ).toBeVisible()
    await page.keyboard.press('Escape')
    await expect(
      page.getByRole('complementary', { name: 'Deployment details' })
    ).not.toBeVisible()
  })
}

test('compact feed opens a panel, preserves selected tab in full page, and restores filters', async ({
  page,
}) => {
  const apiRequests: string[] = []
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  page.on('request', (request) => {
    if (/^\/(v1|api)\//.test(new URL(request.url()).pathname))
      apiRequests.push(request.url())
  })
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  const panel = page.getByRole('complementary', { name: 'Deployment details' })
  const fullPage = page.getByRole('main', { name: 'Deployment full page' })
  const search = page.getByRole('textbox', { name: 'Search deployments' })
  await expect(page.getByRole('article')).toHaveCount(7)
  const running = page.getByRole('article', { name: 'Deploy api + worker' })
  await expect(
    running.getByRole('group', { name: 'Current step', exact: true })
  ).toHaveText('Current stepRun api post-deploy readiness check')
  await expect(
    running.getByRole('group', { name: 'Next step', exact: true })
  ).toHaveText('Up nextDeploy worker')
  await expect(running).toContainText('3/6 steps complete')
  for (const name of ['Run api post-deploy readiness check', 'Deploy worker']) {
    const fontSize = await running
      .getByText(name, { exact: true })
      .evaluate((element) => parseFloat(getComputedStyle(element).fontSize))
    expect(fontSize).toBeGreaterThanOrEqual(16)
  }
  const nextBox = await running
    .getByRole('group', { name: 'Next step', exact: true })
    .boundingBox()
  const currentBox = await running
    .getByRole('group', { name: 'Current step', exact: true })
    .boundingBox()
  const identityBox = (await running
    .locator('.density-identity')
    .boundingBox())!
  if (identityBox.width >= 384) {
    expect(nextBox!.x).toBeGreaterThan(currentBox!.x + currentBox!.width)
    expect(nextBox!.y).toBeCloseTo(currentBox!.y, 0)
  } else {
    expect(nextBox!.x).toBeCloseTo(currentBox!.x, 0)
    expect(nextBox!.y).toBeGreaterThanOrEqual(
      currentBox!.y + currentBox!.height
    )
  }
  expect((await running.boundingBox())!.width).toBeGreaterThan(768)
  const currentStyle = await running
    .getByText('Run api post-deploy readiness check', { exact: true })
    .evaluate((element) => ({
      weight: Number(getComputedStyle(element).fontWeight),
      color: getComputedStyle(element).color,
    }))
  const nextStyle = await running
    .getByText('Deploy worker', { exact: true })
    .evaluate((element) => ({
      weight: Number(getComputedStyle(element).fontWeight),
      color: getComputedStyle(element).color,
    }))
  expect(currentStyle.weight).toBeGreaterThan(nextStyle.weight)
  expect(currentStyle.color).not.toBe(nextStyle.color)
  const footerBox = await running.locator('footer').boundingBox()
  expect(footerBox!.y).toBeGreaterThanOrEqual(nextBox!.y + nextBox!.height)
  await expect(running.locator('footer')).toContainText('3/6 steps complete')
  await expect(running).not.toContainText(
    /Sandbox|Stack|components complete|Actions/
  )
  await expect(running.getByRole('link')).toHaveCount(0)
  await search.fill('api + worker')
  await page.getByRole('button', { name: 'Status', exact: true }).click()
  await page.locator('input[name="deployments-filter-status-running"]').check()
  await page.getByRole('button', { name: 'Status (1)', exact: true }).click()
  await running.getByRole('button', { name: 'View details' }).click()
  await expect(panel).toBeVisible()
  await expect(fullPage).toHaveCount(0)
  await expect(
    panel.getByRole('heading', { name: 'Workflow steps' })
  ).toBeVisible()
  await expect(panel.locator('.tab-group button')).toHaveText([
    'Template updates',
    'Workflow',
    'Change summary',
  ])
  await panel
    .getByRole('button', { name: 'Template updates', exact: true })
    .click()
  await expect(panel.getByRole('code')).toContainText('replicaCount: 3')
  await expect(
    panel.getByRole('button', { name: 'Sandbox modified' })
  ).toHaveCount(0)
  await panel.getByRole('button', { name: 'worker docker build added' }).click()
  await expect(panel.getByRole('code')).toContainText('/usr/local/bin/worker')
  const openFullPage = panel.getByRole('link', { name: 'Open full page' })
  await expect(openFullPage).toHaveAttribute(
    'href',
    /wf-preview-running\/template-updates$/
  )
  await openFullPage.click()
  await expect(panel).toHaveCount(0)
  await expect(fullPage).toBeVisible()
  await expect(
    fullPage.getByRole('heading', { name: 'Deploy api + worker', exact: true })
  ).toBeVisible()
  await expect(fullPage.getByRole('code')).toContainText('replicaCount: 3')
  await fullPage.getByRole('link', { name: 'Workflow', exact: true }).click()
  await expect(
    fullPage.getByRole('heading', { name: 'Workflow steps' })
  ).toBeVisible()
  await fullPage
    .getByRole('link', { name: 'Change summary', exact: true })
    .click()
  await expect(fullPage.getByText(/these are not final results/)).toBeVisible()
  await page.getByRole('link', { name: 'Back to deployments' }).click()
  await expect(search).toHaveValue('api + worker')
  await page.getByRole('button', { name: 'Status (1)', exact: true }).click()
  await expect(
    page.locator('input[name="deployments-filter-status-running"]')
  ).toBeChecked()
  await page.getByRole('button', { name: 'Status (1)', exact: true }).click()
  await expect(page.getByRole('article')).toHaveCount(1)
  await running.getByRole('button', { name: 'View details' }).click()
  await page.keyboard.press('Escape')
  await expect(panel).toHaveCount(0)
  await expect(search).toHaveValue('api + worker')
  expect(apiRequests).toEqual([])
  expect(errors).toEqual([])
})

test('completed and failed details stay distinct, and direct full-page links bypass the panel', async ({
  page,
}) => {
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  const panel = page.getByRole('complementary', { name: 'Deployment details' })
  const fullPage = page.getByRole('main', { name: 'Deployment full page' })
  const stackRuns = page.getByRole('article', {
    name: 'Update stack permissions',
    exact: true,
  })
  await expect(stackRuns.first()).toContainText('Previous stack update failed')
  await expect(
    stackRuns.last().getByRole('group', { name: 'Stack: Failed', exact: true })
  ).toBeVisible()
  await expect(stackRuns.last()).toContainText('invalid resource ARN')
  await expect(stackRuns.last()).not.toContainText('steps complete')
  await expect(
    stackRuns.getByRole('group', { name: 'Next step', exact: true })
  ).toHaveCount(2)
  await stackRuns.last().getByRole('button', { name: 'View details' }).click()
  await expect(
    panel.getByRole('heading', { name: 'Workflow steps' })
  ).toBeVisible()
  await expect(
    panel.getByText('Validate stack policy', { exact: true })
  ).toBeVisible()
  await panel
    .getByRole('button', { name: 'Change summary', exact: true })
    .click()
  await expect(
    panel.getByText('No changes applied', { exact: true })
  ).toBeVisible()
  await panel.getByRole('button', { name: 'Close panel' }).click()
  await stackRuns.first().getByRole('button', { name: 'View details' }).click()
  await expect(panel.locator('.tab-group button')).toHaveText([
    'Change summary',
    'Workflow',
    'Template updates',
  ])
  await expect(
    panel.getByRole('heading', { name: 'Change summary', exact: true })
  ).toBeVisible()
  await panel
    .getByRole('button', { name: 'Template updates', exact: true })
    .click()
  await expect(panel.getByRole('code')).toContainText(
    'arn:aws:s3:::acme-artifacts/*'
  )
  await expect(
    panel.getByRole('button', { name: 'api helm chart modified' })
  ).toHaveCount(0)
  await panel.getByRole('button', { name: 'Close panel' }).click()
  await page
    .getByRole('link', { name: 'Preview an existing full-page link' })
    .click()
  await expect(panel).toHaveCount(0)
  await expect(
    fullPage.getByRole('heading', {
      name: 'Roll out template v12',
      exact: true,
    })
  ).toBeVisible()
  await expect(fullPage.getByRole('navigation').getByRole('link')).toHaveText([
    'Change summary',
    'Workflow',
    'Template updates',
  ])
  await fullPage
    .getByRole('link', { name: 'Template updates', exact: true })
    .click()
  await fullPage.getByRole('button', { name: 'Sandbox modified' }).click()
  await expect(fullPage.getByRole('code')).toContainText('t3.large')
  await fullPage.getByRole('link', { name: 'Workflow', exact: true }).click()
  await expect(fullPage.getByText(/^Completed in/)).toHaveCount(6)
  await page.getByRole('link', { name: 'Back to deployments' }).click()
  await expect(page.getByRole('article')).toHaveCount(7)
  await page.getByRole('button', { name: 'Status', exact: true }).click()
  await page.locator('input[name="deployments-filter-status-failed"]').check()
  await page.getByRole('button', { name: 'Status (1)', exact: true }).click()
  await expect(page.getByRole('article')).toHaveCount(2)
  await page
    .getByRole('textbox', { name: 'Search deployments' })
    .fill('does not exist')
  await expect(
    page.getByText('No deployments found', { exact: true })
  ).toBeVisible()
})

test('original type, resource and date filters combine and clear', async ({
  page,
}) => {
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  for (const name of ['Status', 'Type', 'Resource', 'Date']) {
    await expect(page.getByRole('button', { name, exact: true })).toBeVisible()
  }
  await page.getByRole('button', { name: 'Type', exact: true }).click()
  await page
    .locator('input[name="deployments-filter-type-stack_update"]')
    .check()
  await expect(page.getByRole('article')).toHaveCount(2)
  await page
    .locator('input[name="deployments-filter-type-component_deploy"]')
    .check()
  await page.getByRole('button', { name: 'Type (2)', exact: true }).click()
  await expect(page.getByRole('article')).toHaveCount(3)
  await page.getByRole('button', { name: 'Resource', exact: true }).click()
  await page.getByRole('radio', { name: 'worker', exact: true }).check()
  await page.getByRole('button', { name: 'worker', exact: true }).click()
  await expect(page.getByRole('article')).toHaveCount(1)
  await expect(page.getByRole('article')).toHaveAttribute(
    'aria-label',
    'Deploy api + worker'
  )
  await page.getByRole('button', { name: 'Clear filters', exact: true }).click()
  await expect(page.getByRole('article')).toHaveCount(7)
  await page.getByRole('button', { name: 'Resource', exact: true }).click()
  await page.getByRole('radio', { name: 'sandbox', exact: true }).check()
  await page.getByRole('button', { name: 'sandbox', exact: true }).click()
  await expect(page.getByRole('article')).toHaveCount(4)
  await page.getByRole('button', { name: 'Date', exact: true }).click()
  await page.getByRole('radio', { name: 'Last 24 hours', exact: true }).check()
  await expect(page.getByRole('article')).toHaveCount(3)
  await expect(page.getByRole('article').first()).toHaveAttribute(
    'aria-label',
    'Roll out template v14'
  )
  await expect(page.getByRole('article').last()).toHaveAttribute(
    'aria-label',
    'Roll out template v13'
  )
  await page.getByRole('radio', { name: 'Last 7 days', exact: true }).check()
  await page.getByRole('button', { name: 'Last 7 days', exact: true }).click()
  await expect(page.getByRole('article')).toHaveCount(4)
  await expect(page.getByRole('article').last()).toHaveAttribute(
    'aria-label',
    'Roll out template v12'
  )
  await page.getByRole('button', { name: 'Clear filters', exact: true }).click()
  await expect(page.getByRole('article')).toHaveCount(7)
  await expect(
    page.getByRole('button', { name: 'Clear filters', exact: true })
  ).toHaveCount(0)
})

test('completed infrastructure with unstarted components remains in progress', async ({
  page,
}) => {
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  const run = page.getByRole('article', {
    name: 'Roll out template v14',
    exact: true,
  })
  for (const name of [
    'Stack: Completed',
    'Sandbox: Completed',
    'Components: Not started',
  ]) {
    await expect(run.getByRole('group', { name, exact: true })).toBeVisible()
  }
  await expect(run).toContainText('Running now')
  await expect(run).not.toContainText(/Partial rollout|Failed|Error/)
  await expect(
    run.getByRole('group', { name: 'Current step', exact: true })
  ).toContainText('Prepare component deployment')
  await expect(
    run.getByRole('group', { name: 'Next step', exact: true })
  ).toContainText('Deploy api')
  await expect(run).toContainText('4/9 steps complete')
  await run.getByRole('button', { name: 'View details' }).click()
  const panel = page.getByRole('complementary', { name: 'Deployment details' })
  const outcomes = panel.getByRole('region', {
    name: 'Resource rollout outcomes',
  })
  for (const name of [
    'Stack: Update completed',
    'Sandbox: Update completed',
    'api: Not started',
    'worker: Not started',
    'legacy: Removal not started',
  ]) {
    await expect(
      outcomes.getByRole('listitem', { name, exact: true })
    ).toBeVisible()
  }
  await expect(
    panel.getByRole('heading', { name: 'Workflow steps' })
  ).toBeVisible()
  await panel.getByRole('link', { name: 'Open full page' }).click()
  const fullPage = page.getByRole('main', { name: 'Deployment full page' })
  await expect(
    fullPage
      .getByRole('region', { name: 'Resource rollout outcomes' })
      .getByRole('listitem', { name: 'api: Not started', exact: true })
  ).toBeVisible()
})

test('approval wait is distinct from execution and stays fixture-only', async ({
  page,
}) => {
  const apiRequests: string[] = []
  page.on('request', (request) => {
    if (/^\/(v1|api)\//.test(new URL(request.url()).pathname))
      apiRequests.push(request.url())
  })
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  const run = page.getByRole('article', {
    name: 'Roll out template v15',
    exact: true,
  })
  await expect(run.locator('.density-run-meta')).not.toContainText(
    'In progress'
  )
  await expect(run).toContainText('Pending approval')
  await expect(run).not.toContainText(
    /Running now|After approval|Partial rollout/
  )
  await expect(
    run.getByRole('group', { name: 'Current step', exact: true })
  ).toHaveText('Current stepComponent plan approval')
  await expect(
    run.getByRole('group', { name: 'Next step', exact: true })
  ).toHaveText('Up nextDeploy api')
  for (const name of [
    'Stack: Completed',
    'Sandbox: Completed',
    'Components: Not started',
  ]) {
    await expect(run.getByRole('group', { name, exact: true })).toBeVisible()
  }
  await expect(run.locator('footer .bg-orange-500')).toHaveCount(0)
  await expect(run.locator('footer .bg-blue-500')).toHaveCount(0)
  await expect(run.locator('footer .bg-green-500')).toHaveCount(4)
  await expect(run.locator('footer .bg-cool-grey-200')).toHaveCount(5)
  await expect(run).toContainText('4/9 steps complete')
  await page.getByRole('button', { name: 'Status', exact: true }).click()
  await page.locator('input[name="deployments-filter-status-running"]').check()
  await page.getByRole('button', { name: 'Status (1)', exact: true }).click()
  await expect(run).toBeVisible()
  await run.getByRole('button', { name: 'View details' }).click()
  const panel = page.getByRole('complementary', { name: 'Deployment details' })
  await expect(
    panel.getByText('Helm chart plan requires review', { exact: true })
  ).toBeVisible()
  for (const label of ['Retry plan', 'Deny plan', 'Approve plan']) {
    await panel.getByRole('button', { name: label, exact: true }).click()
    const dialog = page.getByRole('dialog')
    await expect(dialog).toContainText('No deployment will be changed.')
    await dialog.getByRole('button', { name: label, exact: true }).click()
    await expect(dialog).toHaveCount(0)
    await expect(panel.getByRole('status')).toHaveText(
      `${label} simulated. The fixture is unchanged.`
    )
  }
  await expect(
    panel.getByRole('listitem', {
      name: 'api: Not started — awaiting plan approval',
      exact: true,
    })
  ).toBeVisible()
  await panel
    .getByRole('button', { name: 'Change summary', exact: true })
    .click()
  await expect(
    panel.getByText(/Component changes are planned and awaiting approval/)
  ).toBeVisible()
  const openFullPage = panel.locator('a[href$="/wf-preview-approval"]')
  await openFullPage.click()
  const fullPage = page.getByRole('main', { name: 'Deployment full page' })
  await expect(
    fullPage.locator('header').getByText('Pending approval', { exact: true })
  ).toBeVisible()
  await expect(
    fullPage.locator('header').getByText('In progress', { exact: true })
  ).toBeVisible()
  await expect(
    fullPage.getByText(/Component changes are planned and awaiting approval/)
  ).toBeVisible()
  expect(apiRequests).toEqual([])
})

for (const width of [1280, 390]) {
  test(`resource outcomes distinguish partial failure from unapplied changes at ${width}px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 1000 })
    await page.goto(STORY, { waitUntil: 'domcontentloaded' })
    const completed = page.getByRole('article', {
      name: 'Roll out template v12',
      exact: true,
    })
    for (const category of ['Stack', 'Sandbox', 'Components']) {
      await expect(
        completed.getByRole('group', {
          name: `${category}: Completed`,
          exact: true,
        })
      ).toBeVisible()
    }
    await expect(completed).not.toContainText(
      /updates completed|steps complete/
    )
    const stack = page
      .getByRole('article', { name: 'Update stack permissions', exact: true })
      .first()
    await expect(
      stack.getByRole('group', { name: 'Stack: Completed', exact: true })
    ).toBeVisible()
    await expect(stack).not.toContainText(/Sandbox|Components/)

    const partial = page.getByRole('article', {
      name: 'Roll out template v13',
      exact: true,
    })
    for (const category of ['Stack', 'Sandbox']) {
      await expect(
        partial.getByRole('group', {
          name: `${category}: Completed`,
          exact: true,
        })
      ).toBeVisible()
    }
    await expect(
      partial.getByRole('group', {
        name: 'Components: Partial rollout',
        exact: true,
      })
    ).toBeVisible()
    await expect(partial).toContainText(
      'worker: post-deploy readiness check failed'
    )
    await expect(partial).not.toContainText(/steps complete|scheduler/)
    await partial.getByRole('button', { name: 'View details' }).click()
    const panel = page.getByRole('complementary', {
      name: 'Deployment details',
    })
    const outcomes = panel.getByRole('region', {
      name: 'Resource rollout outcomes',
    })
    await expect(outcomes.getByRole('listitem')).toHaveCount(5)
    for (const name of [
      'api: Rollout completed',
      'worker: Deployed; readiness check failed',
      'scheduler: Not started',
    ]) {
      await expect(
        outcomes.getByRole('listitem', { name, exact: true })
      ).toBeVisible()
    }
    await panel
      .getByRole('button', { name: 'Change summary', exact: true })
      .click()
    await expect(panel.getByText(/Some changes were applied/)).toBeVisible()
    await expect(
      panel.getByText('No changes applied', { exact: true })
    ).toHaveCount(0)
    const fullPageLink = panel.getByRole('link', { name: 'Open full page' })
    await expect(fullPageLink).toHaveCount(1)
    await fullPageLink.click()
    const fullPage = page.getByRole('main', { name: 'Deployment full page' })
    await expect(
      fullPage
        .getByRole('region', { name: 'Resource rollout outcomes' })
        .getByRole('listitem', { name: 'scheduler: Not started' })
    ).toBeVisible()
    await expect(fullPage.getByText(/Some changes were applied/)).toBeVisible()
    expect(
      await page.evaluate(() => document.documentElement.scrollWidth)
    ).toBe(width)
  })
}

test('all detail tabs fit on mobile in both panel and full-page views', async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })
  const running = page.getByRole('article', {
    name: 'Deploy api + worker',
    exact: true,
  })
  const current = running.getByRole('group', {
    name: 'Current step',
    exact: true,
  })
  const next = running.getByRole('group', { name: 'Next step', exact: true })
  await expect(next).toBeVisible()
  const currentBox = await current.boundingBox()
  const nextBox = await next.boundingBox()
  expect(nextBox!.y).toBeGreaterThanOrEqual(currentBox!.y + currentBox!.height)
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(
    390
  )
  await page.getByRole('button', { name: 'View details' }).first().click()
  const panel = page.getByRole('complementary', { name: 'Deployment details' })
  const lastTab = panel.getByRole('button', {
    name: 'Change summary',
    exact: true,
  })
  await expect(lastTab).toBeVisible()
  await expect
    .poll(async () => {
      const box = await lastTab.boundingBox()
      return box ? box.x + box.width : Infinity
    })
    .toBeLessThanOrEqual(390)
  await lastTab.click()
  const openFullPage = panel.locator('a[href$="/wf-preview-running"]')
  await expect(openFullPage).toHaveAccessibleName('Open full page')
  await openFullPage.click()
  const fullPage = page.getByRole('main', { name: 'Deployment full page' })
  const pageTab = fullPage.getByRole('link', {
    name: 'Change summary',
    exact: true,
  })
  await expect(pageTab).toBeVisible()
  const pageTabBox = await pageTab.boundingBox()
  expect(pageTabBox!.x + pageTabBox!.width).toBeLessThanOrEqual(390)
  await fullPage.getByRole('link', { name: 'Workflow', exact: true }).click()
  await expect(
    fullPage.getByRole('heading', { name: 'Workflow steps' })
  ).toBeVisible()
})

for (const width of [390, 1440]) {
  test(`routed deployment stories use the production feed and details at ${width}px`, async ({
    page,
  }) => {
    await page.setViewportSize({ width, height: 1100 })
    const errors: string[] = []
    page.on('pageerror', (error) => errors.push(error.message))

    for (const state of ['in-progress', 'succeeded', 'failed']) {
      await page.goto(
        `/?story=views--installs--deployment-details--${state}&mode=preview`,
        { waitUntil: 'domcontentloaded' }
      )
      const tabs = page
        .getByRole('navigation', { name: 'tab navigation', exact: true })
        .filter({
          has: page.getByRole('link', { name: 'Change summary', exact: true }),
        })
      await expect(tabs).toBeVisible()
      if (width < 600 && state === 'in-progress') {
        await page
          .locator('aside')
          .getByRole('button', { name: 'Close sidebar', exact: true })
          .click()
      }
      for (const name of ['Template updates', 'Workflow', 'Change summary']) {
        const tab = tabs.getByRole('link', { name, exact: true })
        await expect(tab).toBeVisible()
        const box = await tab.boundingBox()
        expect(box!.x).toBeGreaterThanOrEqual(0)
        expect(box!.x + box!.width).toBeLessThanOrEqual(width)
      }
      await tabs.getByRole('link', { name: 'Workflow', exact: true }).click()
      await expect(page.getByPlaceholder('Search workflow steps')).toBeVisible()
      await page
        .getByRole('link', { name: 'Back to deployments', exact: true })
        .click()
      const row = page.getByRole('article', {
        name: 'Update production install',
        exact: true,
      })
      await expect(row).toBeVisible()
      const section = page.getByRole('region', {
        name: state === 'in-progress' ? 'In progress' : 'History',
        exact: true,
      })
      await expect(section.getByRole('article')).toHaveCount(1)
      for (const name of ['Current step', 'Next step']) {
        const context = row.getByRole('group', { name, exact: true })
        if (state === 'in-progress') await expect(context).toBeVisible()
        else await expect(context).toHaveCount(0)
      }
      for (const name of ['Status', 'Type', 'Resource', 'Date']) {
        await expect(
          page.getByRole('button', { name, exact: true })
        ).toBeVisible()
      }
      await row
        .getByRole('button', { name: 'View details', exact: true })
        .click()
      const panel = page.getByRole('complementary', {
        name: 'Deployment details',
      })
      await expect(panel).toHaveCSS('opacity', '1')
      await expect(
        panel.getByRole('link', { name: 'main', exact: true })
      ).toBeVisible()
      await expect(
        panel.getByLabel('Commit abc123def456', { exact: true })
      ).toBeVisible()
      await panel.getByRole('button', { name: 'Workflow', exact: true }).click()
      await expect(
        panel.getByPlaceholder('Search workflow steps')
      ).toBeVisible()
      await page.keyboard.press('Escape')
      await expect(panel).not.toBeVisible()
    }
    expect(errors).toEqual([])
  })
}
