import '@test/mock-fetch-options'
import { expect, test } from 'vitest'
import { getComponentBuilds } from './get-component-builds'

const componentId = 'component-id'
const orgId = 'org-id'

test.skip('getComponentBuilds should return an array of build object', async () => {
  const spec = await getComponentBuilds({
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

test.skip('getComponentBuilds should throw an error when it can not find any builds for the component', async () => {
  try {
    await getComponentBuilds({
      componentId,
      orgId,
    })
  } catch (error) {
    expect(error).toMatchInlineSnapshot(
      `[Error: Failed to fetch component builds]`
    )
  }
})
