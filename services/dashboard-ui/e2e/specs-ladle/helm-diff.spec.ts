import { expect, test, type Locator, type Page } from '@playwright/test'

const STORY = '/?story=diffs--helmdiff--cert-manager-install&mode=preview'

const renderedCode = async (viewer: Locator) =>
  (await viewer.locator('code').allTextContents()).join('\n')

const escapeRegExp = (value: string) =>
  value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

const resource = (page: Page, description: string) => {
  const trigger = page.getByRole('button', {
    name: new RegExp(`^cert-manager ${escapeRegExp(description)}(?:\\s|$)`),
  })

  return {
    trigger,
    viewer: trigger.locator('xpath=../..').locator('diffs-container'),
  }
}

test('resources sharing a filename render their own diff content', async ({
  page,
}) => {
  await page.goto(STORY, { waitUntil: 'domcontentloaded' })

  const serviceAccount = resource(page, 'ServiceAccount · v1 · cert-manager')
  await serviceAccount.trigger.click()
  await expect
    .poll(() => renderedCode(serviceAccount.viewer), { timeout: 10_000 })
    .toContain('kind: ServiceAccount')

  const deployment = resource(page, 'Deployment · apps/v1 · cert-manager')
  await deployment.trigger.click()
  await expect
    .poll(() => renderedCode(deployment.viewer), { timeout: 10_000 })
    .toContain('kind: Deployment')
  await expect(
    page.getByText(
      'DiffHunksRenderer.processDiffResult: deletionLine and additionLine are null, something is wrong'
    )
  ).toHaveCount(0)

  await page.getByRole('button', { name: 'Expand all' }).click()

  const resources = [
    ['Deployment · apps/v1 · cert-manager', 'kind: Deployment'],
    ['ServiceAccount · v1 · cert-manager', 'kind: ServiceAccount'],
    [
      'ClusterRole · rbac.authorization.k8s.io/v1 · cert-manager',
      'kind: ClusterRole',
    ],
    [
      'ClusterRoleBinding · rbac.authorization.k8s.io/v1 · cert-manager',
      'kind: ClusterRoleBinding',
    ],
  ] as const

  await expect(
    page.getByText('cert-manager.yaml', { exact: true })
  ).toHaveCount(resources.length)

  for (const [description, manifestKind] of resources) {
    const { trigger, viewer } = resource(page, description)

    await expect(trigger).toHaveAttribute('aria-expanded', 'true')
    await expect(viewer).toBeVisible()
    await expect
      .poll(() => renderedCode(viewer), { timeout: 10_000 })
      .toContain(manifestKind)
  }
})
