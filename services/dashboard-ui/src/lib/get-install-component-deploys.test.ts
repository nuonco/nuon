import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getInstallComponentDeploys } from './get-install-component-deploys'

const componentId = 'component-id'
const installId = 'install-id'
const orgId = 'org-id'

test('getInstallComponentDeploys should return an array of deploy object', async () => {
  const spec = await getInstallComponentDeploys({
    installId,
    componentId,
    orgId,
  })

  expect(spec).toHaveLength(3)
  spec.map((s) => {
    expect(s).toHaveProperty('id')
    expect(s).toHaveProperty('created_at')
    expect(s).toHaveProperty('updated_at')
    expect(s).toHaveProperty('component_name')
  })
})

test('getInstallComponentDeploys should throw an error when it can not find any deploy for the install component', async () => {
  try {
    await getInstallComponentDeploys({
      installId,
      componentId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: Failed to fetch install component deploys]`
    )
  }
})
